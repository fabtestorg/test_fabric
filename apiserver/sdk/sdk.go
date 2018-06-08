package sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	listener "github.com/hyperledger/fabric-sdk-go-peersafe/pkg/block-listener"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/chaincode"
	pkg_couchdb "github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/couchdb"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/user"
	"github.com/hyperledger/fabric/core/ledger/util/couchdb"
	"github.com/hyperledger/fabric/core/scc/qscc"
	"github.com/hyperledger/fabric/protos/common"
	propeer "github.com/hyperledger/fabric/protos/peer"
	proutils "github.com/hyperledger/fabric/protos/utils"
	"github.com/fabtestorg/test_fabric/apiserver/define"
	"github.com/fabtestorg/test_fabric/apiserver/utils"
	"github.com/peersafe/gohfc"
	"github.com/spf13/viper"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"path/filepath"
	"time"
)

// SDKHanlder sdk handler
type SDKHanlder struct {
	handler *chaincode.Handler
}

type BlockData struct {
	FabricBlockData   string
	FabricBlockPre    string
	FabricBlockHash   string
	FabricBlockHeight uint64
	EventMap          map[string]*propeer.ChaincodeEvent
}

var (
	EventNames                   string
	ordersNames, orRulePeerNames []string
	orgMap                       = make(map[string][]string) //[orgname]peerNames
	Handler                      SDKHanlder
)

// Handler sdk handler

func InitSDK(path, name string) error {
	configFilePath := filepath.Join(path, name+".yaml")
	if err := gohfc.InitSDK(configFilePath); err != nil {
		return err
	}
	if err := SetSendPeerOrder(configFilePath); err != nil {
		return err
	}
	return nil
}

// GetSDK get sdk handler
func GetSDK() *SDKHanlder {
	return &Handler
}

// GetTxId invoke cc
func (sdk *SDKHanlder) GetTxId() (string, []byte, error) {
	return user.GenerateTxId()
}

// Invoke invoke cc
func (sdk *SDKHanlder) Invoke(txId string, nonce []byte, trackId, function string, carrier map[string]string, request []byte) (string, error) {
	peerNames := GetSendPeerName()
	orderName := getSendOrderName()
	if len(peerNames) == 0 || orderName == "" {
		return "", fmt.Errorf("config peer order is err")
	}
	result, err := gohfc.GetHandler().Invoke([]string{function, "txid", string(request)}, peerNames, orderName)
	if err != nil {
		return "", err
	}
	return result.TxID, nil
}

func (sdk *SDKHanlder) PeerKeepalive(function string) bool {
	peerNames := GetSendPeerName()
	if len(peerNames) == 0 {
		fmt.Errorf("config peer order is err")
	}
	response, err := gohfc.GetHandler().Query([]string{function, "reduPara", "reduPara"}, []string{peerNames[0]})
	if err != nil || len(response) == 0 {
		fmt.Println(err)
		return false
	} else {
		keepaliveResult := string(response[0].Response.Payload)
		if keepaliveResult == "Reached" {
			return true
		} else {
			return false
		}
	}
}

func removeDataWrapper(wrappedValue []byte) (interface{}, error) {
	//create a generic map for the json
	jsonResult := make(map[string]interface{})

	//unmarshal the selected json into the generic map
	decoder := json.NewDecoder(bytes.NewBuffer(wrappedValue))
	decoder.UseNumber()
	err := decoder.Decode(&jsonResult)

	//place the result json in the data key
	return jsonResult["data"], err
}

