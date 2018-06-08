package handler

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"fmt"
	"encoding/json"

	logging "github.com/op/go-logging"
	"github.com/gin-gonic/gin"

	"github.com/fabtestorg/test_fabric/apiserver/define"
	"github.com/fabtestorg/test_fabric/apiserver/sdk"
	"github.com/fabtestorg/test_fabric/apiserver/utils"
	"math/rand"
	"time"
)

var logger = logging.MustGetLogger("handler")
var targetOrderAddr string

func init() {
	format := logging.MustStringFormatter("%{shortfile} %{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter).SetLevel(logging.DEBUG, "handler")
}

// SaveData 保存保理信息
func SaveData(c *gin.Context) {
	logger.Debug("SaveData.....")
	var request []define.Factor
	var err error
	responseStatus := &define.ResponseStatus{
		StatusCode: 0,
		StatusMsg:  "SUCCESS",
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	logger.Debugf("SaveData header : %v", c.Request.Header)
	logger.Debugf("SaveData body : %s", string(body))
	if err != nil {
		responseStatus.StatusCode = 1
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("SaveData read body : %s", err.Error())
		utils.Response(nil, c, http.StatusNoContent, responseStatus, nil)
		return
	}

	if err = json.Unmarshal(body, &request); err != nil {
		responseStatus.StatusCode = 1
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("SaveData Unmarshal : %s", err.Error())
		utils.Response(nil, c, http.StatusBadRequest, responseStatus, nil)
		return
	}
	logger.Debug(request)
	var txIds []string

	for _, factor := range request {
		factor.FabricTxId = GetRandomString(6)
		b, err := utils.FormatRequestMessage(factor)
		if err != nil {
			responseStatus.StatusCode = 1
			responseStatus.StatusMsg = err.Error()
			logger.Errorf("SaveData FormatRequestMessage : %s", err.Error())
			utils.Response(nil, c, http.StatusBadRequest, responseStatus, nil)
			return
		}

		// invoke
		txId, err := sdk.Handler.Invoke("", nil, "", define.SAVE_DATA, nil, b)
		if err != nil {
			responseStatus.StatusCode = 1
			responseStatus.StatusMsg = err.Error()
			logger.Errorf("SaveData Invoke : %s", err.Error())
			utils.Response(nil, c, http.StatusBadRequest, responseStatus, nil)
			return
		}
		txIds = append(txIds, txId)
	}

	utils.Response(txIds, c, http.StatusOK, responseStatus, nil)
}

// DslQuery 按条件查询信息
func DslQuery(c *gin.Context) {
	logger.Debug("DslQuery.....")
	var request define.QueryRequest
	var err error
	responseStatus := &define.ResponseStatus{
		StatusCode: 0,
		StatusMsg:  "SUCCESS",
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	logger.Debugf("DslQuery header : %v", c.Request.Header)
	logger.Debugf("DslQuery body : %s", string(body))
	if err != nil {
		responseStatus.StatusCode = 1
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("DslQuery read body : %s", err.Error())
		utils.Response(nil, c, http.StatusNoContent, responseStatus, nil)
		return
	}

	// query
	status := http.StatusOK
	requestPage := c.Request.Header.Get("page")
	json.Unmarshal([]byte(requestPage), &request.SplitPage)
	request.DslSyntax = string(body)
	response, retStatus, page, err := sdk.Handler.DSL("", define.DSL_QUERY, nil, request)
	if err != nil {
		responseStatus.StatusCode = 1
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("DslQuery Query : %s", err.Error())
		status = http.StatusBadRequest
	} else {
		responseStatus = retStatus
	}
	utils.Response(response, c, status, responseStatus, page)
}

// BlockQuery 获取指定业务编号相关的区块信息
func BlockQuery(c *gin.Context) {
	logger.Debug("BlockQuery.....")
	var request define.QueryRequest
	var err error
	responseStatus := &define.ResponseStatus{
		StatusCode: 0,
		StatusMsg:  "SUCCESS",
	}
	// query
	status := http.StatusOK
	businessNo := c.Param("id")
	requestPage := c.Request.Header.Get("page")
	json.Unmarshal([]byte(requestPage), &request.SplitPage)
	request.DslSyntax = fmt.Sprintf("{\"selector\":{\"businessNo\":\"%s\"}}", businessNo)
	request.BlockFlag = true
	b, _ := json.Marshal(request)
	responseData, retStatus, page, err := sdk.Handler.Query("", define.DSL_QUERY, nil, string(b), true)
	if err != nil {
		responseStatus.StatusCode = 1
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("DslQuery Query : %s", err.Error())
		status = http.StatusBadRequest
	} else {
		responseStatus = retStatus
	}

	txIdList, ok := responseData.Payload.([]string)
	if !ok {
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("DslQuery Query : %s", err.Error())
		status = http.StatusBadRequest
	}

	if len(txIdList) > 0 {
		response, retStatus, _, err := sdk.Handler.BlockQuery(txIdList)
		if err != nil {
			responseStatus.StatusCode = 1
			responseStatus.StatusMsg = err.Error()
			logger.Errorf("DslQuery Query : %s", err.Error())
			status = http.StatusBadRequest
		} else {
			responseStatus = retStatus
		}
		utils.Response(response, c, status, responseStatus, page)
	}
	utils.Response(define.QueryContents{}, c, status, responseStatus, page)
}

// BlockQuery 获取指定业务编号相关的区块信息
func BlockQueryEx(c *gin.Context) {
	logger.Debug("BlockQuery.....")
	var request define.QueryRequest
	var err error
	responseStatus := &define.ResponseStatus{
		StatusCode: 0,
		StatusMsg:  "SUCCESS",
	}
	// query
	status := http.StatusOK
	businessNo := c.Param("id")
	requestPage := c.Request.Header.Get("page")
	json.Unmarshal([]byte(requestPage), &request.SplitPage)
	request.DslSyntax = fmt.Sprintf("{\"selector\":{\"businessNo\":\"%s\"}}", businessNo)
	request.BlockFlag = true
	b, _ := json.Marshal(request)
	responseData, retStatus, page, err := sdk.Handler.Query("", define.DSL_QUERY, nil, string(b), true)
	if err != nil {
		responseStatus.StatusCode = 1
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("DslQuery Query : %s", err.Error())
		status = http.StatusBadRequest
	} else {
		responseStatus = retStatus
	}

	txIdList, ok := responseData.Payload.([]string)
	if !ok {
		responseStatus.StatusMsg = err.Error()
		logger.Errorf("DslQuery Query : %s", err.Error())
		status = http.StatusBadRequest
	}

	if len(txIdList) > 0 {
		response, retStatus, _, err := sdk.Handler.BlockQueryEx(txIdList)
		if err != nil {
			responseStatus.StatusCode = 1
			responseStatus.StatusMsg = err.Error()
			logger.Errorf("DslQuery Query : %s", err.Error())
			status = http.StatusBadRequest
		} else {
			responseStatus = retStatus
		}
		utils.Response(response, c, status, responseStatus, page)
	}
	utils.Response(define.QueryContents{}, c, status, responseStatus, page)
}

func KeepaliveQuery(c *gin.Context) {
    status := http.StatusOK
    responseStatus := &define.ResponseStatus{
        StatusCode: 200,
        StatusMsg:  "SUCCESS",
    }   

    if !sdk.Handler.PeerKeepalive(define.KEEPALIVE_QUERY) {
        responseStatus.StatusCode = define.PEER_FAIL_CODE 
        responseStatus.StatusMsg = "Peer FAILED"
        status = define.PEER_FAIL_CODE
        logger.Error("peer cann't be reached.")
    } else if !OrderKeepalive() {
        responseStatus.StatusCode = define.ORDER_FAIL_CODE 
        responseStatus.StatusMsg = "Order FAILED"
        status = define.ORDER_FAIL_CODE
        logger.Error("order cann't be reached.")
    }   

    utils.Response(nil, c, status, responseStatus, nil)
}

func OrderKeepalive() bool {
    //use nc command to detect whether the order's port is available
    orderCommand := fmt.Sprintf("nc -v %s", targetOrderAddr)
    cmd := exec.Command("/bin/bash", "-c", orderCommand)
    err := cmd.Run()
    if nil != err {
        logger.Errorf("Order(%s) cann't be reached: %s", targetOrderAddr, err.Error())
        return false
    } else {
        return true
    }
}

func SetOrderAddrToProbe(addr string) bool {
	if addr == "" {
		logger.Error("order address to be Probed is null!")
		return false
	}

	targetOrderAddr = addr
	logger.Debug("order address to be Probed is", targetOrderAddr)

	return true
}
func  GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}