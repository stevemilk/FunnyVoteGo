package huawei

import (
	"encoding/json"
	"github.com/glog"
	"hyperbaas/src/api/vm"
	"hyperbaas/src/util"
)

type appInfoJSON struct {
	AppName       string `json:"app-name"`
	ZeusChainInfo string `json:"zeus-CHAININFO"`
	ZeusEIP       string `json:"zeus-EIP"`
	ZeusEPort     int    `json:"zeus-EPORT"`
	ZeusImage     string `json:"zeus-IMAGE"`
}

type appWithMysqlInfoJSON struct {
	AppName            string `json:"app-name"`
	ChainNodeInfo      string `json:"chain-NodeInfo"`
	ChainJSONRPCPort   string `json:"chain-JsonRpcPort"`
	ChainWebSocketPort string `json:"chain-WebSocketPort"`
	ZeusEIP            string `json:"zeus-EIP"`
	ZeusEPort          int    `json:"zeus-EPORT"`
	ZeusImage          string `json:"zeus-IMAGE"`
	MysqlDatabase      string `json:"mysql-database"`
	MysqlPassword      string `json:"mysql-password"`
	MysqlPort          int    `json:"mysql-port"`
	MysqlRootPassword  string `json:"mysql-root-password"`
	MysqlServiceName   string `json:"mysql-service-name"`
	MysqlUser          string `json:"mysql-user"`
}

type appStackInfo struct {
	Name             string               `json:"name"`
	TemplateID       string               `json:"template_id"`
	InputsJSON       appWithMysqlInfoJSON `json:"inputs_json"`
	Description      string               `json:"description"`
	Namespace        string               `json:"namespace"`
	ClusterID        string               `json:"cluster_id"`
	ActionParameters actionParameters     `json:"action_parameters"`
}

func deployApp(stackName, image, chainNodeInfo, jsonRPCPort, webSocketPort, database, ZeusEIP string, ZeusEPORT int, reqAppInfo *vm.ReqAppInfo) (string, error) {

	var stackID = ""
	var postFix = util.RandomNumStr()
	inputsJSON := appWithMysqlInfoJSON{
		AppName:            reqAppInfo.DAppName + "-" + postFix,
		ChainNodeInfo:      chainNodeInfo,
		ChainJSONRPCPort:   jsonRPCPort,
		ChainWebSocketPort: webSocketPort,
		ZeusEIP:            ZeusEIP,
		ZeusEPort:          ZeusEPORT,
		ZeusImage:          image,
		MysqlDatabase:      database,
		MysqlPassword:      "app@myql",
		MysqlPort:          3306,
		MysqlRootPassword:  "app@rootmyql",
		MysqlServiceName:   "mysql-service-" + postFix,
		MysqlUser:          "zeus",
	}
	description := "crate " + reqAppInfo.DAppName + " app"
	actionParameters := actionParameters{true}
	stackInfo := appStackInfo{
		Name:             stackName,
		TemplateID:       DappTemplateID,
		InputsJSON:       inputsJSON,
		Description:      description,
		Namespace:        "default",
		ClusterID:        DappClusterID,
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

func getAppOutput(stackid string) (string, error) {

	url := "https://aos.cn-north-1.myhuaweicloud.com/v2/stacks/" + stackid + "/outputs"

	var err error

	respmap, err := doHTTP("GET", url, "")

	if err != nil {
		return "", err
	}

	var stackOutput stackOutput

	if err = json.Unmarshal([]byte(respmap["body"].(string)), &stackOutput); err == nil {

		return stackOutput.Outputs["zeus-addr"].Value, err

	}

	return "", err

}
