package chaincode

import (
	protos_peer "github.com/hyperledger/fabric/protos/peer"
)

type AsyncInvokeResp struct {
	Event *protos_peer.ChaincodeEvent
	Error error
}

func NewAsyncResp(txId string, err error) *AsyncInvokeResp {
	return &AsyncInvokeResp{
		Error: err,
		Event: &protos_peer.ChaincodeEvent{
			TxId: txId,
		}}
}

func shorttxid(txid string) string {
	if len(txid) < 8 {
		return txid
	}
	return txid[0:8]
}
