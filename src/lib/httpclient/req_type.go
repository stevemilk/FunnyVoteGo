package httpclient

// ReqTransaction req
type ReqTransaction struct {
	Counters  []string `json:"counters"`
	StartTime int64    `json:"start_time"`
	Hostnames []string `json:"hostnames"`
	EndTime   int64    `json:"end_time"`
	ConsolFun string   `json:"consol_fun"`
}

// ReqBlock req
type ReqBlock struct {
	Counters  []string `json:"counters"`
	StartTime int64    `json:"start_time"`
	Hostnames []string `json:"hostnames"`
	EndTime   int64    `json:"end_time"`
	ConsolFun string   `json:"consol_fun"`
}

// ReqTPS req
type ReqTPS struct {
	Counters  []string `json:"counters"`
	StartTime int64    `json:"start_time"`
	Hostnames []string `json:"hostnames"`
	EndTime   int64    `json:"end_time"`
	ConsolFun string   `json:"consol_fun"`
}

// ReqLatency req
type ReqLatency struct {
	Counters  []string `json:"counters"`
	StartTime int64    `json:"start_time"`
	Hostnames []string `json:"hostnames"`
	EndTime   int64    `json:"end_time"`
	ConsolFun string   `json:"consol_fun"`
}

// ReqLatest req
type ReqLatest struct {
	Counter  string `json:"counter"`
	Endpoint string `json:"endpoint"`
}

//func (body ReqTransaction) struct2Json() ([]byte, error) {
//json, err := json.Marshal(body)
//if err != nil {
//glog.Errorf("json marshall error: %v", err)
//return nil, err
//}
//return json, nil
//}

//func (body ReqBlock) struct2Json() ([]byte, error) {
//json, err := json.Marshal(body)
//if err != nil {
//glog.Errorf("json marshall error: %v", err)
//return nil, err
//}
//return json, nil
//}

//func (body ReqTPS) struct2Json() ([]byte, error) {
//json, err := json.Marshal(body)
//if err != nil {
//glog.Errorf("json marshall error: %v", err)
//return nil, err
//}
//return json, nil
//}

//func (body ReqLatency) struct2Json() ([]byte, error) {
//json, err := json.Marshal(body)
//if err != nil {
//glog.Errorf("json marshall error: %v", err)
//return nil, err
//}
//return json, nil
/*}*/
