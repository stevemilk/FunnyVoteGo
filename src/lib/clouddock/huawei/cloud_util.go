package huawei

import (
	"hyperbaas/src/api/vm"
	"hyperbaas/src/util"
	"time"

	"hyperbaas/src/constant"
	"hyperbaas/src/state"

	"github.com/glog"
	"github.com/pkg/errors"
)

// ServerInfo def basic server info model
type ServerInfo struct {
	ID string `json:"id"`

	IP string `json:"ip"`

	ServerName string `json:"server_name"`

	ServerType int `json:"server_type"` //1表示共识节点，2表示非共识节点

	UserName string `json:"user_name"`

	Password string `json:"password"`

	SecurityGroupID string `json:"security_group_id"`

	Ports []int `json:"ports"`
}

// JobMap store the deploy job status
var JobMap = make(map[string][]vm.JobDetail)

// LastJob store last job of user
var LastJob = make(map[uint]string)

// CreateHyperChainServers apply cloud servers for hyperchain
func CreateHyperChainServers(image string, flavor string, reqBlockChain vm.ReqBolckChain, reqServer vm.ReqServer, chainCode string, creatorID uint) (string, []ServerInfo, error) {

	var stackID = ""

	var err error

	var consServerInfo []ServerInfo

	var nonconsServerInfo []ServerInfo

	stackID, err = createStack(image, flavor, reqBlockChain, reqServer, chainCode)

	if err != nil || stackID == "" {
		glog.Errorf("create stack error: %v", err)
		glog.Errorf("create stack error, stackID: %v", stackID)
		util.PostErrorf("create stack error, stackID = %v, error info = %v ", stackID, err)
		return stackID, consServerInfo, errors.New("create stack error")
	}

	glog.Infof("stackID: %v", stackID)

	// init create vpc job status
	UpdateJobMap(chainCode, 0, vm.JobDetail{
		ID:           stackID,
		JobStatus:    1,
		JobStartTime: time.Now(),
	})
	state.GlobalWsClients.SendMessageByUserID(creatorID, constant.StateDeployVCloud)

	vpcStatus, err := QueryStatus(stackID, "myvpc")
	for vpcStatus != "Succeeded" {
		time.Sleep(time.Duration(5) * time.Second)
		vpcStatus, err = QueryStatus(stackID, "myvpc")
		glog.Infof("vpc status: %v", vpcStatus)
		if err != nil || vpcStatus == "Failed" {
			UpdateJobMap(chainCode, 0, vm.JobDetail{
				JobStatus:  3,
				JobEndTime: time.Now(),
			})
			glog.Errorf("query vpc status error: %v", err)
			return stackID, consServerInfo, err
		}
	}
	UpdateJobMap(chainCode, 0, vm.JobDetail{
		JobStatus:   2,
		JobEndTime:  time.Now(),
		JobProgress: 10,
	})

	// init create subnet job status
	UpdateJobMap(chainCode, 1, vm.JobDetail{
		ID:           stackID,
		JobStatus:    1,
		JobStartTime: time.Now(),
	})
	state.GlobalWsClients.SendMessageByUserID(creatorID, constant.StateDeployVnet)

	subnetStatus, err := QueryStatus(stackID, "mysubnet")
	for subnetStatus != "Succeeded" {
		time.Sleep(time.Duration(5) * time.Second)
		subnetStatus, err = QueryStatus(stackID, "mysubnet")
		glog.Infof("subnet status: %v", vpcStatus)
		if err != nil || subnetStatus == "Failed" {
			UpdateJobMap(chainCode, 1, vm.JobDetail{
				JobStatus:  3,
				JobEndTime: time.Now(),
			})
			glog.Errorf("query subnet status error: %v", err)
			return stackID, consServerInfo, err
		}
	}
	UpdateJobMap(chainCode, 1, vm.JobDetail{
		JobStatus:   2,
		JobEndTime:  time.Now(),
		JobProgress: 20,
	})

	// init create ecs job status
	UpdateJobMap(chainCode, 2, vm.JobDetail{
		ID:           stackID,
		JobStatus:    1,
		JobStartTime: time.Now(),
	})
	state.GlobalWsClients.SendMessageByUserID(creatorID, constant.StateDeployCreateVPS)

	ecsStatus, err := QueryStatus(stackID, "ecs_server")
	var count = 0
	for ecsStatus != "Succeeded" && count < 40 {
		time.Sleep(time.Duration(15) * time.Second)
		ecsStatus, err = QueryStatus(stackID, "ecs_server")
		glog.Infof("ecs server status: %v", ecsStatus)
		if err != nil {
			glog.Errorf("query stack status error: %v", err)
		}
		count++
	}
	if count == 40 {
		UpdateJobMap(chainCode, 2, vm.JobDetail{
			JobStatus:  3,
			JobEndTime: time.Now(),
		})
		glog.Error("query stack status time out")
		return stackID, consServerInfo, errors.New("query stack status time out")
	}
	UpdateJobMap(chainCode, 2, vm.JobDetail{
		JobStatus:   2,
		JobEndTime:  time.Now(),
		JobProgress: 90,
	})

	consServerInfo, nonconsServerInfo, err = getStackOutput(reqBlockChain.ConsensusNodeNumber, reqBlockChain.UnconsensusNodeNumber, stackID)

	if err != nil {
		glog.Error("get stack output error:", err)
		return stackID, consServerInfo, err
	}

	var servers = make([]ServerInfo, reqBlockChain.ConsensusNodeNumber+reqBlockChain.UnconsensusNodeNumber)

	for i := range servers {
		if i < reqBlockChain.ConsensusNodeNumber {
			servers[i] = consServerInfo[i]
		} else {
			servers[i] = nonconsServerInfo[i-reqBlockChain.ConsensusNodeNumber]
		}
	}

	return stackID, servers, err

}

