package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/fabtestorg/test_fabric/apiserver/router"
	"github.com/fabtestorg/test_fabric/apiserver/sdk"
	"github.com/fabtestorg/test_fabric/apiserver/handler"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

var (
	configPath = flag.String("configPath", "", "config file path")
	configName = flag.String("configName", "", "config file name")
)

var logger = logging.MustGetLogger("main")

func init() {
	format := logging.MustStringFormatter("%{shortfile} %{time:15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(backendFormatter).SetLevel(logging.DEBUG, "main")
}

func main() {
	// parse init param
	logger.Debug("Usage : ./apiserver -configPath= -configName=")
	flag.Parse()
	if *configPath == "" || *configName == "" {
		*configPath = "./"
		*configName = "client_sdk"
		logger.Debug("becase configPath or configName nil  so  auto set")
		logger.Debug("auto set  configPath = \"./\" , configName = \"client_sdk\"")
	}

	err := sdk.InitSDK(*configPath, *configName)
	if err != nil {
		logger.Errorf("init sdk error : %s\n", err.Error())
		panic(err)
	}
	// 设置使用系统最大CPU
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 运行模式
	gin.SetMode(gin.ReleaseMode) //DebugMode ReleaseMode

	// 构造路由器
	r := router.GetRouter()

	// 调试用,可以看到堆栈状态和所有goroutine状态
	//ginpprof.Wrapper(r)

	//Get the listen port for apiserver
	listenPort := viper.GetInt("apiserver.listenport")
	logger.Debug("The listen port is", listenPort)
	listenPortString := fmt.Sprintf(":%d", listenPort)

	if !handler.SetOrderAddrToProbe(viper.GetString("apiserver.probe_order")) {
		panic("can not get the order address to be probed.")
	}

	// 运行服务
	server := endless.NewServer(listenPortString, r)
	server.BeforeBegin = func(add string) {
		pid := syscall.Getpid()
		logger.Criticalf("Actual pid is %d", pid)
		// 保存pid文件
		pidFile := "apiserver.pid"
		if checkFileIsExist(pidFile) {
			os.Remove(pidFile)
		} else {
			if err := ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0666); err != nil {
				logger.Fatalf("Api server write pid file failed! err:%v\n", err)
			}
		}
	}
	err = server.ListenAndServe()
	if err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			logger.Errorf("%v\n", err)
		} else {
			logger.Errorf("Api server start failed! err:%v\n", err)
			panic(err)
		}
	}
}

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