func (sdk *SDKHanlder) DSL(trackId, function string, carrier map[string]string, request define.QueryRequest) (define.QueryContents, *define.ResponseStatus, *define.Page, error) {
	responseData := define.QueryResponse{}
	var retValue define.QueryContents
	factorList := []define.Factor{}
	messages := []define.Message{}

	if function == define.DSL_QUERY {
		retValue.Schema = "/schema/factorList.json"
	}
	responseData.Page = request.SplitPage

	retFunc := func(action string, err error) (define.QueryContents, *define.ResponseStatus, *define.Page, error) {
		if err != nil {
			err = fmt.Errorf("%s failed:%s", action, err.Error())
			responseData.ResponseStatus.StatusCode = 1
			responseData.ResponseStatus.StatusMsg = err.Error()
		}
		return retValue, &responseData.ResponseStatus, &responseData.Page, err
	}

	var err error

	if request.SplitPage.CurrentPage == 0 {
		request.SplitPage.CurrentPage = 1
		request.SplitPage.PageSize = 1000
	}
	pageSize := request.SplitPage.PageSize
	if pageSize == 0 {
		pageSize = 100
	} else if pageSize > 1000 {
		//if pageSize more than 1000,then api server will be easy to faild
		return retFunc("PageSize get", fmt.Errorf("PageSize is more than 1000!"))
	}

	qrs, err := func() (*[]couchdb.QueryResult, error) {
		clients, err := pkg_couchdb.GetDBClients()
		if err != nil {
			return nil, fmt.Errorf("GetCouchDBClients failed:%s", err.Error())
		}
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		cli := clients[r.Intn(len(clients))]
		return cli.QueryDocuments(request.DslSyntax)
	}()

	if err != nil {
		return retFunc("QueryDocuments", err)
	}

	// buffer is a JSON array containing QueryRecords
	for _, qr := range *qrs {
		responseData.Page.TotalRecords++
		val, err := removeDataWrapper(qr.Value)
		if err != nil {
			fmt.Println(string(qr.Value))
			return retFunc("Unmarshal value", err)
		}
		//fmt.Println("DSL query val sssssssssssssss %v",val)
		valueJson, err := json.Marshal(val)
		if err != nil {
			return retFunc("json mash", err)
		}
		message := define.Message{}
		err = json.Unmarshal(valueJson, &message)
		if err != nil {
			fmt.Println(err)
			continue
		}
		messages = append(messages, message)
	}

	userId := viper.GetString("user.id")
	err = utils.FormatResponseMessage(userId, &factorList, &messages)
	if err != nil {
		fmt.Println(err)
	}

	retValue.Payload = factorList
	return retFunc("", err)
}

// Query query cc
func (sdk *SDKHanlder) Query(trackId, function string, carrier map[string]string, request string, blockFlag bool) (define.QueryContents, *define.ResponseStatus, *define.Page, error) {
	var payload []string
	factorList := []define.Factor{}
	responseData := define.QueryResponse{}
	messages := []define.Message{}
	responseData.Payload = &payload
	var retValue define.QueryContents
	peerNames := GetSendPeerName()
	if len(peerNames) == 0 {
		responseData.ResponseStatus.StatusCode = 1
		fmt.Errorf("config peer order is err")
	}
	response, err := gohfc.GetHandler().Query([]string{function, request}, []string{peerNames[0]})
	if err != nil || len(response) == 0 {
		responseData.ResponseStatus.StatusCode = 1
		responseData.ResponseStatus.StatusMsg = err.Error()
		fmt.Println(err)
	} else {
		responseData.ResponseStatus.StatusCode = 0
		responseData.ResponseStatus.StatusMsg = "SUCCESS"
		retValue.Payload = factorList
		if function == define.DSL_QUERY {
			retValue.Schema = "/schema/factorList.json"
		}
		fmt.Println(string(response[0].Response.Payload))
		err = json.Unmarshal(response[0].Response.Payload, &responseData)
		if err != nil {
			responseData.ResponseStatus.StatusCode = 1
			responseData.ResponseStatus.StatusMsg = err.Error()
			fmt.Println(err)
		} else {
			if blockFlag {
				retValue.Payload = payload
				return retValue, &responseData.ResponseStatus, &responseData.Page, err
			}
			for _, jsonData := range payload {
				message := define.Message{}
				err = json.Unmarshal([]byte(jsonData), &message)
				if err != nil {
					fmt.Println(err)
					continue
				}
				messages = append(messages, message)
			}
			userId := viper.GetString("user.id")
			err = utils.FormatResponseMessage(userId, &factorList, &messages)
			if err != nil {
				fmt.Println(err)
			}
			retValue.Payload = factorList
		}
	}

	return retValue, &responseData.ResponseStatus, &responseData.Page, err
}

// Query query qscc
func (sdk *SDKHanlder) BlockQuery(txIdList []string) (define.QueryContents, *define.ResponseStatus, *define.Page, error) {
	responseData := define.QueryResponse{}
	var retValue define.QueryContents

	info, err := GetBlockByTxids(txIdList)
	if err != nil {
		responseData.ResponseStatus.StatusCode = 1
		responseData.ResponseStatus.StatusMsg = err.Error()
		fmt.Println(err)
	} else {
		var list []define.BlockchainData

		for i := range info {
			itemInfo := info[i]
			for _, v := range itemInfo.EventMap {
				var tempData define.BlockchainData
				businessData, err := analysisCCEvent(v.Payload)
				if err != nil {
					fmt.Println(err)
					continue
				}

				if len(businessData) == 1 {
					if err := json.Unmarshal([]byte(businessData[0]), &tempData); err != nil {
						fmt.Println(err)
						continue
					}
					tempData.TxId = v.TxId
					tempData.TxHash = calHash([]byte(itemInfo.FabricBlockData))
					tempData.BlockData = itemInfo.FabricBlockData
					tempData.BlockHash = itemInfo.FabricBlockHash
					tempData.BlockHeight = itemInfo.FabricBlockHeight
					list = append(list, tempData)
				}
			}
		}

		responseData.ResponseStatus.StatusCode = 0
		responseData.ResponseStatus.StatusMsg = "SUCCESS"
		retValue.Payload = list
		retValue.Schema = "/schema/blockList.json"
	}
	return retValue, &responseData.ResponseStatus, &responseData.Page, nil
}

