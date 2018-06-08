package handle

import (
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric/core/scc/qscc"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/peersafe/gohfc"
	"io/ioutil"
	"os"
	"strconv"
)

type FilterHandler func(*gohfc.EventBlockResponseTransactionEvent) (interface{}, bool)

const (
	fileSaveName = "current.info"
)

var (
	currentBlockHeight uint64
	info               = new(BlockInfo)
)

func SetBlockInfo(info *BlockInfo) error {
	//make a json
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	//write into file
	f, err := os.Create(fileSaveName)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	} else if err == nil {
		err = f.Close()
	}
	return err
}

func GetBlockInfo() error {
	//read from file
	data, err := ioutil.ReadFile(fileSaveName)
	if err != nil {
		return err
	}

	//parse from json
	err = json.Unmarshal(data, info)
	return err
}

func CheckAndRecoverEvent(chainID string, filterHandler FilterHandler, fromListen chan BlockInfoAll) {
	var currentBlockNum uint64 = 0

	if filterHandler == nil {
		logger.Errorf("The filter handler is null!")
		return
	}

	err := GetBlockInfo()
	if err != nil {
		logger.Error("get block info from current.info file failed:", err.Error())
		logger.Warning("Set the info.blockNum to be 0")
		info.Block_number = 0
	}
	currentBlockNum = info.Block_number
	logger.Info("the block height is", currentBlockHeight, "and has processed", currentBlockNum)

	//Retrieve the transactions, which were written during eventserver is not running
	for ; currentBlockNum < currentBlockHeight; currentBlockNum++ {
		block, err := GetBlockByNumber(chainID, currentBlockNum)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		fmt.Println("----getblockbynumber-----")
		fmt.Println(block.Header.Number)
	}

	//Handle the transactions from listen module
	for {
		select {
		case blockInfo := <-fromListen:
			/*
				err = mq.SendMessage(blockInfo.MsgInfo)
				if err != nil {
					logger.Errorf("Send to message queue failed: %s", err.Error())
					continue
				}
			*/
			err = SetBlockInfo(&BlockInfo{Block_number: blockInfo.Block_number, Tx_index: blockInfo.Tx_index})
			if err != nil {
				logger.Errorf("Set block info failed: %s", err.Error())
				continue
			}
		}
	}
}

func GetBlockByNumber(chainId string, blockNum uint64) (*common.Block, error) {
	strBlockNum := strconv.FormatUint(blockNum, 10)
	args := []string{qscc.GetBlockByNumber, chainId, strBlockNum}
	resps, err := gohfc.GetHandler().QueryByQscc(args, []string{"peer0"})
	if err != nil {
		return nil, fmt.Errorf("can not get installed chaincodes :%s", err.Error())
	} else if len(resps) == 0 {
		return nil, fmt.Errorf("GetBlockByNumber empty response from peer")
	}
	if resps[0].Error != nil {
		return nil, resps[0].Error
	}
	data := resps[0].Response.Response.Payload
	var block = new(common.Block)
	err = proto.Unmarshal(data, block)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber Unmarshal from payload failed: %s", err.Error())
	}

	return block, nil
}

func GetBlockHeight(chainId string) bool {
	args := []string{qscc.GetChainInfo, chainId}
	resps, err := gohfc.GetHandler().QueryByQscc(args, []string{"peer0"})
	if err != nil {
		fmt.Errorf("%v",err)
		return false
	} else if len(resps) == 0 {
		fmt.Errorf("GetChainInfo is empty respons from peer qscc")
		return false
	}

	if resps[0].Error != nil {
		fmt.Println(resps[0].Error)
		return false
	}

	data := resps[0].Response.Response.Payload
	var chaininfo= new(common.BlockchainInfo)
	err = proto.Unmarshal(data, chaininfo)
	if err != nil {
		fmt.Errorf("GetChainInfo unmarshal from payload failed: %s", err.Error())
		return false
	}

	currentBlockHeight = chaininfo.Height
	logger.Info("the current block height is", currentBlockHeight)
	return true
}