# datafoundry-mysql-servicebroker
本程序基于GIN框架，遵循Openshift ServiceBroker 2.13规范的MySQL ServiceBroker，版本为v1。


### 需要用户设置数据库库名（非必要）
字段名称：ATTR_dbname
注意：数据库库名遵循mysql官方规范，允许字符A~Z，a~z，0~9，$，_；库名不能以数字开头。任何不合法本程序均以"_"替换，未设定此字段或此字段值为空，则程序自动生成"d"开头的数据库库名。


### 需要的环境变量

ETCD服务入口:
* ETCDENDPOINT

ETCD用户名:
* ETCDUSER

ETCD密码:
* ETCDPASSWORD

服务监听端口:
* BROKERPORT

集群配置文件路径:
* MYCATCLUSTERPATH


### 挂卷的配置文件

$MYCATCLUSTERPATH+myclusterconfig.yaml:具体实例参考myclusterconfig.yaml，需注意每个集群的id必须唯一且与ETCD保存的ServiceID保持一致，需要人员维护。


### 生成镜像命令

工程根目录下：

*先输入make

*docker build -t mybroker:v1 .

*如果删除Makefile产生的多余目录及文件输入make clean



### 镜像运行命令

**docker run -p <宿主端口号>:<程序监听端口号> -e ETCDENDPOINT=<ETCD服务入口> -e ETCDUSER=<ETCD用户名> -e ETCDPASSWORD=<ETCD密码> -e BROKERPORT=<程序监听端口号> -e MYCATCLUSTERPATH=<集群配置文件路径> -v <宿主机配置文件路径>:<环境变量配置路径> <镜像名称>:<版本Tag>**

事例：docker run -p 8000:10013 -e ETCDENDPOINT="http://192.168.1.114:2379" -e ETCDUSER="root" -e ETCDPASSWORD="111111" -e BROKERPORT="10013" -e MYCATCLUSTERPATH="/mnt/" -v /xx/:/mnt/ mybroker:v1



### 服务注册

参考工程下registery/etcdinit.sh


### 镜像与ETCD容易出现的问题

单机版ETCD，启动服务时
etcd  -listen-client-urls  http://192.168.1.114:2379  -advertise-client-urls  http://192.168.1.114:2379
集群ETCD，启动服务时
etcd  -listen-client-urls  http://0.0.0.0:2379  -advertise-client-urls  http://192.168.1.114:2379


### API接口

#### GET /v2/catalog

获取服务列表。

curl样例：
```
curl -i -X GET http://asiainfoLDP:2016asia@127.0.0.1:10013/v2/catalog
```


#### PUT /v2/service_instances/{instanceID}

创建MyCat实例。

Path参数
* `instanceID`: 实例ID，一般程序自动生成唯一ID。

curl样例：
```
curl -i -X PUT http://asiainfoLDP:2016asia@127.0.0.1:10013/v2/service_instances/wu_test_001?accepts_incomplete=true -d '{
 "service_id":"EB1045DE-5FEA-4A60-9455-3727E1C0FD1C",
 "plan_id":"619A6421-386E-4684-BBB4-915E406D58F9",
 "organization_guid": "default",
 "space_guid":"space-guid",
 "parameters": {"ATTR_dbname":"dtest001"}
}'  -H "Content-Type: application/json"
```
注：样例中service_id字段同myclusterconfig.yaml里找到对应的集群ID必须一致，即每一个服务对应一个集群

回应样例（JSON格式）：
```
{"credentials":{"uri":"mysql://8b40c1cc3d88691:015e7bfbf188b6c@10.12.1.171:8777/d24142a599deb70f","host":"10.12.1.171","port":"8777","username":"8b40c1cc3d88691","password":"015e7bfbf188b6c","name":"d24142a599deb70f"}}
```


#### DELETE /v2/service_instances/{instanceID}

删除MyCat实例。

Path参数
* `instanceID`: 实例ID，一般程序自动生成唯一ID。

curl样例：
```
curl -i -X DELETE http://asiainfoLDP:2016asia@127.0.0.1:10013/v2/service_instances/wu_test_001?accepts_incomplete=true -d '{
"service_id":"EB1045DE-5FEA-4A60-9455-3727E1C0FD1C",
"plan_id":"619A6421-386E-4684-BBB4-915E406D58F9"
}'  -H "Content-Type: application/json"
```