// Query query qscc
func (sdk *SDKHanlder) BlockQueryEx(txIdList []string) (define.QueryContents, *define.ResponseStatus, *define.Page, error) {
	responseData := define.QueryResponse{}
	var retValue define.QueryContents

	info, err := GetBlockByTxids(txIdList)
	if err != nil {
		responseData.ResponseStatus.StatusCode = 1
		responseData.ResponseStatus.StatusMsg = err.Error()
		fmt.Println(err)
	} else {
		var block []define.BlockDataObj
		var tempBlock define.BlockDataObj
		var tempFactor define.Factor
		var tempEvents define.Events
		for _, blockInfo := range info {
			itemInfo := blockInfo
			var evnetList []define.Events
			for _, v := range itemInfo.EventMap {
				businessData, err := analysisCCEvent(v.Payload)
				if err != nil {
					fmt.Println(err)
					continue
				}

				if len(businessData) >= 1 {
					if err := json.Unmarshal([]byte(businessData[0]), &tempFactor); err != nil {
						fmt.Println(err)
						continue
					}
					tempEvents.ChaincodeId = v.ChaincodeId
					tempEvents.TxId = v.TxId
					tempEvents.EventName = v.EventName
					tempEvents.Payload = tempFactor
					evnetList = append(evnetList, tempEvents)
				}
			}

			tempBlock.BlockHash = itemInfo.FabricBlockHash
			tempBlock.BlockHeight = itemInfo.FabricBlockHeight
			tempBlock.PreviousHash = itemInfo.FabricBlockPre
			tempBlock.Events = evnetList
			block = append(block, tempBlock)
		}

		responseData.ResponseStatus.StatusCode = 0
		responseData.ResponseStatus.StatusMsg = "SUCCESS"
		retValue.Payload = block
		retValue.Schema = "/schema/blockList.json"
	}

	return retValue, &responseData.ResponseStatus, &responseData.Page, nil
}

func checkAndRemove(keys *[]string, key string) bool {
	for i, val := range *keys {
		if key == val {
			*keys = append((*keys)[:i], (*keys)[i+1:]...)
			return true
		}
	}
	return false
}

func GetBlockByTxids(txids []string) (map[string]*BlockData, error) {
	var getBlockByTxid = func(txid string) (*common.Block, error) {
		chainId := viper.GetString("chaincode.id.chainID")
		args := []string{qscc.GetBlockByTxID, chainId, txid}
		peerNames := GetSendPeerName()
		if len(peerNames) == 0 {
			return nil, fmt.Errorf("peerName is nil")
		}
		resps, err := gohfc.GetHandler().QueryByQscc(args, []string{peerNames[0]})
		if err != nil  {
			return nil, fmt.Errorf("Can not get installed chaincodes", err.Error())
		} else if len(resps) == 0 {
			return nil, fmt.Errorf("Get empty responce from peer!")
		}
		data := resps[0].Response.Payload
		var block = new(common.Block)
		err = proto.Unmarshal(data, block)
		if err != nil {
			return nil, fmt.Errorf("Unmarshal from payload failed: %s", err.Error())
		}
		return block, nil
	}

	var blockInfos map[string]*BlockData
	blockInfos = make(map[string]*BlockData)

	for _, txid := range txids {
		if len(txid) == 0 {
			return blockInfos, nil
		}
		var block, err = getBlockByTxid(txid)
		if err != nil {
			return blockInfos, err
		}

		blockHash := fmt.Sprintf("%x", block.Header.DataHash)

		if len(blockHash) == 0 {
			continue
		}

		_, exist := blockInfos[blockHash]

		if exist == false {
			blockInfos[blockHash] = new(BlockData)
			blockInfos[blockHash].FabricBlockData = block.String()
			blockInfos[blockHash].FabricBlockPre = fmt.Sprintf("%x", block.Header.PreviousHash)
			blockInfos[blockHash].FabricBlockHash = blockHash
			blockInfos[blockHash].FabricBlockHeight = block.Header.Number
			blockInfos[blockHash].EventMap = make(map[string]*propeer.ChaincodeEvent)
		}

		for _, r := range block.Data.Data {
			tx, _ := listener.GetTxPayload(r)
			if tx != nil {
				chdr, err := proutils.UnmarshalChannelHeader(tx.Header.ChannelHeader)
				if err != nil {
					fmt.Println("Error extracting channel header")
					continue
				}
				events, err := GetChainCodeEvents(tx)
				if err != nil {
					fmt.Println("Received failed from channel '%s':%s", chdr.ChannelId, err.Error())
					continue
				}

				for _, y := range events {
					if checkAndRemove(&txids, y.TxId) {
						blockInfos[blockHash].EventMap[y.TxId] = y
					}
				}
			}
		}
	}
	return blockInfos, nil
}

