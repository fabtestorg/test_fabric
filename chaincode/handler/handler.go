package handler

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/op/go-logging"
	"github.com/fabtestorg/test_fabric/chaincode/define"
	"github.com/fabtestorg/test_fabric/chaincode/utils"
	"os"
)

var KEEPALIVETEST = "test"

var myLogger = logging.MustGetLogger("hanldler")

func init() {
	format := logging.MustStringFormatter("%{shortfile} %{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter).SetLevel(logging.DEBUG, "hanldler")
}

func SaveData(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error
	request := &define.InvokeRequest{}
	if err = json.Unmarshal([]byte(args[1]), request); err != nil {
		return utils.InvokeResponse(stub, err, function, nil, false, args[0])
	}

	err = stub.PutState(request.Key, []byte(request.Value))
	if err != nil {
		myLogger.Errorf("saveData err: %s", err.Error())
		return utils.InvokeResponse(stub, err, function, nil, false, args[0])
	}

	var list []string
	list = append(list, request.Value)
	return utils.InvokeResponse(stub, nil, function, list, true, args[0])
}

func DslQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var err error
	request := &define.QueryRequest{}
	if err = json.Unmarshal([]byte(args[1]), request); err != nil {
		err = fmt.Errorf("DslQuery json decode args failed, err = %s", err.Error())
		return utils.QueryResponse(err, nil, request.SplitPage, args[0])
	}
	result, err := utils.GetValueByDSL(stub, request)
	if err != nil {
		return utils.QueryResponse(err, nil, request.SplitPage, args[0])
	}

	return utils.QueryResponse(nil, result, request.SplitPage, args[0])
}

func KeepaliveQuery(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
    targetValue, err := stub.GetState(KEEPALIVETEST)
    if err != nil {
        err = fmt.Errorf("ERROR! KeepaliveQuery get failed, err = %s", err.Error())
        return []byte("UnReached"), err 
    }   

    if string(targetValue) != KEEPALIVETEST {
        err = fmt.Errorf("ERROR! KeepaliveQuery get result is %s", string(targetValue))
        return []byte("UnReached"), err 
    }   

    return []byte("Reached"), nil 
}