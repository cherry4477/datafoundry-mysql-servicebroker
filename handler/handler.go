package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/asiainfoldp/datafoundry-mysql-servicebroker/servicebrokers"
	"github.com/coreos/etcd/client"
	"github.com/gin-gonic/gin"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-golang/lager"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var Logger lager.Logger
var ConfigPath string
var etcdclient EtcdClient
var servcieBrokerName string

const MyCatServcieBroker = "MyCat_MySQL"

func init() {
	servcieBrokerName = "mycatshare"
	ConfigPath = GetENV("MYCATCLUSTERPATH")
	Logger = lager.NewLogger(MyCatServcieBroker)
	Logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	etcdclient = EtcdClient{}

}

func Catalog(c *gin.Context) {

	catalog := brokerapi.CatalogResponse{}

	resp, err := etcdclient.GetEtcdApi().Get(context.Background(), "/servicebroker/"+servcieBrokerName+"/catalog", &client.GetOptions{Recursive: true}) //改为环境变量
	if err != nil {
		Logger.Error("Can not get catalog information from etcd", err) //所有这些出错消息最好命名为常量，放到开始的时候
		catalog.Services = []brokerapi.Service{}
		c.JSON(http.StatusInternalServerError, catalog)
		return
	} else {
		Logger.Debug("Successful get catalog information from etcd. NodeInfo is " + resp.Node.Key)
	}

	services := []brokerapi.Service{}
	for i := 0; i < len(resp.Node.Nodes); i++ {
		Logger.Debug("Start to Parse Service " + resp.Node.Nodes[i].Key)
		myService := brokerapi.Service{}
		myService.ID = strings.Split(resp.Node.Nodes[i].Key, "/")[len(strings.Split(resp.Node.Nodes[i].Key, "/"))-1]
		for j := 0; j < len(resp.Node.Nodes[i].Nodes); j++ {
			if !resp.Node.Nodes[i].Nodes[j].Dir {
				lowerkey := strings.ToLower(resp.Node.Nodes[i].Key)
				switch strings.ToLower(resp.Node.Nodes[i].Nodes[j].Key) {
				case lowerkey + "/name":
					myService.Name = resp.Node.Nodes[i].Nodes[j].Value
				case lowerkey + "/description":
					myService.Description = resp.Node.Nodes[i].Nodes[j].Value
				case lowerkey + "/bindable":
					myService.Bindable, _ = strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Value)
				case lowerkey + "/tags":
					myService.Tags = strings.Split(resp.Node.Nodes[i].Nodes[j].Value, ",")
				case lowerkey + "/planupdatable":
					myService.PlanUpdatable, _ = strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Value)
				case lowerkey + "/metadata":
					json.Unmarshal([]byte(resp.Node.Nodes[i].Nodes[j].Value), &myService.Metadata)
				}
			} else if strings.HasSuffix(strings.ToLower(resp.Node.Nodes[i].Nodes[j].Key), "plan") {

				myPlans := []brokerapi.ServicePlan{}
				for k := 0; k < len(resp.Node.Nodes[i].Nodes[j].Nodes); k++ {
					Logger.Debug("Start to Parse Plan " + resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)
					myPlan := brokerapi.ServicePlan{}
					myPlan.ID = strings.Split(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key, "/")[len(strings.Split(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key, "/"))-1]
					for n := 0; n < len(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes); n++ {
						lowernodekey := strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Key)
						switch strings.ToLower(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Key) {
						case lowernodekey + "/name":
							myPlan.Name = resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value
						case lowernodekey + "/description":
							myPlan.Description = resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value
						case lowernodekey + "/free":
							//这里没有搞懂为什么brokerapi里面的这个bool要定义为传指针的模式
							myPlanfree, _ := strconv.ParseBool(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value)
							myPlan.Free = brokerapi.FreeValue(myPlanfree)
						case lowernodekey + "/metadata":
							json.Unmarshal([]byte(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value), &myPlan.Metadata)
						case lowernodekey + "/schemas":
							json.Unmarshal([]byte(resp.Node.Nodes[i].Nodes[j].Nodes[k].Nodes[n].Value), &myPlan.Schemas)
						}
					}
					//装配plan需要返回的值，按照有多少个plan往里面装
					myPlans = append(myPlans, myPlan)
				}
				//将装配好的Plan对象赋值给Service
				myService.Plans = myPlans

			}
		}

		//装配catalog需要返回的值，按照有多少个服务往里面装
		services = append(services, myService)
	}

	catalog.Services = services

	c.JSON(http.StatusOK, catalog)
	return
}

