package service

import (
	"FunnyVoteGo/src/api/vm"
	"FunnyVoteGo/src/model"
	"FunnyVoteGo/src/util"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/glog"
	"github.com/hyperchain/gosdk/abi"
	"github.com/hyperchain/gosdk/utils/ecdsa"
)

const (
	ContractAddress = "0x7850702ad5c90ed16568f54db6644209a1bdff2b"
)

// GetContractCode get contract code
func GetContractCode() string {
	cbyte, err := ioutil.ReadFile("./conf/contract/vote.sol")

	if err != nil {
		return ""
	}
	contractcode := string(cbyte[:])
	return contractcode

}

// StartVote start  a vote
func StartVote(voteinit *vm.VoteInit) (string, bool) {
	//调用合约新建投票活动
	// get contract addr and Code
	contractcode := GetContractCode()
	// init params
	vote := model.Vote{
		ID:          util.StringUUID(),
		Title:       voteinit.Title,
		Description: voteinit.Description,
		SelectType:  voteinit.SelectType,
		StartTime:   voteinit.StartTime,
		EndTime:     voteinit.EndTime,
		CreateTime:  util.GetNowTimeString(),
		CreatorID:   voteinit.CreatorID,
	}
	params := util.Struct2String(vote)
	if params == "" {

		glog.Info("222")
		return "", false
	}
	key, err := InitKey()
	if err != nil {
		glog.Info("333")
		return "", false
	}

	retu, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "insertVote",
		MethodParams: params,
	}, key)
	if err != nil {
		return "", false
	}
	// 解析合约返回
	ABI, _ := abi.JSON(strings.NewReader(retu.Abi))
	p1, _, err := constructOutput(ABI, retu.Methods, retu.Result)
	// 0 成功 1 失败
	if p1 != 0 {
		return "", false
	}
	glog.Info("新建投票成功，开始插入选项...")

	b := AddOptions(voteinit.Options, vote.ID, key)
	if !b {
		return "", false
	}
	glog.Info("插入选项成功")
	return vote.ID, true

}

// AddOptions add options for a vote
func AddOptions(options []string, voteid string, key *ecdsa.Key) bool {
	for _, option := range options {
		//调用合约添加选项内容
		params := util.Struct2String(model.Option{
			ID:      util.StringUUID(),
			VoteID:  voteid,
			Content: option,
			// TODO 这个要去掉
			Total: 1,
		})

		contractcode := GetContractCode()
		retu, err := InvokeContract(vm.ReqInvokeCon{
			ContractAddr: ContractAddress,
			ContractCode: contractcode,
			MethodName:   "insertVoteOption",
			MethodParams: params,
		}, key)
		if err != nil {
			return false
		}
		// 解析合约返回
		ABI, _ := abi.JSON(strings.NewReader(retu.Abi))
		p1, _, err := constructOutput(ABI, retu.Methods, retu.Result)
		// 0 成功 1 失败
		if p1 != 0 {
			return false
		}

	}
	return true
}

func ChooseOption(chooseoption *vm.ChooseOption) bool {
	contractcode := GetContractCode()
	key, err := InitKey()
	if err != nil {
		return false
	}

	// 第一个合约  选项+1
	params1 := util.Struct2String(model.Option{
		ID: chooseoption.OptionID,
	})
	retu1, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "updateVoteOption",
		MethodParams: params1,
	}, key)
	if err != nil {
		return false
	}
	// 解析合约返回
	ABI, _ := abi.JSON(strings.NewReader(retu1.Abi))
	p1, _, err := constructOutput(ABI, retu1.Methods, retu1.Result)
	if p1 != 0 {
		return false
	}

	glog.Info("1 finish")
	// 第二个合约 插入用户id,选项id等
	params2 := util.Struct2String(model.UserOption{
		ID:            util.StringUUID(),
		VoteID:        chooseoption.VoteID,
		OptionID:      chooseoption.OptionID,
		OptionContent: chooseoption.OptionContent,
		UserID:        chooseoption.UserID,
		Publickey:     util.RandString(25),
		CreateTime:    util.GetNowTimeString(),
	})
	retu2, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "insertVoteResult",
		MethodParams: params2,
	}, key)
	if err != nil {
		return false
	}
	p2, _, err := constructOutput(ABI, retu2.Methods, retu2.Result)
	if p2 != 0 {
		return false
	}
	glog.Info("2 finish")
	return true

}