func analysisCCEvent(payload []byte) ([]string, error) {
	responseData := define.QueryResponse{}
	var messages []string
	responseData.Payload = &messages
	if err := json.Unmarshal(payload, &responseData); err != nil {
		return messages, err
	}
	return messages, nil
}

// getChainCodeEvents parses block events for chaincode events associated with individual transactions
func GetChainCodeEvents(payload *common.Payload) ([]*propeer.ChaincodeEvent, error) {
	chdr, err := proutils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf("Could not extract channel header from envelope, err %s", err)
	}

	if common.HeaderType(chdr.Type) == common.HeaderType_ENDORSER_TRANSACTION {
		tx, err := proutils.GetTransaction(payload.Data)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling transaction payload for block event: %s", err)
		}

		var events []*propeer.ChaincodeEvent
		for _, r := range tx.Actions {
			chaincodeActionPayload, err := proutils.GetChaincodeActionPayload(r.Payload)
			if err != nil {
				return nil, fmt.Errorf("Error unmarshalling transaction action payload for block event: %s", err)
			}
			propRespPayload, err := proutils.GetProposalResponsePayload(chaincodeActionPayload.Action.ProposalResponsePayload)
			if err != nil {
				return nil, fmt.Errorf("Error unmarshalling proposal response payload for block event: %s", err)
			}
			caPayload, err := proutils.GetChaincodeAction(propRespPayload.Extension)
			if err != nil {
				return nil, fmt.Errorf("Error unmarshalling chaincode action for block event: %s", err)
			}
			ccEvent, err := proutils.GetChaincodeEvents(caPayload.Events)
			events = append(events, ccEvent)
		}

		if events != nil {
			return events, nil
		}
	}
	return nil, errors.New("No events found")
}

func calHash(msg []byte) string {
	hash := sha3.New224()
	hash.Write(msg)
	value := hash.Sum(nil)
	return fmt.Sprintf("%x", value)
}

func SetSendPeerOrder(configPath string) error {
	config, err := gohfc.NewClientConfig(configPath)
	if err != nil {
		return err
	}
	policyOrgs := viper.GetStringSlice("policy.orgs")
	policyRule := viper.GetString("policy.rule")
	if len(policyOrgs) == 0 || policyRule == "" {
		return fmt.Errorf("policy config is err")
	}

	for ordname := range config.Orderers {
		ordersNames = append(ordersNames, ordname)
	}
	if len(ordersNames) == 0 {
		return fmt.Errorf("order config is err")
	}

	for k := range config.EventPeers {
		EventNames = k
		break
	}

	for peerName, v := range config.Peers {
		if containsStr(policyOrgs, v.OrgName) {
			orgMap[v.OrgName] = append(orgMap[v.Host], peerName)
		}
	}

	if policyRule == "or" {
		for _, peerNames := range orgMap {
			orRulePeerNames = append(orRulePeerNames, peerNames...)
		}
		if len(orRulePeerNames) == 0 {
			return fmt.Errorf("peer config is err")
		}
	}
	return nil
}
func containsStr(strList []string, str string) bool {
	for _, v := range strList {
		if v == str {
			return true
		}
	}
	return false
}

func getSendOrderName() string {
	return ordersNames[generateRangeNum(0, len(ordersNames))]
}

func generateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

func GetSendPeerName() []string {
	if len(orRulePeerNames) > 0 {
		return []string{orRulePeerNames[generateRangeNum(0, len(orRulePeerNames))]}
	}
	var sendNameList []string
	policyRule := viper.GetString("policy.rule")
	if policyRule == "and" {
		for _, peerNames := range orgMap {
			sendNameList = append(sendNameList, peerNames[generateRangeNum(0, len(peerNames))])
			continue
		}
	}

	return sendNameList
}
