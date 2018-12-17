package huawei

import (
	"bytes"
	"net/http"
	"time"

	"encoding/json"
	"github.com/glog"
	"github.com/spf13/viper"
)

var token *Token

// Token def huawei cloud token model
type Token struct {
	AuthToken string

	UpdateTime time.Time
}

type auth struct {
	Auth authInfo `json:"auth"`
}

type authInfo struct {
	Identity authIdentity `json:"identity"`
	Scope    authScope    `json:"scope"`
}

type authIdentity struct {
	Methods  []string     `json:"methods"`
	Password authPassword `json:"password"`
}

type authPassword struct {
	User authUser `json:"user"`
}

type authUser struct {
	Name     string     `json:"name"`
	Password string     `json:"password"`
	Domain   userDomain `json:"domain"`
}

type userDomain struct {
	Name string `json:"name"`
}

type authScope struct {
	Project scopeProject `json:"project"`
}

type scopeProject struct {
	ID string `json:"id"`
}

func getToken(force bool) string {

	now := time.Now()

	if force {

		token = &Token{newReqToken(), now}

	} else {

		if token == nil || token.AuthToken == "" || (now.Sub(token.UpdateTime)).Hours() > 20 {

			token = &Token{newReqToken(), now}

		}

	}
	return token.AuthToken

}

func newReqToken() string {

	respStatus, authToken := reqToken()

	var count = 0

	for respStatus != "201 Created" && count < 5 {

		glog.Errorf("req token num: %v", count)

		time.Sleep(5 * time.Second)

		respStatus, authToken = reqToken()

	}

	if count == 5 {
		glog.Errorf("req token err")
	}

	return authToken

}

func reqToken() (respStatus string, authToken string) {

	//jsonstr := `{
	//	"auth":{
	//		"identity":{
	//			"methods":["password"],
	//			"password":{
	//				"user":{
	//					"name":"qulian",
	//					"password":"Huawei@ql",
	//					"domain":{
	//						"name":"qulian"
	//					}
	//				}
	//			}
	//		},
	//		"scope":{
	//			"project":{
	//				"id":"ca7f44bee05d492ba9dfe0d67d31e383"
	//			}
	//		}
	//	}
	//}`

	scopeProject := scopeProject{ID: viper.GetString("huaweicloud.project_id")}

	authScope := authScope{Project: scopeProject}

	userDomain := userDomain{Name: viper.GetString("huaweicloud.domain_name")}

	authUser := authUser{
		Name:     viper.GetString("huaweicloud.name"),
		Password: viper.GetString("huaweicloud.password"),
		Domain:   userDomain,
	}

	authPassword := authPassword{User: authUser}

	authIdentity := authIdentity{
		Methods:  []string{"password"},
		Password: authPassword,
	}

	authInfo := authInfo{
		Identity: authIdentity,
		Scope:    authScope,
	}

	auth := auth{
		Auth: authInfo,
	}

	jsonstr, errs := json.Marshal(auth)

	if errs != nil {
		glog.Errorf("token authInfo to json err: %v", errs)
	}

	url := "https://iam.cn-north-1.myhuaweicloud.com/v3/auth/tokens"

	repbody := bytes.NewBuffer([]byte(string(jsonstr)))

	req, _ := http.NewRequest("POST", url, repbody)

	req.Header.Set("Content-Type", "application/json;charset=utf8")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		glog.Errorf("request token err: %v", err)
		glog.Errorf("request token resp: %v", resp)
	}

	if resp != nil {

		defer resp.Body.Close()

		glog.Info("response Status:", resp.Status)

		glog.Info("response Headers:", resp.Header)

		authToken = resp.Header.Get("X-Subject-Token")

		return resp.Status, authToken

	}

	return

}
