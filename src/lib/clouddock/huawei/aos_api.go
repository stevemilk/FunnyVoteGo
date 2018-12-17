package huawei

import (
	"encoding/json"
	"hyperbaas/src/api/vm"
	"hyperbaas/src/util"
	"strconv"
	"strings"

	"time"

	"fmt"

	"github.com/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type relationship struct {
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	TargetID      string                 `json:"target_id"`
	TypeHierarchy []string               `json:"type_hierarchy"`
	Properties    map[string]interface{} `json:"properties"`
}

type actionStatus struct {
	Phase    string `json:"phase"`
	Message  string `json:"message"`
	Reason   string `json:"reason"`
	UpdateAt string `json:"update_at"`
}

type stackElements struct {
	ID                string                 `json:"id"`
	Description       string                 `json:"description"`
	Properties        map[string]interface{} `json:"properties"`
	RuntimeProperties map[string]interface{} `json:"runtime_properties"`
	Relationships     []relationship         `json:"relationships"`
	StackID           string                 `json:"stack_id"`
	Type              string                 `json:"type"`
	CreateAt          string                 `json:"create_at"`
	UpdateAt          string                 `json:"update_at"`
	TypeHierarchy     []string               `json:"type_hierarchy"`
	ActionStatus      actionStatus           `json:"action_status"`
}

type inputsJSON struct {
	AZ               string `json:"az"`
	ConsNodeCount    int    `json:"cons-nodecount"`
	ConsensusImage   string `json:"consensus-image"`
	ConsensusVMName  string `json:"consensus-vm-name"`
	DiskSize         int    `json:"disk_size"`
	BandWidth        int    `json:"band_width"`
	Flavor           string `json:"flavor"`
	NonconsNodeCount int    `json:"noncons-nodecount"`
	NonconsVMName    string `json:"noncons-vm-name"`
	Password         string `json:"passwd"`
	SubnetCidr       string `json:"subnet-cidr"`
	SubnetGateway    string `json:"subnet-gateway"`
	VpcCidr          string `json:"vpc-cidr"`
}

type actionParameters struct {
	AutoCreate bool `json:"auto_create"`
}

type stackInfo struct {
	Name             string           `json:"name"`
	TemplateID       string           `json:"template_id"`
	InputsJSON       inputsJSON       `json:"inputs_json"`
	Description      string           `json:"description"`
	ActionParameters actionParameters `json:"action_parameters"`
}

type stackResponse struct {
	Name            string `json:"name"`
	GUID            string `json:"guid"`
	Description     string `json:"description"`
	ProjectID       string `json:"project_id"`
	DomainID        string `json:"domain_id"`
	TemplateID      string `json:"template_id"`
	TemplateName    string `json:"template_name"`
	InputsJSON      string `json:"inputs_json"`
	Status          string `json:"status"`
	CreateAt        string `json:"create_at"`
	UpdateAt        string `json:"update_at"`
	Force           bool   `json:"force"`
	Labels          string `json:"labels"`
	ClusterID       string `json:"cluster_id"`
	ClusterName     string `json:"cluster_name"`
	Namespace       string `json:"namespace"`
	TemplateVersion string `json:"template_version"`
	DslVersion      string `json:"dsl_version"`
}

type outputParameter struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}

type stackOutput struct {
	Outputs map[string]outputParameter `json:"outputs"`
}

type deleteStackOutput struct {
	ActionID     string `json:"action_id"`
	LastActionID string `json:"last_action_id"`
}

//UpgradeStackOutput upgrade stack output
type UpgradeStackOutput struct {
	ActionID     string `json:"action_id"`
	LastActionID string `json:"last_action_id"`
}

type aosErrorMessage struct {
	Message    string `json:"message"`
	Code       string `json:"code"`
	Extend     string `json:"extend"`
	ShowDetail bool   `json:"showdetail"`
}

