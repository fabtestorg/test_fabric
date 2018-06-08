package user

import (
	"fmt"
	"github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric/peer/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"sync"
)

var signer msp.SigningIdentity
var mutexGet sync.Mutex

func GetSigner() (msp.SigningIdentity, error) {
	if signer == nil {
		mutexGet.Lock()
		defer mutexGet.Unlock()
		if signer == nil {
			var err error
			signer, err = common.GetDefaultSigner()
			if err != nil {
				return nil, fmt.Errorf("Error getting default signer: %s", err)
			}
		}
	}
	return signer, nil
}

func GetSignedProposal(prop *peer.Proposal) (*peer.SignedProposal, error) {
	signer, err := GetSigner()
	if err != nil {
		return nil, fmt.Errorf("ProcessProposal failed:%s", err.Error())
	}

	var signedProp *peer.SignedProposal
	signedProp, err = utils.GetSignedProposal(prop, signer)
	if err != nil {
		return nil, fmt.Errorf("Error creating signed proposal %s", err.Error())
	}
	return signedProp, nil
}

func GenerateTxId() (string, []byte, error) {
	signer, err := GetSigner()
	if err != nil {
		return "", nil, err
	}

	creator, err := signer.Serialize()
	if err != nil {
		return "", nil, fmt.Errorf("Error serializing identity for %s: %s", signer.GetIdentifier(), err)
	}

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return "", nil, err
	}

	// compute txid
	txId, err := utils.ComputeProposalTxID(nonce, creator)
	return txId, nonce, err
}