// CreateHyperChainContainers apply cloud containers for hyperchain
func CreateHyperChainContainers(ports []int, image string, reqBlockChain vm.ReqBolckChain, reqServer vm.ReqServer, chainCode string, creatorID uint) (string, []ServerInfo, error) {

	var stackID = ""

	var err error

	var consServerInfo []ServerInfo

	// init create ecs job status
	UpdateJobMap(chainCode, 2, vm.JobDetail{
		ID:           stackID,
		JobStatus:    1,
		JobStartTime: time.Now(),
	})
	state.GlobalWsClients.SendMessageByUserID(creatorID, constant.StateDeployCreateVPS)

	//var ZeusEIP string
	//ips := GetAllClusterNodes(viper.GetString("huaweicloud.project_id"), viper.GetString("huaweicloud.cluster_id"))
	//ZeusEIP = ips[0]

	stackID, err = createContainerStack(ports, image, reqBlockChain, reqServer, chainCode)

	if err != nil || stackID == "" {
		glog.Errorf("create container stack error: %v", err)
		glog.Errorf("create container stack error, stackID: %v", stackID)
		return stackID, consServerInfo, errors.New("create container stack error")
	}

	glog.Infof("container stackID: %v", stackID)

	status, err := QueryStatus(stackID, "container")

	glog.Infof("container status: %v", status)

	var count = 0

	for status != "Succeeded" && count < 12 {

		time.Sleep(time.Duration(10) * time.Second)

		status, err = QueryStatus(stackID, "container")

		glog.Infof("container stack status: %v", status)

		if err != nil {
			glog.Errorf("query container stack staus error: %v", err)
			return stackID, consServerInfo, err
		}

		count++

	}

	if count == 12 {
		glog.Error("query app stack status time out")
		UpdateJobMap(chainCode, 2, vm.JobDetail{
			JobStatus:  3,
			JobEndTime: time.Now(),
		})
		return stackID, consServerInfo, errors.New("query stack status time out")
	}

	UpdateJobMap(chainCode, 2, vm.JobDetail{
		JobStatus:   2,
		JobEndTime:  time.Now(),
		JobProgress: 90,
	})

	consServerInfo, err = getContainerStackOutput(stackID)

	return stackID, consServerInfo, err

}

// CreateApp create app in the huawei cloud
func CreateApp(stackName, image, chainNodeInfo, jsonRPCPort, webSocketPort, database, ZeusEIP string, ZeusEPORT int, reqAppInfo *vm.ReqAppInfo) (string, error) {

	var stackID = ""

	var err error

	stackID, err = deployApp(stackName, image, chainNodeInfo, jsonRPCPort, webSocketPort, database, ZeusEIP, ZeusEPORT, reqAppInfo)

	if err != nil || stackID == "" {
		glog.Errorf("create app stack error: %v", err)
		glog.Errorf("create app stack error, stackID: %v", stackID)
		return stackID, errors.New("create app stack error")
	}

	glog.Infof("app stackID: %v", stackID)

	status, err := QueryStatus(stackID, "app")

	var count = 0

	for status != "Succeeded" && count < 12 {

		time.Sleep(time.Duration(10) * time.Second)

		status, err = QueryStatus(stackID, "app")

		glog.Infof("app stack status: %v", status)

		if err != nil {
			glog.Errorf("query app stack staus error: %v", err)
			return stackID, err
		}

		count++

	}

	if count == 12 {
		glog.Error("query app stack status time out")
		return stackID, errors.New("query stack status time out")
	}

	visitURL, err := getAppOutput(stackID)

	glog.Infof("visitUrl:%v", visitURL)

	return stackID, err
}

// UpdateJobMap update jobMap
func UpdateJobMap(chainCode string, index int, detail vm.JobDetail) {

	if detail.ID != "" {
		JobMap[chainCode][index].ID = detail.ID
	}

	if detail.ChainID != 0 {
		for i := 0; i < 5; i++ {
			JobMap[chainCode][i].ChainID = detail.ChainID
		}
	}

	if detail.JobStatus != 0 {
		JobMap[chainCode][index].JobStatus = detail.JobStatus
	}

	if !detail.JobStartTime.IsZero() {
		JobMap[chainCode][index].JobStartTime = detail.JobStartTime
	}

	if !detail.JobEndTime.IsZero() {
		JobMap[chainCode][index].JobEndTime = detail.JobEndTime
		JobMap[chainCode][index].JobTime = int(JobMap[chainCode][index].JobEndTime.Sub(JobMap[chainCode][index].JobStartTime).Seconds())
	}

	if detail.JobProgress != 0 {
		JobMap[chainCode][index].JobProgress = detail.JobProgress
	}

}