func Provision(c *gin.Context) {

	errorRep := brokerapi.ErrorResponse{}
	instance_id := c.Param("instance_id")
	accepts_incomplete := c.Query("accepts_incomplete")
	asyncAllowed := false
	if accepts_incomplete == "true" {
		asyncAllowed = true
	}

	rBody, _ := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	details := brokerapi.ProvisionDetails{}
	err := json.Unmarshal(rBody, &details)
	if err != nil {
		errorRep.Error = err.Error()
		errorRep.Description = "ProvisionDetails format error"
		c.JSON(http.StatusBadRequest, errorRep)
		return
	}

	//判断实例是否已经存在，如果存在就报错
	resp, err := etcdclient.Etcdget("/servicebroker/" + servcieBrokerName + "/instance") //改为环境变量

	if err != nil {
		Logger.Error("Can't connet to etcd", err)
		errorRep.Error = err.Error()
		errorRep.Description = "Can't connet to etcd"
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	}

	for i := 0; i < len(resp.Node.Nodes); i++ {
		if resp.Node.Nodes[i].Dir && strings.HasSuffix(resp.Node.Nodes[i].Key, instance_id) {
			Logger.Info("ErrInstanceAlreadyExists")

			errorRep.Error = errors.New("instance already exists").Error()
			errorRep.Description = "instance-already-exists"
			c.JSON(http.StatusConflict, errorRep)
			return
		}
	}

	//判断servcie_id和plan_id是否正确
	service_name := findServiceNameInCatalog(details.ServiceID)
	plan_name := findServicePlanNameInCatalog(details.ServiceID, details.PlanID)
	if service_name == "" || plan_name == "" {
		Logger.Info("Service_id or plan_id not correct!!")
		errorRep.Error = errors.New("Service_id or plan_id not correct!!").Error()
		errorRep.Description = "Service_id or plan_id not correct!!"
		c.JSON(http.StatusBadRequest, errorRep)
		return
	}

	myHandler, err := servicebrokers.New(MyCatServcieBroker)
	if err != nil {
		Logger.Error("Can not found handler for service "+service_name+" plan "+plan_name, err)
		errorRep.Error = err.Error()
		errorRep.Description = "Can not found handler for service " + service_name + " plan " + plan_name
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	}
	provsiondetail, myServiceInfo, err := myHandler.DoProvision(instance_id, details, asyncAllowed)
	if err != nil {
		Logger.Error("Error do handler for service "+service_name+" plan "+plan_name, err)
		errorRep.Error = err.Error()
		errorRep.Description = "Error do handler for service " + service_name + " plan " + plan_name
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	}

	succeeded := false
	defer func() {
		if succeeded {
			return
		}

		// if etcd failed to save, then deprovision the just created service.
		// The reason is bsi controller will call provision periodically until provision succeeds.

		Logger.Info("ETCD save failed, so to Deprovision service " + service_name + " plan " + plan_name)

		_, err := myHandler.DoDeprovision(&myServiceInfo, asyncAllowed)
		if err != nil {
			Logger.Error("ETCD save failed, do Deprovision service "+service_name+" plan "+plan_name+", but error:", err)
			errorRep.Error = err.Error()
			errorRep.Description = "ETCD save failed, do Deprovision service " + service_name + " plan " + plan_name
			c.JSON(http.StatusInternalServerError, errorRep)
			return
		}
	}()

	myServiceInfo.Service_name = service_name
	myServiceInfo.Plan_name = plan_name

	//写入etcd 话说如果这个时候写入失败，那不就出现数据不一致的情况了么！todo
	//先创建instanceid目录
	_, err = etcdclient.GetEtcdApi().Set(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instance_id, "", &client.SetOptions{Dir: true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		Logger.Error("Can not create instance "+instance_id+" in etcd", err) //todo都应该改为日志key
		errorRep.Error = err.Error()
		errorRep.Description = "Can not create instance " + instance_id + " in etcd"
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	} else {
		Logger.Debug("Successful create instance "+instance_id+" in etcd", nil)
	}

	//然后创建一系列属性
	etcdclient.Etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/organization_guid", details.OrganizationGUID)
	etcdclient.Etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/space_guid", details.SpaceGUID)
	etcdclient.Etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/service_id", details.ServiceID)
	etcdclient.Etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/plan_id", details.PlanID)
	tmpval, _ := json.Marshal(details.RawParameters)
	etcdclient.Etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/parameters", string(tmpval))
	etcdclient.Etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/dashboardurl", provsiondetail.DashboardURL)
	//存储隐藏信息_info
	tmpval, _ = json.Marshal(myServiceInfo)
	etcdclient.Etcdset("/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/_info", string(tmpval))

	//创建绑定目录
	_, err = etcdclient.GetEtcdApi().Set(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instance_id+"/binding", "", &client.SetOptions{Dir: true})
	if err != nil {
		Logger.Error("Can not create banding directory of  "+instance_id+" in etcd", err) //todo都应该改为日志key
		errorRep.Error = err.Error()
		errorRep.Description = "Can not create banding directory of  " + instance_id + " in etcd"
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	} else {
		Logger.Debug("Successful create banding directory of  "+instance_id+" in etcd", nil)
	}

	//完成所有操作后，返回DashboardURL和是否异步的标志
	Logger.Info("Successful create instance " + instance_id)

	succeeded = true

	prvsRsp := brokerapi.ProvisioningResponse{
		DashboardURL: provsiondetail.DashboardURL,
	}

	if provsiondetail.IsAsync {
		c.JSON(http.StatusAccepted, prvsRsp)
		return
	}
	c.JSON(http.StatusCreated, prvsRsp)
	return

}

func Deprovision(c *gin.Context) {

	var myServiceInfo servicebrokers.ServiceInfo

	errorRep := brokerapi.ErrorResponse{}

	instance_id := c.Param("instance_id")
	accepts_incomplete := c.Query("accepts_incomplete")
	asyncAllowed := false
	if accepts_incomplete == "true" {
		asyncAllowed = true
	}

	rBody, _ := ioutil.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	details := brokerapi.ProvisionDetails{}
	err := json.Unmarshal(rBody, &details)
	if err != nil {
		errorRep.Error = err.Error()
		errorRep.Description = "ProvisionDetails format error"
		c.JSON(http.StatusBadRequest, errorRep)
		return
	}

	//判断实例是否已经存在，如果不存在就报错
	resp, err := etcdclient.GetEtcdApi().Get(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instance_id, &client.GetOptions{Recursive: true})

	if err != nil || !resp.Node.Dir {
		Logger.Error("Can not get instance information from etcd", err)
		errorRep.Error = err.Error()
		errorRep.Description = "instance does not exist"
		c.JSON(http.StatusGone, errorRep)
		return
	} else {
		Logger.Debug("Successful get instance information from etcd. NodeInfo is " + resp.Node.Key)
	}

	var servcie_id, plan_id string
	//从etcd中取得参数。
	for i := 0; i < len(resp.Node.Nodes); i++ {
		if !resp.Node.Nodes[i].Dir {
			switch strings.ToLower(resp.Node.Nodes[i].Key) {
			case strings.ToLower(resp.Node.Key) + "/service_id":
				servcie_id = resp.Node.Nodes[i].Value
			case strings.ToLower(resp.Node.Key) + "/plan_id":
				plan_id = resp.Node.Nodes[i].Value
			}
		}
	}

	//并且要核对一下detail里面的service_id和plan_id。出错消息现在是500，需要更改一下源代码，以便更改出错代码
	if servcie_id != details.ServiceID || plan_id != details.PlanID {
		Logger.Info("Service_id or plan_id not correct!!")
		errorRep.Error = errors.New("Service_id or plan_id not correct!!").Error()
		errorRep.Description = "Service_id or plan_id not correct!!"
		c.JSON(http.StatusBadRequest, errorRep)
		return
	}
	//是否要判断里面有没有绑定啊？todo

	//根据存储在etcd中的service_name和plan_name来确定到底调用那一段处理。注意这个时候不能像Provision一样去catalog里面读取了。
	//因为这个时候的数据不一定和创建的时候一样，plan等都有可能变化。同样的道理，url，用户名，密码都应该从_info中解码出来

	//隐藏属性不得不单独获取
	resp, err = etcdclient.Etcdget("/servicebroker/" + servcieBrokerName + "/instance/" + instance_id + "/_info")
	json.Unmarshal([]byte(resp.Node.Value), &myServiceInfo)

	//生成具体的handler对象
	myHandler, err := servicebrokers.New(MyCatServcieBroker)

	//没有找到具体的handler，这里如果没有找到具体的handler不是由于用户输入的，是不对的，报500错误
	if err != nil {
		Logger.Error("Can not found handler for service "+myServiceInfo.Service_name+" plan "+myServiceInfo.Plan_name, err)
		errorRep.Error = err.Error()
		errorRep.Description = "Can not found handler for service " + myServiceInfo.Service_name + " plan " + myServiceInfo.Plan_name
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	}

	//执行handler中的命令
	deprovisionServiceSpec, err := myHandler.DoDeprovision(&myServiceInfo, asyncAllowed)

	//如果出错
	if err != nil {
		Logger.Error("Error do handler for service "+myServiceInfo.Service_name+" plan "+myServiceInfo.Plan_name, err)
		errorRep.Error = err.Error()
		errorRep.Description = "Error do handler for service " + myServiceInfo.Service_name + " plan " + myServiceInfo.Plan_name
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	}

	//然后删除etcd里面的纪录，这里也有可能有不一致的情况
	_, err = etcdclient.GetEtcdApi().Delete(context.Background(), "/servicebroker/"+servcieBrokerName+"/instance/"+instance_id, &client.DeleteOptions{Recursive: true, Dir: true}) //todo这些要么是常量，要么应该用环境变量
	if err != nil {
		errorRep.Error = err.Error()
		errorRep.Description = "Can not delete instance " + instance_id + " in etcd"
		c.JSON(http.StatusInternalServerError, errorRep)
		return
	} else {
		Logger.Debug("Successful delete instance " + instance_id + " in etcd")
	}

	Logger.Info("Successful Deprovision instance " + instance_id)

	if deprovisionServiceSpec.IsAsync {
		c.JSON(http.StatusAccepted, brokerapi.EmptyResponse{})
		return
	}
	c.JSON(http.StatusOK, brokerapi.EmptyResponse{})
	return

}

func findServiceNameInCatalog(service_id string) string {
	resp, err := etcdclient.Etcdget("/servicebroker/" + servcieBrokerName + "/catalog/" + service_id + "/name")
	if err != nil {
		Logger.Error("Can not get "+"/servicebroker/"+servcieBrokerName+"/catalog/"+service_id+"/name"+" from etcd", err)
		return ""
	}
	return resp.Node.Value
}

func findServicePlanNameInCatalog(service_id, plan_id string) string {
	resp, err := etcdclient.Etcdget("/servicebroker/" + servcieBrokerName + "/catalog/" + service_id + "/plan/" + plan_id + "/name")
	if err != nil {
		Logger.Error("Can not get "+"/servicebroker/"+servcieBrokerName+"/catalog/"+service_id+"/plan/"+plan_id+"/name"+" from etcd", err)
		return ""
	}
	return resp.Node.Value
}
