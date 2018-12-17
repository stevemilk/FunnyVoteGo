package huawei

import (
	"encoding/json"
	"fmt"
	"github.com/glog"
	"github.com/spf13/viper"
)

type ruleReqInfo struct {
	SecurityGroupID string `json:"security_group_id"`
	Description     string `json:"description"`
	Direction       string `json:"direction"`
	Ethertype       string `json:"ethertype"`
	Protocol        string `json:"protocol"`
	PortRangeMin    int    `json:"port_range_min"`
	PortRangeMax    int    `json:"port_range_max"`
	//RemoteIPPrefix  string `json:"remote_ip_prefix"`
	//RemoteGroupID   string `json:"remote_group_id"`
}

type ruleReqJSON struct {
	SecurityGroupRule ruleReqInfo `json:"security_group_rule"`
}

type ruleResInfo struct {
	ID              string `json:"id"`
	SecurityGroupID string `json:"security_group_id"`
	Description     string `json:"description"`
	Direction       string `json:"direction"`
	Ethertype       string `json:"ethertype"`
	Protocol        string `json:"protocol"`
	PortRangeMin    int    `json:"port_range_min"`
	PortRangeMax    int    `json:"port_range_max"`
	RemoteIPPrefix  string `json:"remote_ip_prefix"`
	RemoteGroupID   string `json:"remote_group_id"`
}

type ruleResJSON struct {
	SecurityGroupRule ruleResInfo `json:"security_group_rule"`
}

// CreateRule create security group rule in huawei cloud
func CreateRule(portMin, portMax int, groupID, direction string) (err error) {

	securityGroupInfo := ruleReqInfo{
		SecurityGroupID: groupID,
		Description:     "",
		Direction:       direction,
		Ethertype:       "IPv4",
		Protocol:        "TCP",
		PortRangeMin:    portMin,
		PortRangeMax:    portMax,
		//RemoteIPPrefix:  "",
		//RemoteGroupID:   "",
	}

	ruleInputJSON := ruleReqJSON{
		SecurityGroupRule: securityGroupInfo,
	}

	fmt.Print("createRule security group id: " + securityGroupInfo.SecurityGroupID)

	jsonstr, err := json.Marshal(ruleInputJSON)

	if err != nil {
		return
	}

	url := "https://vpc.cn-north-1.myhuaweicloud.com/v1/" + viper.GetString("huaweicloud.project_id") + "/security-group-rules"

	fmt.Print("url: " + url)

	respmap, err := doHTTP("POST", url, string(jsonstr))
	if err != nil {
		return err
	}

	var ruleResonse ruleResJSON

	if err := json.Unmarshal([]byte(respmap["body"].(string)), &ruleResonse); err == nil {
		glog.Info("get rule response success, rule ID :", ruleResonse.SecurityGroupRule.ID)
	} else {
		glog.Error("get rule response error:", err)
	}

	return
}

// DeleteRule delete security group rule in huawei cloud
func DeleteRule(ruleID string) (err error) {

	url := "https://vpc.cn-north-1.myhuaweicloud.com/v1/" + viper.GetString("huaweicloud.project_id") + "/security-group-rules/" + ruleID

	respmap, err := doHTTP("DELETE", url, "")

	if err == nil && respmap["status"] == "200 OK" {
		glog.Info("delete security group rule success")
	} else {
		glog.Error("delete security group rule error", err)
	}

	return
}