func GetVoteStatus(getvotestatus *vm.GetVoteStatus) (*model.Vote, bool) {
	contractcode := GetContractCode()
	key, err := InitKey()
	if err != nil {
		glog.Info("333")
		return nil, false
	}

	// 第一个合约 获取投票信息判断活动时间
	params1 := util.Struct2String(model.Vote{
		ID: getvotestatus.VoteID,
	})
	retu1, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "queryVote",
		MethodParams: params1,
	}, key)
	if err != nil {
		return nil, false
	}
	// 处理第一个合约返回
	ABI, _ := abi.JSON(strings.NewReader(retu1.Abi))
	var p_ok int32
	var p_title [32]byte
	var p_desc [32]byte
	var p_type int32
	var p_st [32]byte
	var p_et [32]byte
	var p_ct [32]byte
	var p_cid [32]byte
	res := []interface{}{&p_ok, &p_title, &p_desc, &p_type, &p_st, &p_et, &p_ct, &p_cid}
	if sysErr := ABI.UnpackResult(&res, "queryVote", retu1.Result); sysErr != nil {
		glog.Info(sysErr)
		return nil, false
	}

	var vote model.Vote
	vote.ID = getvotestatus.VoteID
	vote.Title = util.Byte32ToString(p_title)
	vote.Description = util.Byte32ToString(p_desc)
	vote.SelectType = int(p_type)
	vote.StartTime = util.Byte32ToString(p_st)
	vote.EndTime = util.Byte32ToString(p_et)
	vote.CreateTime = util.Byte32ToString(p_et)
	creatorid, _ := strconv.Atoi(util.Byte32ToString(p_cid))
	vote.CreatorID = uint(creatorid)

	//add vote  status
	starttime, _ := strconv.Atoi(vote.StartTime)
	endtime, _ := strconv.Atoi(vote.EndTime)
	nowtime := int(time.Now().Unix())
	if starttime < nowtime {
		vote.Status = 1
	} else if endtime < starttime {
		vote.Status = 3
	} else {
		vote.Status = 2
	}

	glog.Infof("vote: %+v", vote)
	glog.Info("1 finish")
	// 第二个合约 获得选项内容
	params2 := util.Struct2String(model.Vote{
		ID: getvotestatus.VoteID,
	})
	retu2, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "queryVoteOption",
		MethodParams: params2,
	}, key)
	if err != nil {
		return nil, false
	}
	// 处理第二个合约返回
	var p_ok2 int32
	var p_oarray [][32]byte
	var p_barray [][32]byte
	var p_iarray []int32
	res2 := []interface{}{&p_ok2, &p_oarray, &p_barray, &p_iarray}
	if sysErr := ABI.UnpackResult(&res2, "queryVoteOption", retu2.Result); sysErr != nil {
		glog.Info(sysErr)
		return nil, false
	}
	if p_ok2 == 1 {
		return nil, false
	}
	glog.Info(len(p_oarray))
	glog.Info(len(p_barray))
	glog.Info(len(p_iarray))
	var options []model.Option
	for i := 0; i < len(p_iarray); i++ {
		var option model.Option
		optionid := util.ByteToString(p_oarray[i][:])
		tal := util.ByteToString(p_barray[i][:])
		option.ID = optionid
		option.Total = uint(p_iarray[i])
		option.Content = tal
		options = append(options, option)

	}
	vote.Options = options

	glog.Infof("vote: %+v", vote)
	glog.Info("2 finish")
	// 第三个合约 判断是否投过票
	params3 := util.Struct2String(model.UserOption{
		UserID: getvotestatus.UserID,
		VoteID: getvotestatus.VoteID,
	})
	retu3, err := InvokeContract(vm.ReqInvokeCon{
		ContractAddr: ContractAddress,
		ContractCode: contractcode,
		MethodName:   "queryVoteResult",
		MethodParams: params3,
	}, key)
	if err != nil {
		return nil, false
	}
	// 处理第三个合约返回
	var p_ok3 int32
	var p_bo bool
	res3 := []interface{}{&p_ok3, &p_bo}
	if sysErr := ABI.UnpackResult(&res3, "queryVoteResult", retu3.Result); sysErr != nil {
		glog.Info(sysErr)
		return nil, false
	}
	glog.Info(p_bo)
	if p_bo {
		vote.UserVoted = 2
	} else {
		vote.UserVoted = 1
	}
	glog.Info("3 finish")
	return &vote, true
}
