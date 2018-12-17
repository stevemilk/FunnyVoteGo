package huawei

import "encoding/json"

type subjobEntity struct {
	ServerID string `json:"server_id"`
}

type subjob struct {
	Status     string       `json:"status"`
	Entities   subjobEntity `json:"entities"`
	JobID      string       `json:"job_id"`
	BeginTime  string       `json:"begin_time"`
	EndTime    string       `json:"end_time"`
	ErrorCode  string       `json:"error_code"`
	FailReason string       `json:"fail_reason"`
}

type entity struct {
	SubJobsTotal int      `json:"sub_jobs_total"`
	SubJobs      []subjob `json:"sub_jobs"`
}

// JobInfo define huawei cloud job model
type JobInfo struct {
	Status     string `json:"status"`
	Entities   entity `json:"entities"`
	JobID      string `json:"job_id"`
	JobType    string `json:"job_type"`
	BeginTime  string `json:"begin_time"`
	EndTime    string `json:"end_time"`
	ErrorCode  string `json:"error_code"`
	FailReason string `json:"fail_reason"`
}

// QueryJob query create ecs server job status
func QueryJob(jobid string) JobInfo {

	url := "https://ecs.cn-north-1.myhuaweicloud.com/v1/ca7f44bee05d492ba9dfe0d67d31e383/jobs/" + jobid

	respmap, err := doHTTP("GET", url, "")

	var job JobInfo

	if err != nil {
		panic(err)
	} else {

		if err := json.Unmarshal([]byte(respmap["body"].(string)), &job); err != nil {

			panic(err)

		}

	}

	return job
}
