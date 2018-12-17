package huawei

import (
	"bytes"
	"hyperbaas/src/util"
	"io/ioutil"
	"net/http"

	"fmt"

	"github.com/glog"
)

func doHTTP(method string, url string, jsonstr string) (map[string]interface{}, error) {

	defer httpPanicHandler()

	var reqbody *bytes.Buffer

	if jsonstr != "" {

		reqbody = bytes.NewBuffer([]byte(jsonstr))

		glog.Infof("reqbody: %v", reqbody)

	}

	var req *http.Request

	if jsonstr == "" {
		req, _ = http.NewRequest(method, url, nil)
	} else {
		req, _ = http.NewRequest(method, url, reqbody)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", getToken(false))

	client := &http.Client{}

	resp, err := client.Do(req)

	var resmap = make(map[string]interface{})

	if err != nil {
		glog.Errorf("do http request failed: %v", err)
		util.PostErrorf("do http request failed, req = %v, err = %v", req, err)
		return resmap, err
	}

	if resp.Status == "401 Unauthorized" {

		req.Header.Set("X-Auth-Token", getToken(true))

		client := &http.Client{}

		resp, err = client.Do(req)

	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.Status != "200 OK" && resp.Status != "201 Created" {

		glog.Errorf("response Status: %v", resp.Status)

		glog.Errorf("response Headers: %v", resp.Header)

		glog.Errorf("response Body: %v", string(body))

	}

	resmap["status"] = resp.Status

	resmap["headers"] = resp.Header

	resmap["body"] = string(body)

	return resmap, err

}

// post http panic error and resume
func httpPanicHandler() {
	ers := "网络请求错误\r\n"
	if err := recover(); err != nil {
		ers += fmt.Sprintf(fmt.Sprintf("%v\r\n", err)) // output panic info
		ers += fmt.Sprintf("========\r\n")
		util.PostErrorf(ers)
	}
}
