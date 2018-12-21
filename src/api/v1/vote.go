package v1

import (
	"FunnyVoteGo/src/api/vm"
	"FunnyVoteGo/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/glog"
)

// StartVote start a vote
func StartVote(c *gin.Context) {
	var voteinit vm.VoteInit
	if err := c.ShouldBindJSON(&voteinit); err != nil {
		glog.Error(err)
		vm.MakeFail(c, http.StatusBadRequest, "参数错误")
		return
	}
	glog.Info(voteinit)

	//vm.MakeSuccess(c, http.StatusOK, "oo")
	//return

	glog.Info("1111111111111111111")
	glog.Info(voteinit.Options)
	glog.Info(len(voteinit.Options))

	voteid, b := service.StartVote(&voteinit)
	if !b {
		vm.MakeFail(c, http.StatusInternalServerError, "fail")
		return
	}
	vm.MakeSuccess(c, http.StatusOK, voteid)
	return
}

// VoteStatus returns status of the vote
func VoteStatus(c *gin.Context) {
	var getvotestatus vm.GetVoteStatus
	if err := c.ShouldBind(&getvotestatus); err != nil {
		vm.MakeFail(c, http.StatusBadRequest, "参数错误")
		return
	}
	vote, b := service.GetVoteStatus(&getvotestatus)
	if !b {
		vm.MakeFail(c, http.StatusInternalServerError, "fail")
		return
	}
	vm.MakeSuccess(c, http.StatusOK, vote)
	return

}

// Vote chooses one option
func Vote(c *gin.Context) {
	var chooseoption vm.ChooseOption
	if err := c.ShouldBind(&chooseoption); err != nil {
		vm.MakeFail(c, http.StatusBadRequest, "参数错误")
		return
	}
	b := service.ChooseOption(&chooseoption)
	if !b {
		vm.MakeFail(c, http.StatusInternalServerError, "fail")
		return
	}

	vm.MakeSuccess(c, http.StatusOK, "success")
	return
}
