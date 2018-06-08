package test

import (
	"encoding/json"

	"fmt"

	"github.com/fabtestorg/test_fabric/apiserver/define"
)

func getSaveDataRequest(user string) []byte {
	var factorList []define.Factor
	data := define.Factor{
		CreateBy:        user,
		Sender:          user,
		Receiver:        []string{user},
		TxData:          "test",
		CryptoFlag:      0,
		//CryptoAlgorithm is used by hoperun, we don't need to care about it. By wuxu, 20170901
		//CryptoAlgorithm: "aes",
	}

	factorList = append(factorList, data)

	request, _ := json.Marshal(factorList)
	fmt.Println(string(request))

	return request
}
