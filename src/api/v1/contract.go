package v1

import (
	"FunnyVoteGo/src/api/vm"
	"FunnyVoteGo/src/service"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glog"
)

func GetContractInfo(c *gin.Context) {
	str, _ := json.Marshal(vm.VoteInit{
		Title:      "aaaa",
		Options:    []string{"111", "222"},
		SelectType: 1,
	})
	glog.Info(string(str[:]))
	info, err := service.GetContractInfo("test1")
	if err != nil {
		vm.MakeFail(c, http.StatusNotFound, err.Error())
	} else {
		vm.MakeSuccess(c, http.StatusOK, info)
	}
	return
}
