package v1

import (
	"FunnyVoteGo/src/api/vm"
	"FunnyVoteGo/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StartVote start a vote
func StartVote(c *gin.Context) {
	var voteinit vm.VoteInit
	if err := c.ShouldBindJSON(&voteinit); err != nil {
		vm.MakeFail(c, http.StatusBadRequest, "参数错误")
		return
	}
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
	// 是否投过票
	// 投票是否开始、结束
	//
	var getvotestatus vm.GetVoteStatus
	if err := c.ShouldBindJSON(&getvotestatus); err != nil {
		vm.MakeFail(c, http.StatusBadRequest, "参数错误")
		return
	}
	//service.GetVoteStatus()

}

// Vote chooses one option
func Vote(c *gin.Context) {
	var chooseoption vm.ChooseOption
	if err := c.ShouldBindJSON(&chooseoption); err != nil {
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
