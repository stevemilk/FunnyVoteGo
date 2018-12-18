package v1

import (
	"FunnyVoteGo/src/api/vm"
	"FunnyVoteGo/src/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartVote(c *gin.Context) {
	var voteinit vm.VoteInit
	if err := c.ShouldBindJSON(&voteinit); err != nil {
		vm.MakeFail(c, http.StatusBadRequest, "参数错误")
		return
	}
	_, err := service.InitKey(c)
	if err != nil {
		vm.MakeFail(c, http.StatusInternalServerError, err.Error())
		return
	}
	//service.StartVote(&voteinit, key)
}
