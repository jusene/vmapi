package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"vmapi/controller"
	_ "vmapi/docs"
)


// @title Virtual Machine Api
// @version 1.0
// @description Virtual Machine Api
// @schemes http
// @BasePath /api/v1
func main() {
	gin.SetMode("release")
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(Cors())
	v1 := router.Group("/api/v1")
	{
		v1.GET("/vms/:phy", controller.VMList)
		v1.POST("/vms", controller.VMCreate)
		v1.GET("/vms/:phy/:vm", controller.VMDetail)
		v1.DELETE("/vms/:phy/:vm", controller.VMDelete)
		v1.PUT("/vms/:phy/:vm", controller.VMController)

		v1.GET("/ips", controller.IPS)
		v1.POST("/ips", controller.IPSCreate)
		v1.DELETE("/ips/:ip", controller.IPRemove)
		v1.PUT("/ips/:ip", controller.IPAppend)

		v1.GET("/host/:host", controller.HOSTDetail)
		v1.GET("/hosts/vm/:host", controller.HOSTVMDetail)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.Run(":8998")
}

func Cors() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Header("Access-Control-Allow-Origin", "*")
		context.Header("Server", "GIN")
		context.Next()
	}
}