type execution struct {
	Kind       string                 `json:"kind"`
	APIVersion string                 `json:"apiVersion"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
	Status     executionStatus        `json:"status"`
}

type executionStatus struct {
	ActionNAme      string                 `json:"actionName"`
	Progress        int                    `json:"progress"`
	ObjectStatus    objectStatus           `json:"objectStatus"`
	SubObjectStatus map[string]interface{} `json:"subObjectStatus"`
}

type objectStatus struct {
	Phase    string `json:"phase"`
	Message  string `json:"message"`
	Reason   string `json:"reason"`
	UpdateAt string `json:"update_at"`
}

type inputs struct {
	ConsNodeCount int `json:"cons-nodecount"`
}

// upgradeInputJson upgrade, add new node input json
type upgradeInputJSON struct {
	Inputs    inputs `json:"inputs"`
	LifeCycle string `json:"lifecycle"`
}

type containerInfoJSON struct {
	AppName       string `json:"app-name"`
	ZeusEIP       string `json:"zeus-EIP"`
	DaemonEPort   int    `json:"zeus-daemonPort"`
	Jsonrpc1EPort int    `json:"zeus-jsonRpc1Port"`
	Jsonrpc2EPort int    `json:"zeus-jsonRpc2Port"`
	Jsonrpc3EPort int    `json:"zeus-jsonRpc3Port"`
	Jsonrpc4EPort int    `json:"zeus-jsonRpc4Port"`
	SSHPort       int    `json:"zeus-sshPort"`
	ZeusImage     string `json:"zeus-IMAGE"`
}

type containerStackInfo struct {
	Name             string            `json:"name"`
	TemplateID       string            `json:"template_id"`
	InputsJSON       containerInfoJSON `json:"inputs_json"`
	Description      string            `json:"description"`
	Namespace        string            `json:"namespace"`
	ClusterID        string            `json:"cluster_id"`
	ActionParameters actionParameters  `json:"action_parameters"`
}

// UpgradeStack upgrade stack, add new VP node into chain
func UpgradeStack(stackID string, ConsNodeCount int) (*UpgradeStackOutput, error) {
	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackID + "/actions"
	upgradeInputJSON := upgradeInputJSON{
		inputs{ConsNodeCount: ConsNodeCount},
		"upgrade",
	}
	jsonByte, err := json.Marshal(upgradeInputJSON)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(jsonByte))
	resMap, err := doHTTP("PUT", url, string(jsonByte))
	if err != nil {
		return nil, err
	}

	var upgradeStackOutput UpgradeStackOutput
	err = json.Unmarshal([]byte(resMap["body"].(string)), &upgradeStackOutput)
	if err != nil {
		return nil, err
	}
	return &upgradeStackOutput, nil

}

// GetStackOutput get after upgrade servers info
func GetStackOutput(conscount int, stackid string) ([]ServerInfo, error) {
	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackid + "/outputs"
	var consServerInfo = make([]ServerInfo, conscount)
	resMap, err := doHTTP("GET", url, "")
	if err != nil {
		return consServerInfo, err
	}
	var stackOutput stackOutput
	err = json.Unmarshal([]byte(resMap["body"].(string)), &stackOutput)
	if err != nil {
		return nil, err
	}
	// fill consensus nodes server info
	consIPArray := stringToArray(stackOutput.Outputs["consensus-node-eip"].Value, conscount)
	consIDArray := stringToArray(stackOutput.Outputs["consensus-node-eid"].Value, conscount)
	consNameArray := stringToArray(stackOutput.Outputs["consensus-node-ename"].Value, conscount)
	for i := range consServerInfo {
		consServerInfo[i].IP = consIPArray[i]
		consServerInfo[i].ID = consIDArray[i]
		consServerInfo[i].ServerName = consNameArray[i]
		consServerInfo[i].ServerType = 1
		consServerInfo[i].UserName = "root"
		consServerInfo[i].Password = viper.GetString("huaweicloud.ecs_password")
	}
	return consServerInfo, err
}

func createStack(image string, flavor string, reqBlockChain vm.ReqBolckChain, reqServer vm.ReqServer, chainCode string) (string, error) {

	var stackID = ""
	name := chainCode
	templateID := viper.GetString("huaweicloud.template_id") //"9bfed23a-76a9-345a-d672-dd5142e869a2"
	inputsJSON := inputsJSON{
		"cn-north-1a",
		reqBlockChain.ConsensusNodeNumber,
		image,
		"consensus-server",
		reqServer.DiskSize,
		reqServer.BandWidth,
		flavor,
		0,
		"nonconsensus-server",
		viper.GetString("huaweicloud.ecs_password"),
		"192.168.1.0/24",
		"192.168.1.1",
		"192.168.0.0/16",
	}
	description := "create " + reqBlockChain.Name + " servers"
	actionParameters := actionParameters{true}
	stackInfo := stackInfo{
		name,
		templateID,
		inputsJSON,
		description,
		actionParameters,
	}
	jsonstr, err := json.Marshal(stackInfo)
	if err != nil {
		return stackID, err
	}
	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks"
	respmap, err := doHTTP("POST", url, string(jsonstr))
	if err != nil {
		return stackID, err
	}
	var stackResponse stackResponse
	if err := json.Unmarshal([]byte(respmap["body"].(string)), &stackResponse); err == nil {
		stackID = stackResponse.GUID
	} else {
		glog.Error("get stack response error:", err)
	}
	glog.Info("stackID:", stackID)

	return stackID, err
}

func getStackOutput(conscount int, nonconscount int, stackid string) ([]ServerInfo, []ServerInfo, error) {

	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackid + "/outputs"
	var consServerInfo = make([]ServerInfo, conscount)
	var nonconsServerInfo = make([]ServerInfo, nonconscount)
	var err error
	respmap, err := doHTTP("GET", url, "")
	if err != nil {
		return consServerInfo, nonconsServerInfo, err
	}
	var stackOutput stackOutput
	if err = json.Unmarshal([]byte(respmap["body"].(string)), &stackOutput); err == nil {
		// fill consensus nodes server info
		consIPArray := stringToArray(stackOutput.Outputs["consensus-node-eip"].Value, conscount)
		consIDArray := stringToArray(stackOutput.Outputs["consensus-node-eid"].Value, conscount)
		consNameArray := stringToArray(stackOutput.Outputs["consensus-node-ename"].Value, conscount)
		securityGroupID := stackOutput.Outputs["security-group-id"].Value
		for i := range consServerInfo {
			consServerInfo[i].IP = consIPArray[i]
			consServerInfo[i].ID = consIDArray[i]
			consServerInfo[i].ServerName = consNameArray[i]
			consServerInfo[i].ServerType = 1
			consServerInfo[i].UserName = "root"
			consServerInfo[i].Password = viper.GetString("huaweicloud.ecs_password")
			consServerInfo[i].SecurityGroupID = securityGroupID
		}
		// fill nonconsensus nodes server info
		nonconsIPArray := stringToArray(stackOutput.Outputs["consensus-node-eip"].Value, nonconscount)
		nonconsIDArray := stringToArray(stackOutput.Outputs["consensus-node-eid"].Value, nonconscount)
		nonconsNameArray := stringToArray(stackOutput.Outputs["consensus-node-ename"].Value, nonconscount)
		for i := range nonconsServerInfo {
			nonconsServerInfo[i].IP = nonconsIPArray[i]
			nonconsServerInfo[i].ID = nonconsIDArray[i]
			nonconsServerInfo[i].ServerName = nonconsNameArray[i]
			nonconsServerInfo[i].ServerType = 2
			nonconsServerInfo[i].UserName = "root"
			nonconsServerInfo[i].Password = viper.GetString("huaweicloud.ecs_password")
			nonconsServerInfo[i].SecurityGroupID = securityGroupID
		}
	} else {
		return consServerInfo, nonconsServerInfo, err
	}

	return consServerInfo, nonconsServerInfo, err
}

func createContainerStack(ports []int, image string, reqBlockChain vm.ReqBolckChain, reqServer vm.ReqServer, chainCode string) (string, error) {
	var stackID = ""
	var postFix = util.RandomNumStr()
	inputsJSON := containerInfoJSON{
		AppName:       reqBlockChain.Name + "-" + postFix,
		ZeusImage:     image,
		ZeusEIP:       viper.GetString("huaweicloud.elb_ip"),
		DaemonEPort:   ports[0],
		Jsonrpc1EPort: ports[1],
		Jsonrpc2EPort: ports[2],
		Jsonrpc3EPort: ports[3],
		Jsonrpc4EPort: ports[4],
		SSHPort:       ports[5],
	}
	description := "crate " + reqBlockChain.Name + " container"
	actionParameters := actionParameters{true}
	stackInfo := containerStackInfo{
		Name:             chainCode,
		TemplateID:       viper.GetString("huaweicloud.contanier_template_id"),
		InputsJSON:       inputsJSON,
		Description:      description,
		Namespace:        "default",
		ClusterID:        viper.GetString("huaweicloud.cluster_id"),
		ActionParameters: actionParameters,
	}
	jsonstr, err := json.Marshal(stackInfo)
	if err != nil {
		return stackID, err
	}
	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks"
	respmap, err := doHTTP("POST", url, string(jsonstr))
	if err != nil {
		return stackID, err
	}
	var stackResponse stackResponse
	if err := json.Unmarshal([]byte(respmap["body"].(string)), &stackResponse); err == nil {
		stackID = stackResponse.GUID
	} else {
		glog.Error("get stack response error:", err)
	}
	glog.Info("stackID:", stackID)

	return stackID, err
}

func getContainerStackOutput(stackid string) ([]ServerInfo, error) {

	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackid + "/outputs"

	var err error

	var serverInfo = make([]ServerInfo, 1)

	respmap, err := doHTTP("GET", url, "")

	if err != nil {
		return serverInfo, err
	}

	var stackOutput stackOutput

	if err = json.Unmarshal([]byte(respmap["body"].(string)), &stackOutput); err == nil {

		glog.Info("stackOutput: ", stackOutput.Outputs)
		serverInfo[0].IP = viper.GetString("huaweicloud.elb_ip")
		serverInfo[0].ServerName = stackOutput.Outputs["zeusName"].Value
		serverInfo[0].UserName = "root"
		serverInfo[0].Password = viper.GetString("huaweicloud.ecs_password")
		portstr := stackOutput.Outputs["ports"].Value
		glog.Info("portstr: ", portstr)
		portstr = portstr[1 : len(portstr)-1]
		array := strings.Split(portstr, ",")
		ports := make([]int, len(array))
		for i := range ports {
			ports[i], _ = strconv.Atoi(array[i])
		}
		serverInfo[0].Ports = ports

		return serverInfo, err

	}

	return serverInfo, err

}

func stringToArray(str string, count int) []string {

	var array = make([]string, count)

	if str != "" {
		str = strings.Replace(str, "\\\"", "", -1)
		str = str[1 : len(str)-1]
		array = strings.Split(str, ",")
	}

	return array

}

// QueryStatus return stack status
func QueryStatus(stackID string, jobType string) (string, error) {

	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackID + "/elements"
	var status string
	var err error
	respmap, err := doHTTP("GET", url, "")
	var stackElemetns []stackElements
	if err != nil {
		glog.Errorf("http request error: %v", err)
		return status, err
	}
	err = json.Unmarshal([]byte(respmap["body"].(string)), &stackElemetns)

	if err != nil {
		glog.Errorf("json unmarshal error: %v", err)
		return status, err
	}

	if jobType == "ecs_server" {
		// need to check consensus and unconsensus server status
		var consensstatus, unconsensstatus string
		for i := range stackElemetns {
			if stackElemetns[i].ID == "consensus-node" {
				consensstatus = stackElemetns[i].ActionStatus.Phase
			} else if stackElemetns[i].ID == "noncons" {
				unconsensstatus = stackElemetns[i].ActionStatus.Phase
			} else {
				continue
			}
		}
		if consensstatus == "Succeeded" {
			if unconsensstatus == "" || unconsensstatus == "Succeeded" {
				status = "Succeeded"
			}
		}
	} else if jobType == "container" {
		allJobStatus := true
		for i := range stackElemetns {
			status = stackElemetns[i].ActionStatus.Phase
			if status != "Succeeded" {
				allJobStatus = false
				break
			}
		}
		if allJobStatus {
			status = "Succeeded"
		}
	} else {
		allJobStatus := true
		for i := range stackElemetns {
			if stackElemetns[i].ID == jobType {
				status = stackElemetns[i].ActionStatus.Phase
				if status != "Succeeded" {
					allJobStatus = false
					break
				}
			}
		}
		if allJobStatus {
			status = "Succeeded"
		}
	}

	return status, err
}

// DeleteStack delete stack in huawei cloud
func DeleteStack(stackID string) error {

	glog.Infof("begin to query delete stack request, stackID is :%v", stackID)

	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackID
	respmap, err := doHTTP("DELETE", url, "")
	if err == nil && respmap["status"] == "200 OK" {
		var deleteStackOutput deleteStackOutput
		if err = json.Unmarshal([]byte(respmap["body"].(string)), &deleteStackOutput); err == nil {
			actionID := deleteStackOutput.ActionID
			var phase string
			for i := 0; i < 30; i++ {
				phase = GetExecution(stackID, actionID)
				if phase == "Deleted" {
					break
				}
				time.Sleep(time.Duration(10) * time.Second)
			}
			if phase == "Deleted" {
				glog.Infof("delete stack success, stackID is :%v", stackID)
				return nil
			}
			glog.Infof("delete stack success, stackID is :%v", stackID)
		} else {
			glog.Errorf("delete stack failed, err is :%v", err)
			util.PostErrorf("delete stack failed, err is :%v", err)
		}
	}
	return errors.New("delete stack failed")
}

// GetExecution return delete stack status
func GetExecution(stackID string, actionID string) (phase string) {

	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackID + "/actions/" + actionID
	respmap, err := doHTTP("GET", url, "")
	if err == nil {
		if respmap["status"] == "200 OK" {
			var execution execution
			if err = json.Unmarshal([]byte(respmap["body"].(string)), &execution); err == nil {
				phase = execution.Status.ObjectStatus.Phase
			} else {
				glog.Error("GetExecution error:", err)
			}
		} else if respmap["status"] == "404 Not Found" {
			phase = "Deleted"
		} else {
			phase = "Failed"
		}
	} else {
		phase = "Failed"
	}
	return
}
