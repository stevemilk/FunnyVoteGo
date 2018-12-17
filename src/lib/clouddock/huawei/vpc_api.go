package huawei

import (
	"encoding/json"
	"hyperbaas/src/api/vm"
	"time"

	"github.com/glog"
)

type vpcReqInfo struct {
	Name string `json:"name"`
	Cidr string `json:"cidr"`
}

type vpcReqJSON struct {
	Vpc vpcReqInfo `json:"vpc"`
}

type vpcResInfo struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Cidr   string   `json:"cidr"`
	Status string   `json:"status"`
	Routes []string `json:"routes"`
}

// VpcResJSON define vpc response json model
type VpcResJSON struct {
	Vpc vpcResInfo `json:"vpc"`
}

type subnetReq struct {
	Name      string `json:"name"`
	Cidr      string `json:"cidr"`
	Gatewayip string `json:"gateway_ip"`
	Vpcid     string `json:"vpc_id"`
}

type subnetReqJSON struct {
	Subnet subnetReq `json:"subnet"`
}

type subnetRes struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Cidr             string   `json:"cidr"`
	DNSList          []string `json:"dnsList"`
	Status           string   `json:"status"`
	Vpcid            string   `json:"vpc_id"`
	Gatewayip        string   `json:"gateway_ip"`
	DhcpEnable       bool     `json:"dhcp_enable"`
	NeutronNetworkID string   `json:"neutron_network_id"`
	NeutronSubnetID  string   `json:"neutron_subnet_id"`
}

// SubnetResJSON define subnet response json model
type SubnetResJSON struct {
	Subnet subnetRes `json:"subnet"`
}

func createVPC(name string, jobID string) string {

	//jsonstr :=
	//	`{
	//		"vpc":{
	//			"name": "testvpc",
	//			"cidr": "192.168.0.0/16"
	//		}
	//	}`

	var vpcid = ""

	vpc := vpcReqInfo{name + "Vpc", "192.168.0.0/16"}

	jsonvpc := vpcReqJSON{vpc}

	jsonstr, _ := json.Marshal(jsonvpc)

	url := "https://vpc.cn-north-1.myhuaweicloud.com/v1/ca7f44bee05d492ba9dfe0d67d31e383/vpcs"

	respmap, err := doHTTP("POST", url, string(jsonstr))

	if err != nil {

		panic(err)

	} else {

		var vpcRes VpcResJSON

		if err := json.Unmarshal([]byte(respmap["body"].(string)), &vpcRes); err == nil {

			vpcid = vpcRes.Vpc.ID
		}

	}

	//update jobMap
	jobDetail := vm.JobDetail{vpcid, 0, 1, "创建虚拟私有云", 1, time.Now(), time.Now(), 0, 0, 1}
	JobMap[jobID][0] = jobDetail

	return vpcid

}

// QueryVPC return vpc create status
func QueryVPC(vpcid string) VpcResJSON {

	var vpcRes VpcResJSON

	url := "https://vpc.cn-north-1.myhuaweicloud.com/v1/ca7f44bee05d492ba9dfe0d67d31e383/vpcs/" + vpcid

	respmap, err := doHTTP("GET", url, "")

	if err != nil {

		panic(err)

	} else {

		glog.Info(respmap["body"].(string))

		if err := json.Unmarshal([]byte(respmap["body"].(string)), &vpcRes); err != nil {

			panic(err)

		}

	}

	return vpcRes

}

func createSubnet(vpcid string, jobID string) string {

	//jsonstr :=
	//
	//`{
	//	"subnet":
	//	{
	//		"name": "test_subnet",
	//		"cidr": "192.168.20.0/24",
	//		"gateway_ip": "192.168.20.1",
	//		"vpc_id":"ec946396-ac3a-4fd9-8cdf-1415ca427314"
	//	}
	//}`

	var subnetID = ""

	subnet := subnetReq{"subnet", "192.168.20.0/24", "192.168.20.1", vpcid}

	jsonsubnet := subnetReqJSON{subnet}

	jsonstr, _ := json.Marshal(jsonsubnet)

	url := "https://vpc.cn-north-1.myhuaweicloud.com/v1/ca7f44bee05d492ba9dfe0d67d31e383/subnets"

	respmap, err := doHTTP("POST", url, string(jsonstr))

	if err != nil {

		panic(err)

	} else {

		var subnetRes SubnetResJSON

		if err := json.Unmarshal([]byte(respmap["body"].(string)), &subnetRes); err == nil {

			subnetID = subnetRes.Subnet.ID
		}

	}

	//update jobMap
	jobDetail := vm.JobDetail{subnetID, 0, 2, "创建虚拟子网", 1, time.Now(), time.Now(), 0, 0, 1}
	JobMap[jobID][1] = jobDetail

	return subnetID

}

// QuerySubnet return subnet create status
func QuerySubnet(subnetID string) SubnetResJSON {

	var subnetRes SubnetResJSON

	url := "https://vpc.cn-north-1.myhuaweicloud.com/v1/ca7f44bee05d492ba9dfe0d67d31e383/subnets/" + subnetID

	respmap, err := doHTTP("GET", url, "")

	if err != nil {

		panic(err)

	} else {

		if err := json.Unmarshal([]byte(respmap["body"].(string)), &subnetRes); err != nil {

			panic(err)

		}

	}

	return subnetRes

}

func createVMNetwork(vpcName string, jobID string) (vpcid string, subnetID string) {

	vpcid = ""

	subnetID = ""

	vpcid = createVPC(vpcName, jobID)

	vpcstatus := "CREATING"

	subnetstatus := "UNKNOWN"

	for vpcstatus == "CREATING" {
		time.Sleep(time.Duration(10) * time.Second)
		vpcstatus = QueryVPC(vpcid).Vpc.Status
	}

	if vpcstatus == "OK" {

		subnetID = createSubnet(vpcid, jobID)

	}

	for subnetstatus == "UNKNOWN" {
		time.Sleep(time.Duration(10) * time.Second)
		subnetstatus = QuerySubnet(subnetID).Subnet.Status
	}

	if subnetstatus != "ACTIVE" {
		glog.Info("create vm network failed")
	}

	glog.Info("vpcid:", vpcid)
	glog.Info("vpcstatus", vpcstatus)
	glog.Info("subnetid", subnetID)
	glog.Info("subnetstatus", subnetstatus)

	return vpcid, subnetID

}
