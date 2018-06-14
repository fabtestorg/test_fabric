package listenevent

import (
	"fmt"
	"github.com/peersafe/gohfc"
	"github.com/fabtestorg/test_fabric/eventserver/handle"
)

type FilterHandler func(*gohfc.EventBlockResponseTransactionEvent) (interface{}, bool)

func ListenEvent(eventAddress, chainID string, filterHandler FilterHandler, toHandle chan handle.BlockInfoAll) error {
	fmt.Printf("chainId : %s\n",chainID)
	notfy, err := gohfc.GetHandler().ListenEventFullBlock("peer", chainID)
	if err != nil {
		return err
	}

	if filterHandler == nil {
		return fmt.Errorf("The filter handler is null!")
	}

	for {
		select {
		case b := <-notfy:
			fmt.Println("----read response----no: ",b.BlockHeight)
			for _, r := range b.Transactions {
				//filter msg from chiancode event
				if len(r.Events) == 0 {
					continue
				}
				filterHandler(&r.Events[0])
				//msg, ok := filterHandler(&r.Events[0])
				////send msg/blockNum/txIndex to handle module
				//if ok {
				//	//fmt.Printf("---blockHeight = %d ---txIndex = %d---\n",b.BlockHeight,txIndex)
				//	blockInfo := handle.BlockInfoAll{
				//		BlockInfo: handle.BlockInfo{Block_number: b.BlockHeight,
				//			Tx_index: txIndex},
				//		MsgInfo: msg,
				//	}
				//	toHandle <- blockInfo
				//}
			}
		}
	}
	return nil
}
