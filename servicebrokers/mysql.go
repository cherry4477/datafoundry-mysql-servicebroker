package servicebrokers

import (
	"database/sql"
	"errors"
	"github.com/asiainfoldp/datafoundry-mysql-servicebroker/handler"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pivotal-cf/brokerapi"
	"github.com/pivotal-golang/lager"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
)

const MyCatServcieBroker = "MyCat_MySQL"

type clusters struct {
	MyClusters []myCluster `yaml:"clusters"`
}

type myCluster struct {
	ID           string      `yaml:"id"`
	Name         string      `yaml:"name"`
	Description  string      `yaml:"description"`
	Loadbalancer string      `yaml:"loadbalancer"`
	MyCats       []myCatNode `yaml:"mycats"`
	MySqls       []mySQLNode `yaml:"mysqls"`
}

type myCatNode struct {
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	Host          string `yaml:"host"`
	OSUser        string `yaml:"osuser"`
	OSPassword    string `yaml:"ospassword"`
	ConfigPath    string `yaml:"configpath"`
	DataPort      string `yaml:"dataport"`
	ManagePort    string `yaml:"manageport"`
	MyCatUser     string `yaml:"mycatuser"`
	MyCatPassword string `yaml:"mycatpassword"`
}

type mySQLNode struct {
	Url      string `yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

var m *sync.RWMutex
var etcdclient handler.EtcdClient

func init() {
	register(MyCatServcieBroker, &MyCat_sharedHandler{})
	m = new(sync.RWMutex)
	etcdclient = handler.EtcdClient{}

}

type MyCat_sharedHandler struct{}

//配置MyCat，然后重启MyCat
func (myhandler *MyCat_sharedHandler) DoProvision(instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, ServiceInfo, error) {

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	var dbName string
	if ATTR_dbname, ok := details.RawParameters["ATTR_dbname"]; ok {
		dbName = handler.CheckDBname(ATTR_dbname.(string))
	} else {
		dbName = "d" + handler.Getguid()[0:15]
	}

	myClusterID := details.ServiceID
	myClusterInfo, err := getMyClusterInfo(myClusterID)
	if err != nil {
		handler.Logger.Error("Get MyCatInfo false", err)
		return brokerapi.ProvisionedServiceSpec{}, ServiceInfo{}, err
	}

	m.Lock()
	mySqlDb, err := getMySqlInfo(myClusterInfo)
	m.Unlock()

	if err != nil {
		handler.Logger.Error("Get MySQLInfo false", err)
		return brokerapi.ProvisionedServiceSpec{}, ServiceInfo{}, err
	}
	db, err := sql.Open("mysql", mySqlDb.User+":"+mySqlDb.Password+"@tcp("+mySqlDb.Url+")/")
	defer db.Close()
	if err != nil {
		handler.Logger.Error("Open MySQL false", err)
		return brokerapi.ProvisionedServiceSpec{}, ServiceInfo{}, err
	}
	//测试是否能联通
	err = db.Ping()
	if err != nil {
		handler.Logger.Error("Ping MySQL false", err)
		return brokerapi.ProvisionedServiceSpec{}, ServiceInfo{}, err
	}

	//不能以instancdID为数据库名字，需要创建一个不带-的数据库名
	//dbname := "d" + getguid()[0:15]
	_, err = db.Query("CREATE DATABASE " + dbName + " DEFAULT CHARACTER SET utf8 DEFAULT COLLATE utf8_general_ci")

	if err != nil {
		handler.Logger.Error("Create database:"+dbName+" false", err)
		return brokerapi.ProvisionedServiceSpec{}, ServiceInfo{}, err
	}

	newusername := handler.Getguid()[0:15]
	newpassword := handler.Getguid()[0:15]

	_, err = db.Query("GRANT ALL ON " + dbName + ".* TO '" + newusername + "'@'%' IDENTIFIED BY '" + newpassword + "'")

	if err != nil {
		handler.Logger.Error("Grant database:"+dbName+" to user:"+newusername+" false", err)
		return brokerapi.ProvisionedServiceSpec{}, ServiceInfo{}, err
	}

	////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	myServiceInfo := ServiceInfo{
		Url:            mySqlDb.Url,
		Admin_user:     mySqlDb.User,
		Admin_password: mySqlDb.Password,
		Database:       dbName,
		User:           newusername,
		Password:       newpassword,
		Mycluster_id:   myClusterID,
	}

	provsiondetail := brokerapi.ProvisionedServiceSpec{IsAsync: false,
		//Credentials: buildCredentialsFromServiceInfo_MyCat_MySQL(&myServiceInfo, &mycatInfo)}
		Credentials: buildCredentialsFromServiceInfo_MySQL(&myServiceInfo)}

	return provsiondetail, myServiceInfo, nil
}

func (myhandler *MyCat_sharedHandler) DoDeprovision(myServiceInfo *ServiceInfo, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {

	/*
		m.Lock()
		rB := deleteSchema(myServiceInfo)
		m.Unlock()
		if !rB {
			logger.Error("Deleteschema function false", errors.New("MyCat cluster disabled！"))
			return brokerapi.IsAsync(false), errors.New("MyCat cluster disabled！")
		}
	*/
	db, err := sql.Open("mysql", myServiceInfo.Admin_user+":"+myServiceInfo.Admin_password+"@tcp("+myServiceInfo.Url+")/")
	defer db.Close()
	if err != nil {
		handler.Logger.Error("Open mysql false", err)
		return brokerapi.DeprovisionServiceSpec{IsAsync: false}, err
	}
	//测试是否能联通
	err = db.Ping()
	if err != nil {
		handler.Logger.Error("Ping mysql false", err)
		return brokerapi.DeprovisionServiceSpec{IsAsync: false}, err
	}

	//取消用户对数据库里所有表权限
	_, err = db.Query("REVOKE ALL ON " + myServiceInfo.Database + ".* FROM '" + myServiceInfo.User + "'@'%'")
	if err != nil {
		handler.Logger.Error("Revoke user:"+myServiceInfo.User+"from database:"+myServiceInfo.Database+" false", err)
		return brokerapi.DeprovisionServiceSpec{IsAsync: false}, err
	}

	return brokerapi.DeprovisionServiceSpec{IsAsync: false}, err

}

func (handler *MyCat_sharedHandler) DoLastOperation(myServiceInfo *ServiceInfo) (brokerapi.LastOperation, error) {

	return brokerapi.LastOperation{
		State:       brokerapi.Succeeded,
		Description: "It's a sync method!",
	}, nil
}

func (handler *MyCat_sharedHandler) DoBind(myServiceInfo *ServiceInfo, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, Credentials, error) {

	return brokerapi.Binding{}, Credentials{}, nil
}

func (handler *MyCat_sharedHandler) DoUnbind(myServiceInfo *ServiceInfo, mycredentials *Credentials) error {

	return nil
}

func getMySqlInfo(mycluster myCluster) (mySQLNode, error) {

	mysql := mySQLNode{}
	resp, err := etcdclient.Etcdget("/servicebroker/mycatshare/catalog/" + mycluster.ID + "/selection")
	if err != nil {
		handler.Logger.Error("Can not get catalog information from etcd", err)
		return mysql, err
	} else {
		handler.Logger.Debug("Successful get catalog information from etcd. NodeInfo is " + resp.Node.Key)
	}
	selectIndex, _ := strconv.Atoi(resp.Node.Value)
	total := len(mycluster.MySqls)
	if total > 0 {
		mysql = mycluster.MySqls[selectIndex%total]
	}
	selectIndex += 1
	if selectIndex >= total {
		selectIndex = 0
	}
	etcdclient.Etcdset("/servicebroker/mycatshare/catalog/"+mycluster.ID+"/selection", strconv.Itoa(selectIndex))

	return mysql, nil
}

func getMyClusterInfo(myclusterid string) (myCluster, error) {

	mycluster := myCluster{}
	if myclusterid == "" {
		handler.Logger.Debug("Clusterid is empty")
		return mycluster, errors.New("Clusterid is empty")
	}
	data, err := ioutil.ReadFile(handler.ConfigPath + "myclusterconfig.yaml")
	if err != nil {
		handler.Logger.Error("Read "+handler.ConfigPath+"myclusterconfig.yaml error", err)
		return mycluster, err
	}

	t := clusters{}
	//把yaml形式的字符串解析成struct类型
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		handler.Logger.Error("Parsing "+handler.ConfigPath+"myclusterconfig.yaml error", err)
		return mycluster, err
	}

	for i := 0; i < len(t.MyClusters); i++ {
		if t.MyClusters[i].ID == myclusterid {
			mycluster = t.MyClusters[i]
			break
		}
	}
	return mycluster, nil
}

func buildCredentialsFromServiceInfo_MySQL(myServiceInfo *ServiceInfo) Credentials {
	newusername := myServiceInfo.User
	newpassword := myServiceInfo.Password
	hostAndport := strings.Split(myServiceInfo.Url, ":")

	return Credentials{
		Uri:      "mysql://" + newusername + ":" + newpassword + "@" + myServiceInfo.Url + "/" + myServiceInfo.Database,
		Hostname: hostAndport[0],
		Port:     hostAndport[1],
		Username: newusername,
		Password: newpassword,
		Name:     myServiceInfo.Database,
	}

}
