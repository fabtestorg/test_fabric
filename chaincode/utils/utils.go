package utils

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/op/go-logging"
	"github.com/fabtestorg/test_fabric/chaincode/define"
	"os"
)

var myLogger = logging.MustGetLogger("utils")

func init() {
	format := logging.MustStringFormatter("%{shortfile} %{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter).SetLevel(logging.DEBUG, "utils")
}

func InvokeResponse(stub shim.ChaincodeStubInterface, err error, function string, data interface{}, eventFlag bool, trackid string) ([]byte, error) {
	errString := ""
	code := 0
	if err != nil {
		errString = err.Error()
		code = 1
	} else {
		errString = "Success"
	}

	response := define.InvokeResponse{
		TrackId:   trackid,
		Payload:   data,
		ResStatus: define.ResponseStatus{code, errString},
	}

	payload, errTmp := json.Marshal(response)
	myLogger.Debug("**************************factoring response****************************")
	myLogger.Debug(string(payload))
	if errTmp != nil {
		myLogger.Debug("response Json  encode error.")
	}

	if eventFlag {
		myLogger.Debugf("Set event  %s\n.", function)
		if errTmp := stub.SetEvent(function, payload); errTmp != nil {
			myLogger.Errorf("set event error : %s", errTmp.Error())
		}
	}

	return payload, err
}

func QueryResponse(err error, data interface{}, pageItem define.Page, trackid string) ([]byte, error) {
	errString := ""
	code := 0
	if err != nil {
		errString = err.Error()
		code = 1
	} else {
		errString = "Success"
	}

	response := define.QueryResponse{
		TrackId:   trackid,
		Payload:   data,
		ResStatus: define.ResponseStatus{code, errString},
		Page:      pageItem,
	}

	payload, err := json.Marshal(response)
	myLogger.Debug("**************************QueryResponse****************************")
	myLogger.Debug(string(payload))
	if err != nil {
		myLogger.Debug("QueryResponse Json  encode error.")
	}

	return payload, err
}

func GetValueByDSL(stub shim.ChaincodeStubInterface, request *define.QueryRequest) (interface{}, error) {
	if len(request.DslSyntax) == 0 {
		return nil, fmt.Errorf(" dsl syntax is empty!")
	}

	resultsIterator, err := stub.GetQueryResult(request.DslSyntax)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	skipSize := (request.SplitPage.CurrentPage - 1) * request.SplitPage.PageSize
	needSize := request.SplitPage.PageSize

	var resultList []string
	request.SplitPage.TotalRecords = 0
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		request.SplitPage.TotalRecords++
		if skipSize > 0 && request.SplitPage.PageSize > 0 {
			skipSize--
			continue
		} else if needSize > 0 || request.SplitPage.PageSize == 0 {
			needSize--
			//if block query need txid which resultsIterator key
			if request.BlockFlag {
				resultList = append(resultList, queryResponse.Key)
			} else {
				resultList = append(resultList, string(queryResponse.Value))
			}
		}
	}
	return resultList, nil
}
