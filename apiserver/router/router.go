package router

import (
	"net/http"

	"github.com/fabtestorg/test_fabric/apiserver/handler"

	"sync"

	"github.com/gin-gonic/gin"
)

// Router 全局路由
var router *gin.Engine
var onceCreateRouter sync.Once

func GetRouter() *gin.Engine {
	onceCreateRouter.Do(func() {
		router = createRouter()
	})

	return router
}

func createRouter() *gin.Engine {
	router := gin.Default()
	// 版本控制
	// v1 := Router.Group("/v1")
	// {
	// factor
	factor := router.Group("/factor")
	{
		factor.POST("/saveData", handler.SaveData)
		factor.POST("/dslQuery", handler.DslQuery)
		factor.HEAD("/keepaliveQuery", handler.KeepaliveQuery)
		factor.GET("/block/:id", handler.BlockQuery)
		factor.GET("/blockQuery/:id", handler.BlockQueryEx)
	}
	//upload schema json file
	router.StaticFS("/schema", http.Dir("./schema"))
	return router
}
