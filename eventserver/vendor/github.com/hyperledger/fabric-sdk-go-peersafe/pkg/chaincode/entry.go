package chaincode

import (
	"fmt"
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/block-listener"
	pkg_common "github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/orderer"
	cp "github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/peer"
	"github.com/hyperledger/fabric-sdk-go-peersafe/pkg/common/user"
	util "github.com/hyperledger/fabric/core/ledger/util"
	protcos_common "github.com/hyperledger/fabric/protos/common"
	protos_peer "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

var (
	logger = logging.MustGetLogger("chaincode_sdk")
	once   sync.Once
)

type Handler struct {
	sync.Mutex
	responseChannel map[string]chan *AsyncInvokeResp
	ChannelId       string
	CCName          string
}

func NewHandler(channelName, chaincodeName string) *Handler {
	return &Handler{
		responseChannel: make(map[string]chan *AsyncInvokeResp),
		ChannelId:       channelName,
		CCName:          chaincodeName,
	}
}

func (h *Handler) listenerParse() {
	notfy := block_listener.GetListenChannel(viper.GetString("peer.listenAddress"), h.ChannelId)
	if notfy == nil {
		return
	}
	for {
		select {
		case b := <-notfy:
			txsFltr := util.TxValidationFlags(b.Block.Metadata.Metadata[protcos_common.BlockMetadataIndex_TRANSACTIONS_FILTER])
			for i, r := range b.Block.Data.Data {
				tx, _ := block_listener.GetTxPayload(r)
				if tx != nil {
					chdr, err := utils.UnmarshalChannelHeader(tx.Header.ChannelHeader)
					if err != nil {
						logger.Debug("Error extracting channel header\n")
						return
					}
					if txsFltr.IsInvalid(i) {
						h.sendChannel(chdr.TxId, NewAsyncResp(
							chdr.TxId, fmt.Errorf("Received invalid transaction from channel '%s'", chdr.ChannelId)))
					} else {
						if event, err := block_listener.GetChainCodeEvents(tx); err != nil {
							h.sendChannel(chdr.TxId, NewAsyncResp(
								chdr.TxId, fmt.Errorf("Received failed from channel '%s':%s", chdr.ChannelId, err.Error())))
						} else {
							h.sendChannel(chdr.TxId, &AsyncInvokeResp{
								Error: nil,
								Event: event},
							)
						}
					}
				}
			}
		}
	}
}

func (h *Handler) createChannel(txid string) (chan *AsyncInvokeResp, error) {
	h.Lock()
	defer h.Unlock()
	if h.responseChannel == nil {
		return nil, fmt.Errorf("[%s]Cannot create response channel", shorttxid(txid))
	}
	if h.responseChannel[txid] != nil {
		return nil, fmt.Errorf("[%s]Channel exists", shorttxid(txid))
	}
	c := make(chan *AsyncInvokeResp)
	h.responseChannel[txid] = c
	return c, nil
}

func (h *Handler) sendChannel(txId string, msg *AsyncInvokeResp) error {
	h.Lock()
	defer h.Unlock()
	if h.responseChannel == nil {
		return fmt.Errorf("[%s]Cannot send message response channel", shorttxid(txId))
	}
	if h.responseChannel[txId] == nil {
		return fmt.Errorf("[%s]sendChannel does not exist", shorttxid(txId))
	}

	h.responseChannel[txId] <- msg

	return nil
}

//sends a message and selects
func (h *Handler) sendReceive(c chan *AsyncInvokeResp) (*AsyncInvokeResp, error) {
	for {
		select {
		case <-time.After(time.Second * 180):
			return nil, errors.New("Block invoke time out")
		case outmsg, val := <-c:
			if !val {
				return nil, fmt.Errorf("unexpected failure on receive")
			}
			return outmsg, nil
		}
	}
}

func (h *Handler) deleteChannel(txid string) {
	h.Lock()
	defer h.Unlock()
	if h.responseChannel != nil {
		delete(h.responseChannel, txid)
	}
}

func (h *Handler) Invoke(peerClients []*cp.PeerClient, ordererClients []*orderer.OrdererClient, nonce []byte, carrier map[string]string, checkAccount bool, asyncChan chan *AsyncInvokeResp, args ...string) (string, error) {
	if checkAccount {
		once.Do(func() { go h.listenerParse() })
	}
	_, txid, err := h.invokeOrQuery(peerClients, ordererClients, checkAccount, asyncChan, nonce, args, carrier)
	if err != nil {
		return txid, err
	}
	return txid, nil
}

func (h *Handler) Query(peerClients []*cp.PeerClient, carrier map[string]string, args ...string) ([]*protos_peer.ProposalResponse, string, error) {
	resps, txid, err := h.invokeOrQuery(peerClients, nil, false, nil, nil, args, carrier)
	if err != nil {
		return nil, txid, err
	} else if len(resps) == 0 {
		return nil, txid, fmt.Errorf("Query function(%s) return null", args)
	}
	return resps, txid, nil
}

func (h *Handler) invokeOrQuery(clients []*cp.PeerClient, ordererClients []*orderer.OrdererClient, checkAccount bool, asyncChan chan *AsyncInvokeResp, nonce []byte, args []string, carrier map[string]string) ([]*protos_peer.ProposalResponse, string, error) {
	if len(clients) == 0 {
		return nil, "", fmt.Errorf("No available peers")
	}

	var prop *protos_peer.Proposal
	var err error
	var txId string
	prop, txId, err = pkg_common.CreateProposal(h.CCName, h.ChannelId, nonce, args, protcos_common.HeaderType_ENDORSER_TRANSACTION, carrier)
	if err != nil {
		return nil, txId, fmt.Errorf("Error creating proposal  invokeOrQuery: %s", err)
	}

	handlerFunc := func() ([]*protos_peer.ProposalResponse, *protos_peer.ChaincodeEvent, error) {
		signedProp, err := user.GetSignedProposal(prop)
		if err != nil {
			return nil, nil, fmt.Errorf("Error process proposal %invokeOrQuery: %s", err)
		}

		resps, err := pkg_common.ProcessProposal(clients, signedProp)
		if err != nil {
			return resps, nil, fmt.Errorf("Error process proposal invokeOrQuery: %s", err)
		} else if len(resps) == 0 {
			return resps, nil, fmt.Errorf("Invoke function invokeOrQuery response null")
		}

		if len(ordererClients) == 0 {
			return resps, nil, nil
		}

		if checkAccount {
			var c chan *AsyncInvokeResp
			c, err = h.createChannel(txId)
			if err != nil {
				return resps, nil, err
			}
			defer h.deleteChannel(txId)

			// send the envelope for ordering
			if err = pkg_common.Broadcast(ordererClients, prop, resps); err != nil {
				return resps, nil, fmt.Errorf("Error sending transaction invokeOrQuery: %s", err)
			}

			msg, err := h.sendReceive(c)
			if err != nil {
				return resps, nil, err
			} else {
				return resps, msg.Event, msg.Error
			}
		} else {
			// send the envelope for ordering
			if err = pkg_common.Broadcast(ordererClients, prop, resps); err != nil {
				return resps, nil, fmt.Errorf("Error sending transaction invokeOrQuery: %s", err)
			} else {
				return resps, nil, nil
			}
		}
	}

	if asyncChan != nil {
		go func() {
			_, event, err := handlerFunc()
			if event != nil {
				if event.TxId != txId {
					asyncChan <- NewAsyncResp(txId, fmt.Errorf("The resp event txid is not fit!"))
				}
				asyncChan <- &AsyncInvokeResp{Error: err, Event: event}
			} else {
				asyncChan <- NewAsyncResp(txId, err)
			}
		}()
		return nil, txId, nil
	}

	resps, _, err := handlerFunc()
	return resps, txId, err
}

func (h *Handler) Close() {
	pkg_common.Close()
}
