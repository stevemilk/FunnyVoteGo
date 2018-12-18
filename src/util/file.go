package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"

	"github.com/glog"
)

// FileToMap parse file to map
func FileToMap(f *multipart.FileHeader) (m map[string]interface{}, err error) {
	//m = make(map[string]interface{})
	fr, err := f.Open()
	defer fr.Close()
	if err != nil {
		return m, err
	}
	fByte, err := ioutil.ReadAll(fr)
	if err != nil {
		glog.Errorf("read private key fail:%s", err.Error())
		return m, err
	}

	glog.Infof("debug private: %s", string(fByte))

	var i interface{}
	e := json.Unmarshal(fByte, &i)
	if e != nil {

		return m, err
	}
	m, ok := i.(map[string]interface{})
	if !ok {
		return m, fmt.Errorf("类型转换失败")
	}
	return m, err

}
