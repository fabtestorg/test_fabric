package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/op/go-logging"
	"github.com/fabtestorg/test_fabric/chaincode/handler"
)

var logger = logging.MustGetLogger("factorChaincode")

type handlerFunc func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error)

var funcHandler = map[string]handlerFunc{
	"SaveData": handler.SaveData,
	"DslQuery": handler.DslQuery,
	"KeepaliveQuery": handler.KeepaliveQuery,
}

type FactorChaincode struct {
}

func init() {
	format := logging.MustStringFormatter("%{shortfile} %{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter).SetLevel(logging.DEBUG, "factorChaincode")
}

// Init method will be called during deployment.
// The deploy transaction metadata is supposed to contain the administrator cert
func (t *FactorChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("Init Chaincode...")

	err := stub.PutState(handler.KEEPALIVETEST, []byte(handler.KEEPALIVETEST))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("SUCCESS"))
}

func (t *FactorChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Debugf("Invoke function=%v,args=%v\n", function, args)

	if len(args) < 2 || len(args[1]) == 0 {
		logger.Error("the invoke args length < 2 or arg[1] is empty")
		return shim.Error("the invoke args length < 2 or arg[1] is empty")
	}

	currentFunc := funcHandler[function]
	if currentFunc == nil {
		logger.Error("the function name not exist!!")
		return shim.Error("the function name not exist!!")
	}

	payload, err := currentFunc(stub, function, args)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(payload)
}

func main() {
	err := shim.Start(new(FactorChaincode))
	if err != nil {
		fmt.Printf("Error starting Factorchaincode: %s", err)
	}
}
