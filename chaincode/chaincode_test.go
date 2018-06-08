/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"testing"

	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/op/go-logging"
	"github.com/peersafe/factoring/chaincode/define"
	"os"
	"time"
)

var myCCID = "test_fabric"

type unitObj struct {
	CreateBy       string   `json:"createBy"`       //创建者
	CreateTime     uint64   `json:"createTime"`     //开标时间
	Sender         string   `json:"sender"`         //发送者
	Receiver       []string `json:"receiver"`       //接收者列表
	LastUpdateTime uint64   `json:"lastUpdateTime"` //最近一次修改时间
	LastUpdateBy   string   `json:"lastUpdateBy"`   //最近一次修改者
	DocType        string   `json:"docType"`        //业务类型
	FabricTxId     string   `json:"fabricTxId"`     //Fabric交易id(uuid)
	BusinessNo     string   `json:"businessNo"`     //业务编号（交易编号）
	Keys           []string `json:"keys"`           //用receiver证书加密后的密钥
	RootId         string   `json:"rootId"`         // 系列交易id
	NextSender     []string `json:"nextSender"`     //下一笔交易发起人
	TxData         string   `json:"txData"`         //业务数据
	CryptoFlag     int      `json:"cryptoFlag"`     //加密标识（0:不加密，1:加密）
}

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func SetLogLevel(value logging.Level) {
	format := logging.MustStringFormatter("%{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(backendFormatter).SetLevel(value, "mock")
}

func Test_Invoke(t *testing.T) {
	scc := new(FactorChaincode)
	stub := shim.NewMockStub(myCCID, scc)
	SetLogLevel(logging.ERROR)
	checkInit(t, stub, [][]byte{[]byte("init")})

	fmt.Println("--------普通交易 测试savedata 接口 发送者 zhengfu0 接收者qiye0----------")
	checkSaveData(t, stub, "txId", "", "zhengfu0", "qiye0", "")
	fmt.Println("--------测试QueryDsl 接口 发送者为zhengfu0 (由于利用couchDB单元测试会有无法实现问题)")
	checkQueryRequest(t, stub, "zhengfu0")
}

func checkSaveData(t *testing.T, stub *shim.MockStub, txid string, rootid string, sender string, receiver string, next string) {
	tempValue := &unitObj{
		CreateBy:       sender,
		CreateTime:     getCurTime(),
		Sender:         sender,
		Receiver:       []string{receiver},
		LastUpdateTime: getCurTime(),
		LastUpdateBy:   sender,
		DocType:        "factorData",
		FabricTxId:     txid,
		BusinessNo:     "BusinessNo",
		Keys:           []string{"key1"},
		RootId:         rootid,
		NextSender:     []string{next},
		TxData:         "data",
		CryptoFlag:     0,
	}
	tempData, _ := json.Marshal(tempValue)
	def := &define.InvokeRequest{
		Key:   txid,
		Value: string(tempData),
	}
	data, _ := json.Marshal(def)
	args := [][]byte{[]byte("SaveData"), []byte("trackid"), data}
	res := stub.MockInvoke(txid, args)
	SaveResponse(res, t, sender)
}

func checkQueryRequest(t *testing.T, stub *shim.MockStub, sender string) {
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"factorData\",\"sender\":\"%s\"}}", sender)
	def := define.QueryRequest{
		DslSyntax: queryString,
		SplitPage: define.Page{1, 10, 0},
	}

	data, _ := json.Marshal(def)
	args := [][]byte{[]byte("DslQuery"), []byte("trackid"), data}
	res := stub.MockInvoke("1", args)

	if res.Status != shim.OK && res.Message != "Not Implemented" {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
}

//获取当前时间
func getCurTime() uint64 {
	return uint64(time.Now().UTC().Unix())
}

func SaveResponse(res peer.Response, t *testing.T, sender string) {
	if res.Status != shim.OK {
		fmt.Println("Invoke", "failed", string(res.Message))
		t.FailNow()
	} else {
		infoList := new([]define.InvokeRequest)
		response := &define.InvokeResponse{}
		response.Payload = infoList
		json.Unmarshal(res.Payload, response)
		var data map[string]interface{}
		json.Unmarshal([]byte((*infoList)[0].Value), &data)
		fmt.Println("------response----", data)
		if data["sender"].(string) != sender {
			fmt.Printf("unit test save response sender is not same %s\n", sender)
			t.FailNow()
		}
	}
	return
}
