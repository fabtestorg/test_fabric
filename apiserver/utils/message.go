package utils

import (
	"encoding/json"
	"encoding/base64"
	"fmt"

	"github.com/fabtestorg/test_fabric/apiserver/define"
)

// FormatRequestMessage format requset to message json
func FormatRequestMessage(request define.Factor) ([]byte, error) {
	//CryptoAlgorithm is used by hoperun, we don't need to care about it. By wuxu, 20170901.
	//request.CryptoAlgorithm = "aes"
	var invokeRequest define.InvokeRequest
	peersafeData := define.PeersafeData{}
	peersafeData.Keys = make(map[string][]byte)
	key := GenerateKey(32)
	if request.CryptoFlag == 1 {
		cryptoData, err := AesEncrypt([]byte(request.TxData), key)
		if err != nil {
			return nil, fmt.Errorf("AesEncrypt data error : %s", err.Error())
		}
		request.TxData = base64.StdEncoding.EncodeToString(cryptoData)
		cryptoRecvSend := append(request.Receiver, request.Sender)
		for _, receiver := range cryptoRecvSend {
			path := define.CRYPTO_PATH + receiver + "/" + "enrollment.cert"
			cert, err := ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("ReadFile %s cert error : %s", receiver, err.Error())
			}

			cryptoKey, err := EciesEncrypt(key, cert)
			if err != nil {
				return nil, fmt.Errorf("EciesEncrypt %s reandom key error : %s", receiver, err.Error())
			}

			peersafeData.Keys[receiver] = cryptoKey
		}
	}

	message := &define.Message{}
	message.CreateBy = request.CreateBy
	message.CreateTime = request.CreateTime
	message.Sender = request.Sender
	message.Receiver = request.Receiver
	message.TxData = request.TxData
	message.LastUpdateTime = request.LastUpdateTime
	message.LastUpdateBy = request.LastUpdateBy
	message.CryptoFlag = request.CryptoFlag
	message.CryptoAlgorithm = request.CryptoAlgorithm
	message.DocType = request.DocType
	message.FabricTxId = request.FabricTxId
	message.BusinessNo = request.BusinessNo
	message.Expand1 = request.Expand1
	message.Expand2 = request.Expand2
	message.DataVersion = request.DataVersion
	message.PeersafeData = peersafeData

	b, _ := json.Marshal(message)
	invokeRequest.Value = string(b)
	invokeRequest.Key = request.FabricTxId

	return json.Marshal(invokeRequest)
}

// FormatResponseMessage format response to message json
func FormatResponseMessage(userId string, request *[]define.Factor, messages *[]define.Message) error {
	for i := 0; i < len(*messages); i++ {
		item := (*messages)[i]
		if item.CryptoFlag == 1 {
			cryptoKey, ok := item.PeersafeData.Keys[userId]
			if ok {
				path := define.CRYPTO_PATH + userId + "/" + "enrollment.key"
				privateKey, err := ReadFile(path)
				if err == nil {
					key, err := EciesDecrypt(cryptoKey, privateKey)
					if err == nil {
						tempData, err := base64.StdEncoding.DecodeString(item.TxData)
						if err == nil {
							data, err := AesDecrypt(tempData, key)
							if err == nil {
								item.TxData = string(data)
							}
						}
					}
				}
			}
		}

		factor := define.Factor{
			CreateBy:        item.CreateBy,
			CreateTime:      item.CreateTime,
			Sender:          item.Sender,
			Receiver:        item.Receiver,
			TxData:          item.TxData,
			LastUpdateTime:  item.LastUpdateTime,
			LastUpdateBy:    item.LastUpdateBy,
			CryptoFlag:      item.CryptoFlag,
			CryptoAlgorithm: item.CryptoAlgorithm,
			DocType:         item.DocType,
			FabricTxId:      item.FabricTxId,
			BusinessNo:      item.BusinessNo,
			Expand1:         item.Expand1,
			Expand2:         item.Expand2,
			DataVersion:     item.DataVersion,
		}
		*request = append(*request, factor)
	}
	return nil
}
