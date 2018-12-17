package huawei

import (
	"encoding/json"
	"hyperbaas/src/api/vm"
	"time"

	"github.com/glog"
	"github.com/spf13/viper"
)

type bandwidthInfo struct {
	Size      int    `json:"size"`
	Sharetype string `json:"sharetype"`
}

type eipInfo struct {
	Iptype    string        `json:"iptype"`
	Bandwidth bandwidthInfo `json:"bandwidth"`
}

type publicipInfo struct {
	Eip eipInfo `json:"eip"`
}

type subnetInfo struct {
	SubnetID string `json:"subnet_id"`
}

type volumeInfo struct {
	Volumetype string `json:"volumetype"`
	Size       int    `json:"size"`
}

type securityGroup struct {
	ID string `json:"id"`
}

type serverInfo struct {
	AvailabilityZone string          `json:"availability_zone"`
	Name             string          `json:"name"`
	ImageRef         string          `json:"imageRef"`
	RootvolumeInfo   volumeInfo      `json:"root_volume"`
	DataVolumes      []volumeInfo    `json:"data_volumes"`
	FlavorRef        string          `json:"flavorRef"`
	Vpcid            string          `json:"vpcid"`
	SecurityGroups   []securityGroup `json:"security_groups"`
	Nics             []subnetInfo    `json:"nics"`
	Publicip         publicipInfo    `json:"publicip"`
	AdminPass        string          `json:"adminPass"`
	Count            int             `json:"count"`
}

type serverJSON struct {
	Server serverInfo `json:"server"`
}

type floatingIP struct {
	ID         string `json:"id"`
	Pool       string `json:"pool"`
	IP         string `json:"ip"`
	FixedIP    string `json:"fixed_ip"`
	InstanceID string `json:"instance_id"`
}

type floatingIPJSON struct {
	FloatingIps []floatingIP `json:"floating_ips"`
}

func createServer(vpcid string, subnetid string, count int, name string, reqServer vm.ReqServer, jobID string) string {

	//jsonstr :=
	//
	//`{
	//	"server": {
	//		"availability_zone": "cn-north-1a",
	//		"name": "ecs_server",
	//		"imageRef": "1189efbf-d48b-46ad-a823-94b942e2a000",
	//		"root_volume": {
	//			"volumetype": "SATA",
	//			"size": 40
	//		},
	//		"flavorRef": "s2.large.2",
	//		"vpcid": "vpcid",
	//		"nics": [
	//			{
	//				"subnet_id": "subnetid"
	//			}
	//		],
	//		"publicip": {
	//			"eip": {
	//				"iptype": "5_bgp",
	//				"bandwidth": {
	//					"size": 1,
	//					"sharetype": "PER"
	//				}
	//			}
	//		},
	//		"adminPass": "LYP@ql123",
	//		"count": 4
	//	}
	//}`

	rootVolume := volumeInfo{"SATA", 50}
	var dataVolumes = make([]volumeInfo, 1)
	dataVolumes[0] = volumeInfo{"SATA", reqServer.DiskSize}
	var securityGroups = make([]securityGroup, 1)
	securityGroups[0] = securityGroup{"62e99223-c40a-460c-9cce-efa13745d972"}
	subnet := subnetInfo{subnetid}
	nics := []subnetInfo{subnet}
	bandwidth := bandwidthInfo{5, "PER"}
	eip := eipInfo{"5_bgp", bandwidth}
	publicip := publicipInfo{eip}

	server := serverInfo{
		AvailabilityZone: "cn-north-1a",
		Name:             name + "Server",
		ImageRef:         "a91dc7f6-fe27-4f6a-b907-f006b03ef474",
		RootvolumeInfo:   rootVolume,
		DataVolumes:      dataVolumes,
		FlavorRef:        "s2.large.2",
		Vpcid:            vpcid,
		SecurityGroups:   securityGroups,
		Nics:             nics,
		Publicip:         publicip,
		AdminPass:        "Baas@ql123",
		Count:            count,
	}

	jsonserver := serverJSON{server}

	jsonstr, errs := json.Marshal(jsonserver)

	if errs != nil {
		glog.Error(errs)
	}

	url := "https://ecs.cn-north-1.myhuaweicloud.com/v1/ca7f44bee05d492ba9dfe0d67d31e383/cloudservers"

	respmap, err := doHTTP("POST", url, string(jsonstr))

	var jobid = ""

	if err != nil {

		panic(err)

	} else {

		var resmap = make(map[string]string)

		if err := json.Unmarshal([]byte(respmap["body"].(string)), &resmap); err == nil {

			jobid = resmap["job_id"]

		}
	}

	//update jobMap
	jobDetail := vm.JobDetail{jobid, 0, 3, "创建虚拟机", 1, time.Now(), time.Now(), 0, 0, 1}
	JobMap[jobID][2] = jobDetail

	return jobid

}

// QueryServerStatus query huawei ecs server run status
func QueryServerStatus(serverid string) (string, error) {

	url := "https://ecs.cn-north-1.myhuaweicloud.com/v2/" + viper.GetString("huaweicloud.project_id") + "/servers/" + serverid

	respmap, err := doHTTP("GET", url, "")

	if err == nil {
		var resmap map[string]interface{}
		if err := json.Unmarshal([]byte(respmap["body"].(string)), &resmap); err == nil {
			if v, ok := resmap["server"]; ok {
				content := v.(map[string]interface{})
				status := content["status"].(string)
				return status, err
			}
		}
	}
	return "", err
}

func queryFloatingIps() floatingIPJSON {

	url := "https://ecs.cn-north-1.myhuaweicloud.com/v2/ca7f44bee05d492ba9dfe0d67d31e383/os-floating-ips"

	var floatingInfo floatingIPJSON

	respmap, err := doHTTP("GET", url, "")

	if err != nil {
		panic(err)
	} else {

		if err := json.Unmarshal([]byte(respmap["body"].(string)), &floatingInfo); err != nil {
			panic(err)
		}
	}
	return floatingInfo

}
