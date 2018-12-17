package huawei

import (
	"encoding/json"
	"github.com/glog"
)

type clusterNodes struct {
	Kind       string      `json:"kind"`
	APIVersion string      `json:"apiVersion"`
	Items      []nodesItem `json:"items"`
}

type nodesItem struct {
	Kind       string                 `json:"kind"`
	APIVersion string                 `json:"apiVersion"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
	Status     nodesStatus            `json:"status"`
}

type nodesStatus struct {
	Phase     string `json:"phase"`
	ServerID  string `json:"serverId"`
	PublicIP  string `json:"publicIP"`
	PrivateIP string `json:"privateIP"`
}

// GetAllClusterNodes return cluster all nodes float ip
func GetAllClusterNodes(projectID, clusterID string) (ips []string) {

	url := "https://cce.cn-north-1.myhuaweicloud.com/api/v3/projects/" + projectID + "/clusters/" + clusterID + "/nodes"

	respmap, err := doHTTP("GET", url, "")

	var clusterNodes clusterNodes

	if err == nil {
		if err = json.Unmarshal([]byte(respmap["body"].(string)), &clusterNodes); err == nil {
			ips = make([]string, len(clusterNodes.Items))
			for i := range clusterNodes.Items {
				ips[i] = clusterNodes.Items[i].Status.PublicIP
			}
		} else {
			glog.Error("query cluster nodes error:", err)
		}
	}

	return
}
