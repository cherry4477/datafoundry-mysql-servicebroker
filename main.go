package main

import (
	"fmt"
	"github.com/asiainfoldp/datafoundry-mysql-servicebroker/handler"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

var servcieBrokerName string = "mycatshare"
var serviceBrokerPort string
var usr, password string
var etcdclient handler.EtcdClient

func main() {
	//初始化BROKERPORT
	serviceBrokerPort = handler.GetENV("BROKERPORT")

	configPath := handler.ConfigPath

	//判断配置文件是否存在
	exist, err := handler.CheckFileIsExist(configPath + "myclusterconfig.yaml")
	if !exist {
		handler.Logger.Error("Can not find config file : "+configPath+"myclusterconfig.yaml !", err)
		os.Exit(1)
	}

	etcdclient = handler.EtcdClient{}

	resp, err := etcdclient.Etcdget("/servicebroker/" + servcieBrokerName + "/username")
	if err != nil {
		handler.Logger.Error("Can not init username,Progrom Exit!", err)
		os.Exit(1)
	} else {
		usr = resp.Node.Value
	}

	resp, err = etcdclient.Etcdget("/servicebroker/" + servcieBrokerName + "/password")
	if err != nil {
		handler.Logger.Error("Can not init password,Progrom Exit!", err)
		os.Exit(1)
	} else {
		password = resp.Node.Value
	}

	router := handle()
	s := &http.Server{
		Addr:    ":" + serviceBrokerPort,
		Handler: router,
		//ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 0,
	}
	fmt.Println("START SERVICE BROKER", servcieBrokerName)
	s.ListenAndServe()
}

func handle() (router *gin.Engine) {
	//设置全局环境：1.开发环境（datafoundry_docker.DebugMode） 2.线上环境（datafoundry_docker.ReleaseMode）
	gin.SetMode(gin.ReleaseMode)
	//获取路由实例
	router = gin.Default()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		usr: password,
	}))
	authorized.GET("/v2/catalog", handler.Catalog)
	authorized.PUT("/v2/service_instances/:instance_id", handler.Provision)
	authorized.DELETE("/v2/service_instances/:instance_id", handler.Deprovision)
	return
}
