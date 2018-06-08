package block_listener

import (
	"errors"
	"fmt"
	"os"

	"github.com/hyperledger/fabric/events/consumer"
	"github.com/hyperledger/fabric/protos/common"
	protos_peer "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
)

const (
	ClientReconnectCount = 3
)

type adapter struct {
	notfy   chan *protos_peer.Event_Block
	chainID string
}

//GetInterestedEvents implements consumer.EventAdapter interface for registering interested events
func (a *adapter) GetInterestedEvents() ([]*protos_peer.Interest, error) {
	return []*protos_peer.Interest{{EventType: protos_peer.EventType_BLOCK, ChainID: a.chainID}}, nil
}

//Recv implements consumer.EventAdapter interface for receiving events
func (a *adapter) Recv(msg *protos_peer.Event) (bool, error) {
	if o, e := msg.Event.(*protos_peer.Event_Block); e {
		a.notfy <- o
		return true, nil
	}
	return false, fmt.Errorf("Receive unknown type event: %v", msg)
}

//Disconnected implements consumer.EventAdapter interface for disconnecting
func (a *adapter) Disconnected(err error) {
	fmt.Print("Disconnected...exiting\n")
	os.Exit(1)
}

func createEventClient(eventAddress string, chainID string) *adapter {
	var obcEHClient *consumer.EventsClient

	done := make(chan *protos_peer.Event_Block)
	adapter := &adapter{notfy: done, chainID: chainID}
	obcEHClient, _ = consumer.NewEventsClient(eventAddress, 5, adapter)
	if err := obcEHClient.Start(); err != nil {
		fmt.Printf("could not start chat %s\n", err)
		obcEHClient.Stop()
		return nil
	}

	return adapter
}

func GetTxPayload(tdata []byte) (*common.Payload, error) {
	if tdata == nil {
		return nil, errors.New("Cannot extract payload from nil transaction")
	}

	if env, err := utils.GetEnvelopeFromBlock(tdata); err != nil {
		return nil, fmt.Errorf("Error getting tx from block(%s)", err)
	} else if env != nil {
		// get the payload from the envelope
		payload, err := utils.GetPayload(env)
		if err != nil {
			return nil, fmt.Errorf("Could not extract payload from envelope, err %s", err)
		}
		return payload, nil
	}
	return nil, nil
}

// getChainCodeEvents parses block events for chaincode events associated with individual transactions
func GetChainCodeEvents(payload *common.Payload) (*protos_peer.ChaincodeEvent, error) {
	chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return nil, fmt.Errorf("Could not extract channel header from envelope, err %s", err)
	}

	if common.HeaderType(chdr.Type) == common.HeaderType_ENDORSER_TRANSACTION {
		tx, err := utils.GetTransaction(payload.Data)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling transaction payload for block event: %s", err)
		}
		chaincodeActionPayload, err := utils.GetChaincodeActionPayload(tx.Actions[0].Payload)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling transaction action payload for block event: %s", err)
		}
		propRespPayload, err := utils.GetProposalResponsePayload(chaincodeActionPayload.Action.ProposalResponsePayload)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling proposal response payload for block event: %s", err)
		}
		caPayload, err := utils.GetChaincodeAction(propRespPayload.Extension)
		if err != nil {
			return nil, fmt.Errorf("Error unmarshalling chaincode action for block event: %s", err)
		}
		ccEvent, err := utils.GetChaincodeEvents(caPayload.Events)

		if ccEvent != nil {
			return ccEvent, nil
		}
	}
	return nil, errors.New("No events found")
}

//used for the sdk
func GetListenChannel(eventAddress, chainID string) chan *protos_peer.Event_Block {
	var a *adapter
	for i := 0; i < ClientReconnectCount && a == nil; i++ {
		a = createEventClient(eventAddress, chainID)
	}

	if a == nil {
		fmt.Println("Error creating event client")
		return nil
	}
	return a.notfy
}
