package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/orderer"
	cp "github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/peer"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/user"
	"github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/peer/common"
	protos_common "github.com/hyperledger/fabric/protos/common"
	protos_peer "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	putils "github.com/hyperledger/fabric/protos/utils"
	"github.com/spf13/viper"
)

func InitSDK(configPath, configFile string) error {
	viper.SetEnvPrefix("core")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	configFilePath := filepath.Join(configPath, configFile+".yaml")
	viper.SetConfigFile(configFilePath)
	//err := common.InitConfig(configFile)
	//if err == nil {
	err := viper.ReadInConfig()
	//}
	if err != nil { // Handle errors reading the config file
		return fmt.Errorf("Fatal error when initializing %s config : %s\n", "SDK", err)
	}

	// Init the MSP
	var mspID = viper.GetString("peer.localMspId")
	var mspMgrConfigDir = viper.GetString("peer.mspConfigPath")
	err = common.InitCrypto(mspMgrConfigDir, mspID,"")
	if err != nil { // Handle errors reading the config file
		return fmt.Errorf("InitSDK failed:%s", err.Error())
	}

	if err = cp.InitPeerClient(); err != nil {
		return err
	} else if err = orderer.InitBroadcastClient(); err != nil {
		return err
	}

	// set the logging level for specific modules defined via environment
	// variables or core.yaml
	overrideLogModules := []string{"msp"}
	for _, module := range overrideLogModules {
		err = common.SetLogLevelFromViper(module)
		if err != nil {
			return fmt.Errorf("Error setting log level for module '%s': %s", module, err.Error())
		}
	}
	return nil
}

func Close() {
	for _, o := range orderer.GetOrdererClients() {
		o.Close()
	}

	for _, p := range cp.GetPeerClients("", false) {
		p.Close()
	}
}

func CreateProposal(ccid, chainID string, nonce []byte, args []string, headerType protos_common.HeaderType, carrier map[string]string) (*protos_peer.Proposal, string, error) {
	// Build the spec
	input := &protos_peer.ChaincodeInput{Args: util.ToChaincodeArgs(args...)}

	spec := &protos_peer.ChaincodeSpec{
		Type:        protos_peer.ChaincodeSpec_Type(protos_peer.ChaincodeSpec_Type_value["GOLANG"]),
		ChaincodeId: &protos_peer.ChaincodeID{Name: ccid},
		Input:       input,
	}

	// Build the ChaincodeInvocationSpec message
	invocation := &protos_peer.ChaincodeInvocationSpec{ChaincodeSpec: spec}

	signer, err := user.GetSigner()
	if err != nil {
		return nil, "", err
	}

	creator, err := signer.Serialize()
	if err != nil {
		return nil, "", fmt.Errorf("Error serializing identity for %s: %s", signer.GetIdentifier(), err)
	}

	if nonce == nil {
		// generate a random nonce
		nonce, err = crypto.GetRandomNonce()
		if err != nil {
			return nil, "", err
		}
	}
	// compute txid
	txId, err := putils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return nil, "", err
	}

	var transientMap map[string][]byte
	if carrier != nil {
		transientMap = make(map[string][]byte)
		for k, v := range carrier {
			transientMap[k] = []byte(v)
		}
	}

	return putils.CreateChaincodeProposalWithTxIDNonceAndTransient(txId, headerType, chainID, invocation, nonce, creator, transientMap)
}

func ProcessProposal(peerClients []*cp.PeerClient, signedProp *protos_peer.SignedProposal) ([]*protos_peer.ProposalResponse, error) {
	var proposalResps []*protos_peer.ProposalResponse
	var err error

	for _, client := range peerClients {
		proposalResp, err := client.ProcessProposal(signedProp)
		if err != nil {
			return proposalResps, err
		}
		proposalResps = append(proposalResps, proposalResp)
	}

	if len(proposalResps) == 0 {
		err = fmt.Errorf("The Client is empty,can't invoke function ProcessProposal.")
	}
	return proposalResps, err
}

func CreateAndProcessProposal(peerClients []*cp.PeerClient, ccId, chainID string, args []string, headerType protos_common.HeaderType) ([]*protos_peer.ProposalResponse, error) {
	prop, _, err := CreateProposal(ccId, chainID, nil, args, headerType, nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot create proposal, due to %s", err)
	}

	signedProp, err := user.GetSignedProposal(prop)
	if err != nil {
		return nil, err
	}

	resps, err := ProcessProposal(peerClients, signedProp)
	if err != nil {
		return nil, fmt.Errorf("Failed sending proposal, got %s", err)
	}
	return resps, nil
}

func Broadcast(ordererClients []*orderer.OrdererClient, proposal *protos_peer.Proposal, resps []*protos_peer.ProposalResponse) error {
	signer, err := user.GetSigner()
	if err != nil {
		return fmt.Errorf("Broadcast failed:%s", err.Error())
	}

	// assemble a signed transaction (it's an Envelope message)
	env, err := utils.CreateSignedTx(proposal, signer, resps...)
	if err != nil {
		return fmt.Errorf("Could not assemble transaction, err %s", err)
	}

	if env == nil {
		return fmt.Errorf("Could not create signed tx Envelope %v", resps)
	}

	for _, client := range ordererClients {
		err = client.BroadcastClientSend(env)
		if err != nil {
			return err
		}
	}
	return nil
}
