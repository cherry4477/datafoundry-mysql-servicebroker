#!/bin/bash

ETCD_USER=xxx
ETCD_PASSWORD=xxx
ETCD_ADDRESS=http://xxx:2379

API_NAME=xxx
API_PASSWORD=xxx

export ETCDCTL="etcdctl --timeout 15s --total-timeout 30s --endpoints $ETCD_ADDRESS --username $ETCD_USER:$ETCD_PASSWORD"

$ETCDCTL mkdir /servicebroker
$ETCDCTL mkdir /servicebroker/mycatshare
$ETCDCTL set /servicebroker/mycatshare/username $API_NAME
$ETCDCTL set /servicebroker/mycatshare/password $API_PASSWORD

$ETCDCTL mkdir /servicebroker/mycatshare/instance

$ETCDCTL mkdir /servicebroker/mycatshare/catalog




###创建服务 GuangFenCluster 广分集群
$ETCDCTL mkdir /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C #服务id
###创建服务级的配置
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/name "GuangFenCluster"
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/description "广分集群"
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/bindable true
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/planupdatable false
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/tags 'MySQL,ServiceBroker'
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/metadata '{"displayName":"GuangFenCluster","longDescription":"Database middleware, highly available sub-database and sub-table.","providerDisplayName":"Asiainfo"}'
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/selection "0"   #轮选标识

###创建套餐目录
$ETCDCTL mkdir /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/plan
###创建套餐
$ETCDCTL mkdir /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/plan/619A6421-386E-4684-BBB4-915E406D58F9
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/plan/619A6421-386E-4684-BBB4-915E406D58F9/name "standalone"
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/plan/619A6421-386E-4684-BBB4-915E406D58F9/description "单独实例"
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/plan/619A6421-386E-4684-BBB4-915E406D58F9/metadata '{"displayName":"Shared and Free"}'
$ETCDCTL set /servicebroker/mycatshare/catalog/EB1045DE-5FEA-4A60-9455-3727E1C0FD1C/plan/619A6421-386E-4684-BBB4-915E406D58F9/free true





###创建服务 MySQL Cluster2 西安集群
$ETCDCTL mkdir /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B #服务id
###创建服务级的配置
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/name "MySQL_Xian"
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/description "MySQL 西安集群"
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/bindable true
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/planupdatable false
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/tags 'MySQL,ServiceBroker'
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/metadata '{"displayName":"MySQL_Xian_Cluster","longDescription":"Database middleware, highly available sub-database and sub-table.","providerDisplayName":"Asiainfo"}'
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/selection "0"   #轮选标识

###创建套餐目录
$ETCDCTL mkdir /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/plan
###创建套餐
$ETCDCTL mkdir /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/plan/252817FA-CF50-4BA0-A1D8-74B8219F395F
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/plan/252817FA-CF50-4BA0-A1D8-74B8219F395F/name "standalone"
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/plan/252817FA-CF50-4BA0-A1D8-74B8219F395F/description "单独实例"
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/plan/252817FA-CF50-4BA0-A1D8-74B8219F395F/metadata '{"displayName":"Shared and Free"}'
$ETCDCTL set /servicebroker/mycatshare/catalog/5B69E4E7-1E1D-41B6-993E-10BEDE540D4B/plan/252817FA-CF50-4BA0-A1D8-74B8219F395F/free true