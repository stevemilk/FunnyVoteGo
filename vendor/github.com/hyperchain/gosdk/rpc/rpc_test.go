package rpc

import (
	"fmt"
	"github.com/hyperchain/gosdk/abi"
	"github.com/hyperchain/gosdk/account"
	"github.com/hyperchain/gosdk/common"
	"github.com/hyperchain/gosdk/utils/gm"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	contractAddress = "0x421a1fb06bd9c9fae9b8cdaf8a662cf3c41ffa10"
	abiContract     = `[{"constant":false,"inputs":[{"name":"num1","type":"uint32"},{"name":"num2","type":"uint32"}],"name":"add","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"archiveSum","outputs":[],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"getSum","outputs":[{"name":"","type":"uint32"}],"payable":false,"type":"function"},{"constant":false,"inputs":[],"name":"increment","outputs":[],"payable":false,"type":"function"}]`
	binContract     = "0x60606040526000805463ffffffff19169055341561001957fe5b5b61012a806100296000396000f300606060405263ffffffff60e060020a6000350416633ad14af38114603e57806348fe842114605c578063569c5f6d14606b578063d09de08a146091575bfe5b3415604557fe5b605a63ffffffff6004358116906024351660a0565b005b3415606357fe5b605a60c2565b005b3415607257fe5b607860d2565b6040805163ffffffff9092168252519081900360200190f35b3415609857fe5b605a60df565b005b6000805463ffffffff808216850184011663ffffffff199091161790555b5050565b6000805463ffffffff191690555b565b60005463ffffffff165b90565b6000805463ffffffff8082166001011663ffffffff199091161790555b5600a165627a7a72305820caa934a33fe993d03f87bdf39706fada68ddde78182e0110fd43e8c323d5984a0029"
)

var (
	rpc           = NewRPCWithPath("../conf")
	rpc2, _       = rpc.BindNodes(2, 3, 4)
	address       = "bfa5bd992e3eb123c8b86ebe892099d4e9efb783"
	privateKey, _ = account.NewAccountFromPriv("a1fd6ed6225e76aac3884b5420c8cdbb4fde1db01e9ef773415b8f2b5a9b77d4")

	guomiPub      = "02739518af5e065b22dabb35ea5369a4c64d4865565874a006399bbb0e62e18004"
	guomiPri      = "6153af264daa4763490f2a51c9d13417ef9f579229be2141574eb339ee9b9d2a"
	guomiKey, err = gm.GetKeyPareFromHex(guomiPri, guomiPub)
)

type TestAsyncHandler struct {
	t        *testing.T
	IsCalled bool
}

func (tah *TestAsyncHandler) OnSuccess(receipt *TxReceipt) {
	tah.IsCalled = true
	fmt.Println(receipt.Ret)
}

func (tah *TestAsyncHandler) OnFailure(err StdError) {
	tah.t.Error(err.String())
}

func TestNewRpc(t *testing.T) {
	rpc := NewRPC()
	logger.Info(rpc)
}

func TestNewRpcWithPath(t *testing.T) {
	rpc := NewRPCWithPath("../conf")
	logger.Info(rpc)
}

func TestRpc_GetNodes(t *testing.T) {
	nodeInfo, err := rpc.GetNodes()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(len(nodeInfo))
	logger.Info(nodeInfo)
}

/*---------------------------------- contract ----------------------------------*/

func TestRpc_CompileContract(t *testing.T) {
	compileContract("../conf/contract/Accumulator.sol")
}

func TestRpc_DeployContract(t *testing.T) {
	cr, _ := compileContract("../conf/contract/Accumulator.sol")
	transaction := NewTransaction(guomiKey.GetAddress()).Deploy(cr.Bin[0])
	transaction.Sign(guomiKey)
	receipt, _ := rpc.DeployContract(transaction)
	fmt.Println("address:", receipt.ContractAddress)
}

func TestRpc_DeployContractAsync(t *testing.T) {
	cr, _ := compileContract("../conf/contract/Accumulator.sol")
	transaction := NewTransaction(guomiKey.GetAddress()).Deploy(cr.Bin[0])
	transaction.Sign(guomiKey)
	asyncHandler := TestAsyncHandler{t: t}
	rpc.DeployContractAsync(transaction, &asyncHandler)
	time.Sleep(3 * time.Second)
	assert.EqualValues(t, true, asyncHandler.IsCalled, "回调未被执行")
}

func TestRpc_InvokeContract(t *testing.T) {
	cr, _ := compileContract("../conf/contract/Accumulator.sol")
	contractAddress, err := deployContract(cr.Bin[0], cr.Abi[0])
	ABI, serr := abi.JSON(strings.NewReader(cr.Abi[0]))
	if err != nil {
		t.Error(serr)
		return
	}
	packed, serr := ABI.Pack("add", uint32(1), uint32(2))
	if err != nil {
		t.Error(serr)
		return
	}
	transaction := NewTransaction(address).Invoke(contractAddress, packed)
	transaction.Sign(privateKey)
	receipt, _ := rpc.InvokeContract(transaction)
	fmt.Println("ret:", receipt.Ret)
}

func TestRpc_InvokeContractAsync(t *testing.T) {
	ABI, err := abi.JSON(strings.NewReader(abiContract))
	if err != nil {
		t.Error(err)
		return
	}
	packed, err := ABI.Pack("add", uint32(1), uint32(2))
	if err != nil {
		t.Error(err)
		return
	}
	transaction := NewTransaction(address).Invoke(contractAddress, packed)
	transaction.Sign(privateKey)
	asyncHandler := TestAsyncHandler{t: t}
	rpc.InvokeContractAsync(transaction, &asyncHandler)
	time.Sleep(3 * time.Second)
	assert.EqualValues(t, true, asyncHandler.IsCalled, "回调未被执行")
}

func TestRpc_DeployContractWithArgs(t *testing.T) {
	cr, _ := compileContract("../conf/contract/Accumulator2.sol")
	var arg [32]byte
	copy(arg[:], "test")
	transaction := NewTransaction(guomiKey.GetAddress()).Deploy(cr.Bin[0]).DeployArgs(cr.Abi[0], uint32(10), arg)
	transaction.Sign(guomiKey)
	receipt, _ := rpc.DeployContract(transaction)
	fmt.Println("address:", receipt.ContractAddress)

	fmt.Println("-----------------------------------")

	ABI, _ := abi.JSON(strings.NewReader(cr.Abi[0]))
	packed, _ := ABI.Pack("getMul")
	transaction1 := NewTransaction(address).Invoke(receipt.ContractAddress, packed)
	transaction1.Sign(privateKey)
	receipt1, err := rpc.InvokeContract(transaction1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("ret:", receipt1.Ret)

	var p0 []byte
	var p1 int64
	var p2 common.Address
	testV := []interface{}{&p0, &p1, &p2}
	fmt.Println(reflect.TypeOf(testV))
	decode(ABI, &testV, "getMul", receipt1.Ret)
	fmt.Println(string(p0), p1, p2.Hex())
}

func TestRPC_UnpackLog(t *testing.T) {
	cr, _ := compileContract("../conf/contract/Accumulator2.sol")
	var arg [32]byte
	copy(arg[:], "test")
	transaction := NewTransaction(guomiKey.GetAddress()).Deploy(cr.Bin[0]).DeployArgs(cr.Abi[0], uint32(10), arg)
	transaction.Sign(guomiKey)
	receipt, _ := rpc.DeployContract(transaction)
	fmt.Println("address:", receipt.ContractAddress)

	fmt.Println("-----------------------------------")

	ABI, _ := abi.JSON(strings.NewReader(cr.Abi[0]))
	packed, _ := ABI.Pack("getHello")
	transaction1 := NewTransaction(address).Invoke(receipt.ContractAddress, packed)
	transaction1.Sign(privateKey)
	receipt1, err := rpc.InvokeContract(transaction1)
	if err != nil {
		t.Error(err)
		return
	}
	test := struct {
		Addr int64   `abi:"addr1"`
		Msg1 [8]byte `abi:"msg"`
	}{}

	// testLog
	sysErr := ABI.UnpackLog(&test, "sayHello", receipt1.Log[0].Data, receipt1.Log[0].Topics)
	if sysErr != nil {
		t.Error(sysErr)
		return
	}
	msg, sysErr := abi.ByteArrayToString(test.Msg1)
	if sysErr != nil {
		t.Error(sysErr)
		return
	}
	assert.Equal(t, int64(1), test.Addr, "解码失败")
	assert.Equal(t, "test", msg, "解码失败")
}

func TestRPC_SendTx(t *testing.T) {
	transaction := NewTransaction(guomiKey.GetAddress()).Transfer(address, int64(0))
	transaction.Sign(guomiKey)
	receipt, err := rpc.SendTx(transaction)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(receipt.Ret)
}

func TestSM2Account(t *testing.T) {
	//accountJson, _ := account.NewAccountSm2("")
	strAcc := `{"address":"0xf97e4add7dfb6ae1b70a499c4cf5fbd722830aeb","publicKey":"04678d408df9cd88ad9e4850a3e292aba356402552b69ce660872e884308d73d62ee8c774b6783424705c5e91080ec8b14ebcb76f55415c93ff3d072d19f2cc096","privateKey":"00e496a9a178cfc856d369abf67445d669a977ccd87db36eb18a9c76baa326cf31","privateKeyEncrypted":false}`
	key, _ := account.NewAccountSm2FromAccountJSON(strAcc, "")
	//fmt.Println(accountJson)
	//key, _ := account.NewAccountSm2FromAccountJSON(accountJson, "")
	transaction := NewTransaction(key.GetAddress()).Transfer(address, int64(0))
	transaction.Sign(key)
	receipt, err := rpc.SendTx(transaction)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, 66, len(receipt.TxHash))

	accountJSON, _ := account.NewAccountSm2("123")
	aKey, syserr := account.NewAccountSm2FromAccountJSON(accountJSON, "123")
	if syserr != nil {
		t.Error(syserr)
	}
	transaction1 := NewTransaction(aKey.GetAddress()).Transfer(address, int64(0))
	transaction1.Sign(aKey)
	receipt1, err := rpc.SendTx(transaction1)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, 66, len(receipt1.TxHash))
}

func TestRPC_SendTxAsync(t *testing.T) {
	transaction := NewTransaction(guomiKey.GetAddress()).Transfer(address, int64(0))
	transaction.Sign(guomiKey)
	asyncHandler := TestAsyncHandler{t: t}
	rpc.SendTxAsync(transaction, &asyncHandler)
	time.Sleep(3 * time.Second)
	assert.EqualValues(t, true, asyncHandler.IsCalled, "回调未被执行")
}

// maintain contract by opcode 1
func TestRPC_MaintainContract(t *testing.T) {
	contractOriginFile := "../conf/contract/Accumulator.sol"
	contractUpdateFile := "../conf/contract/AccumulatorUpdate.sol"
	compileOrigin, _ := compileContract(contractOriginFile)
	compileUpdate, _ := compileContract(contractUpdateFile)
	contractAddress, err := deployContract(compileOrigin.Bin[0], compileOrigin.Abi[0])
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("contractAddress:", contractAddress)

	// test invoke before update
	ABIBefore, serr := abi.JSON(strings.NewReader(compileOrigin.Abi[0]))
	packed, serr := ABIBefore.Pack("add", uint32(11), uint32(1))
	if err != nil {
		t.Error(serr)
		return
	}
	transactionInvokeBe := NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, packed)
	transactionInvokeBe.Sign(guomiKey)
	receiptBe, err := rpc.InvokeContract(transactionInvokeBe)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(receiptBe.Ret)

	var result1 uint32
	decode(ABIBefore, &result1, "add", receiptBe.Ret)
	fmt.Println(result1)

	fmt.Println("-----------------------------")

	transactionUpdate := NewTransaction(guomiKey.GetAddress()).Maintain(1, contractAddress, compileUpdate.Bin[0])
	//transactionUpdate, err := NewMaintainTransaction(guomiKey.GetAddress(), contractAddress, compileUpdate.Bin[0], 1, EVM)
	transactionUpdate.Sign(guomiKey)
	receiptUpdate, err := rpc.MaintainContract(transactionUpdate)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(receiptUpdate.ContractAddress)

	// test invoke after update
	ABI, serr := abi.JSON(strings.NewReader(compileUpdate.Abi[0]))
	if err != nil {
		t.Error(err)
		return
	}
	packed2, serr := ABI.Pack("addUpdate", uint32(1), uint32(2))
	if err != nil {
		t.Error(serr)
		return
	}
	transactionInvoke := NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, packed2)
	//transactionInvoke, err := NewInvokeTransaction(guomiKey.GetAddress(), contractAddress, common.ToHex(packed2), false, EVM)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	transactionInvoke.Sign(guomiKey)
	receiptInvoke, err := rpc.InvokeContract(transactionInvoke)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(receiptInvoke.Ret)
	var result2 uint32
	decode(ABI, &result2, "addUpdate", receiptInvoke.Ret)
	fmt.Println(result2)
}

// maintain contract by opcode 2 and 3
func TestRPC_MaintainContract2(t *testing.T) {
	contractAddress, _ := deployContract(binContract, abiContract)
	ABI, _ := abi.JSON(strings.NewReader(abiContract))
	// invoke first
	packed, _ := ABI.Pack("getSum")
	transaction1 := NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, packed)
	//transaction1, _ := NewInvokeTransaction(guomiKey.GetAddress(), contractAddress, common.ToHex(packed), false, EVM)
	transaction1.Sign(guomiKey)
	receipt1, err := rpc.InvokeContract(transaction1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("invoke first:", receipt1.Ret)

	// freeze contract
	transactionFreeze := NewTransaction(guomiKey.GetAddress()).Maintain(2, contractAddress, "")
	//transactionFreeze, _ := NewMaintainTransaction(guomiKey.GetAddress(), contractAddress, "", 2, EVM)
	transactionFreeze.Sign(guomiKey)
	receiptFreeze, err := rpc.MaintainContract(transactionFreeze)
	fmt.Println(receiptFreeze.TxHash)
	status, err := rpc.GetContractStatus(contractAddress)
	fmt.Println("contract status >>", status)

	// invoke after freeze
	transaction2 := NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, packed)
	//transaction2, _ := NewInvokeTransaction(guomiKey.GetAddress(), contractAddress, common.ToHex(packed), false, EVM)
	transaction2.Sign(guomiKey)
	receipt2, err := rpc.InvokeContract(transaction2)
	if err != nil {
		fmt.Println("invoke second receipt2 is null ", receipt2 == nil)
		fmt.Println(err)
	}

	// unfreeze contract
	transactionUnfreeze := NewTransaction(guomiKey.GetAddress()).Maintain(3, contractAddress, "")
	//transactionUnfreeze, _ := NewMaintainTransaction(guomiKey.GetAddress(), contractAddress, "", 3, EVM)
	transactionUnfreeze.Sign(guomiKey)
	receiptUnFreeze, err := rpc.MaintainContract(transactionUnfreeze)
	fmt.Println(receiptUnFreeze.TxHash)
	status, _ = rpc.GetContractStatus(contractAddress)
	fmt.Println("contract status >>", status)

	// invoke after unfreeze
	transaction3 := NewTransaction(guomiKey.GetAddress()).Invoke(contractAddress, packed)
	//transaction3, _ := NewInvokeTransaction(guomiKey.GetAddress(), contractAddress, common.ToHex(packed), false, EVM)
	transaction3.Sign(guomiKey)
	receipt3, err := rpc.InvokeContract(transaction3)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("invoke third:", receipt3.Ret)
}

func TestRPC_GetContractStatus(t *testing.T) {
	t.Skip("the node can get the account")
	contractAddress, _ := deployContract(binContract, abiContract)
	statu, err := rpc.GetContractStatus(contractAddress)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(statu)
}

func TestRPC_GetDeployedList(t *testing.T) {
	list, err := rpc.GetDeployedList(guomiKey.GetAddress())
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(len(list))
}

func TestRPC_InvokeContractReturnHash(t *testing.T) {
	t.Skip("pressure test, do not put this test in CI")
	cr, _ := compileContract("../conf/contract/Accumulator.sol")
	contractAddress, err := deployContract(cr.Bin[0], cr.Abi[0])
	ABI, serr := abi.JSON(strings.NewReader(cr.Abi[0]))
	if err != nil {
		t.Error(serr)
		return
	}
	packed, serr := ABI.Pack("add", uint32(1), uint32(2))
	if err != nil {
		t.Error(serr)
		return
	}
	transaction := NewTransaction(address).Invoke(contractAddress, packed)
	//transaction, err := NewInvokeTransaction(address, contractAddress, common.ToHex(packed), false, EVM)
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	transaction.Sign(privateKey)
	var hash string
	tt := time.After(1 * time.Minute)
	counter := 0
	for {
		hash, err = rpc.InvokeContractReturnHash(transaction)
		select {
		case <-tt:
			fmt.Println(counter)
			return
		default:
			counter++
		}

	}
	fmt.Println(counter)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, 66, len(hash))
	fmt.Println("hash:", hash)
}

/*---------------------------------- archive ----------------------------------*/

func TestRPC_Snapshot(t *testing.T) {
	t.Skip()
	res, err := rpc.Snapshot(1)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(res)
}

func TestRPC_QuerySnapshotExist(t *testing.T) {
	t.Skip()
	res, err := rpc.QuerySnapshotExist("0x5d86cce7e537cd0e0346468889801196")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(res)
}

func TestRPC_CheckSnapshot(t *testing.T) {
	t.Skip()
	res, err := rpc.CheckSnapshot("0x5d86cce7e537cd0e0346468889801196")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(res)
}

func TestRPC_Archive(t *testing.T) {
	t.Skip()
	res, err := rpc.Archive("0x5d86cce7e537cd0e0346468889801196", false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(res)
}

func TestRPC_QueryArchiveExist(t *testing.T) {
	t.Skip()
	res, err := rpc.QueryArchiveExist("0x5d86cce7e537cd0e0346468889801196")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(res)
}

/*---------------------------------- node ----------------------------------*/

func TestRPC_GetNodeHash(t *testing.T) {
	hash, err := rpc.GetNodeHash()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(hash)
}

func TestRPC_GetNodeHashById(t *testing.T) {
	id := 1
	hash, err := rpc.GetNodeHashByID(id)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(hash)
}

func TestRPC_DeleteNodeVP(t *testing.T) {
	t.Skip("do not delete VP in CI")
	hash1, _ := rpc.GetNodeHashByID(1)
	success, _ := rpc.DeleteNodeVP(hash1)
	assert.Equal(t, true, success)

	hash11, _ := rpc.GetNodeHashByID(1)
	fmt.Println(hash11)
}

func TestRPC_DeleteNodeNVP(t *testing.T) {
	t.Skip("do not delete NVP in CI")
	hash1, _ := rpc.GetNodeHashByID(1)
	success, _ := rpc.DeleteNodeNVP(hash1)
	assert.Equal(t, true, success)
}

func TestRPC_GetNodeStates(t *testing.T) {
	states, err := rpc.GetNodeStates()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, 4, len(states))
}

/*---------------------------------- block ----------------------------------*/

func TestRPC_GetLatestBlock(t *testing.T) {
	block, err := rpc.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(block)
}

func TestRPC_GetBlocks(t *testing.T) {
	latestBlock, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}
	blocks, err := rpc2.GetBlocks(latestBlock.Number-1, latestBlock.Number, true)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(blocks)
}

func TestRPC_GetBlockByHash(t *testing.T) {
	latestBlock, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}

	block, err := rpc2.GetBlockByHash(latestBlock.Hash, true)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(block)
}

func TestRPC_GetBatchBlocksByHash(t *testing.T) {
	latestBlock, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}

	blocks, err := rpc2.GetBatchBlocksByHash([]string{latestBlock.Hash}, true)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(blocks)
}

func TestRPC_GetBlockByNumber(t *testing.T) {
	latestBlock, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}

	rpc.GetBlockByNumber("latest", false)
	block, err := rpc2.GetBlockByNumber(latestBlock.Number, true)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(block)
}

func TestRPC_GetBatchBlocksByNumber(t *testing.T) {
	latestBlock, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}

	blocks, err := rpc2.GetBatchBlocksByNumber([]uint64{latestBlock.Number}, true)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(blocks)
}

func TestRPC_GetAvgGenTimeByBlkNum(t *testing.T) {
	block, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}
	avgTime, err := rpc2.GetAvgGenTimeByBlockNum(block.Number-2, block.Number)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(avgTime)
}

func TestRPC_GetBlockByTime(t *testing.T) {
	t.Skip()
	blockInterval, err := rpc.GetBlocksByTime(1, 1778959217012956575)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(blockInterval)
}

func TestRPC_QueryTPS(t *testing.T) {
	tpsInfo, err := rpc.QueryTPS(1, 1778959217012956575)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(tpsInfo)
}

func TestRPC_GetGenesisBlock(t *testing.T) {
	blkNum, err := rpc.GetGenesisBlock()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, true, strings.HasPrefix(blkNum, "0x"))
}

func TestRPC_GetChainHeight(t *testing.T) {
	blkNum, err := rpc.GetChainHeight()
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, true, strings.HasPrefix(blkNum, "0x"))
}

/*---------------------------------- transaction ----------------------------------*/

func TestRPC_GetTransactions(t *testing.T) {
	block, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}
	txs, err := rpc2.GetTransactionsByBlkNum(block.Number-1, block.Number)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(txs[0].Invalid)
}

func TestRPC_GetDiscardTx(t *testing.T) {
	txs, err := rpc.GetDiscardTx()
	if err != nil {
		//t.Error(err)
		return
	}
	fmt.Println(len(txs))
	fmt.Println(txs[len(txs)-1].Hash)
}

func TestRPC_GetTransactionByHash(t *testing.T) {
	t.Skip()
	transaction := NewTransaction(guomiKey.GetAddress()).Deploy(binContract)
	transaction.Sign(guomiKey)
	receipt, _ := rpc.DeployContract(transaction)
	fmt.Println("txhash:", receipt.TxHash)

	hash := receipt.TxHash
	tx, err := rpc.GetTransactionByHash(hash)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(tx.Hash)
	assert.Equal(t, receipt.TxHash, tx.Hash)
}

func TestRPC_GetBatchTxByHash(t *testing.T) {
	transaction1 := NewTransaction(guomiKey.GetAddress()).Deploy(binContract)
	transaction1.Sign(guomiKey)
	receipt1, _ := rpc.DeployContract(transaction1)
	fmt.Println("txhash1:", receipt1.TxHash)

	transaction2 := NewTransaction(guomiKey.GetAddress()).Deploy(binContract)
	transaction2.Sign(guomiKey)
	receipt2, _ := rpc.DeployContract(transaction2)
	fmt.Println("txhash2:", receipt2.TxHash)

	txhashes := make([]string, 0)
	txhashes = append(txhashes, receipt1.TxHash, receipt2.TxHash)

	txs, err := rpc2.GetBatchTxByHash(txhashes)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(len(txs))
	fmt.Println(txs[0].Hash, txs[1].Hash)
	assert.Equal(t, receipt1.TxHash, txs[0].Hash)
	assert.Equal(t, receipt2.TxHash, txs[1].Hash)
}

func TestRPC_GetTxByBlkHashAndIdx(t *testing.T) {
	block, _ := rpc2.GetLatestBlock()
	info, err := rpc2.GetTxByBlkHashAndIdx(block.Hash, 0)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(info)
	assert.EqualValues(t, 66, len(info.Hash))
}

func TestRPC_GetTxByBlkNumAndIdx(t *testing.T) {
	block, _ := rpc2.GetLatestBlock()
	info, err := rpc2.GetTxByBlkNumAndIdx(block.Number, 0)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(info)
	assert.EqualValues(t, 66, len(info.Hash))
}

func TestRPC_GetTxAvgTimeByBlockNumber(t *testing.T) {
	block, _ := rpc2.GetLatestBlock()
	time, err := rpc2.GetTxAvgTimeByBlockNumber(block.Number-2, block.Number)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(time)
}

func TestRPC_GetBatchReceipt(t *testing.T) {
	block, _ := rpc2.GetLatestBlock()
	trans, _ := rpc2.GetTransactionsByBlkNum(block.Number-2, block.Number)
	hashes := []string{trans[0].Hash, trans[1].Hash}
	txs, err := rpc2.GetBatchReceipt(hashes)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 2, len(txs))
}

func TestRPC_GetTxCountByContractAddr(t *testing.T) {
	cAddress, _ := deployContract(binContract, abiContract)
	ABI, _ := abi.JSON(strings.NewReader(abiContract))
	packed, _ := ABI.Pack("getSum")
	transaction := NewTransaction(guomiKey.GetAddress()).Invoke(cAddress, packed)
	transaction.Sign(guomiKey)
	rpc.InvokeContract(transaction)
	transaction2 := NewTransaction(guomiKey.GetAddress()).Invoke(cAddress, packed)
	transaction2.Sign(guomiKey)
	rpc.InvokeContract(transaction2)

	block, _ := rpc2.GetLatestBlock()
	count, err := rpc2.GetTxCountByContractAddr(block.Number-1, block.Number, cAddress, false)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, 2, count.Count)
}

func TestRPC_GetTxByTime(t *testing.T) {
	t.Skip("the length of result is too long")
	infos, err := rpc.GetTxByTime(1, uint64(time.Now().UnixNano()))
	if err != nil {
		t.Error(err)
		return
	}
	//fmt.Println(infos)
	assert.EqualValues(t, true, len(infos) > 0)
}

func TestRPC_GetNextPageTxs(t *testing.T) {
	t.Skip("hyperchain snapshot will case error")
	cAddress, _ := deployContract(binContract, abiContract)
	ABI, _ := abi.JSON(strings.NewReader(abiContract))
	packed, _ := ABI.Pack("getSum")
	transaction := NewTransaction(guomiKey.GetAddress()).Invoke(cAddress, packed)
	transaction.Sign(guomiKey)
	rpc.InvokeContract(transaction)
	transaction2 := NewTransaction(guomiKey.GetAddress()).Invoke(cAddress, packed)
	transaction2.Sign(guomiKey)
	rpc.InvokeContract(transaction2)

	block, _ := rpc2.GetLatestBlock()

	infos, err := rpc2.GetNextPageTxs(block.Number-10, 0, 1, block.Number, 0, 10, false, cAddress)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, 3, len(infos))
}

func TestRPC_GetPrevPageTxs(t *testing.T) {
	t.Skip("hyperchain snapshot will case error")
	cAddress, _ := deployContract(binContract, abiContract)
	ABI, _ := abi.JSON(strings.NewReader(abiContract))
	packed, _ := ABI.Pack("getSum")
	transaction := NewTransaction(guomiKey.GetAddress()).Invoke(cAddress, packed)
	transaction.Sign(guomiKey)
	rpc.InvokeContract(transaction)
	transaction2 := NewTransaction(guomiKey.GetAddress()).Invoke(cAddress, packed)
	transaction2.Sign(guomiKey)
	rpc.InvokeContract(transaction2)

	block, _ := rpc2.GetLatestBlock()

	infos, err := rpc2.GetPrevPageTxs(block.Number, 0, 1, block.Number, 0, 10, false, cAddress)
	if err != nil {
		t.Error(err)
		return
	}
	assert.EqualValues(t, 2, len(infos))
}

func TestRPC_GetBlkTxCountByHash(t *testing.T) {
	block, err := rpc2.GetLatestBlock()
	if err != nil {
		t.Error(err)
		return
	}
	count, err := rpc2.GetBlkTxCountByHash(block.Hash)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(count)
}

func TestRPC_GetTxCount(t *testing.T) {
	txCount, err := rpc.GetTxCount()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(txCount.Count)
}

/*---------------------------------- cert ----------------------------------*/

func TestRPC_GetTCert(t *testing.T) {
	tCert, err := rpc.GetTCert(1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(tCert)
}

/*---------------------------------- account ----------------------------------*/

func TestRPC_GetBalance(t *testing.T) {
	account := "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"
	balance, err := rpc.GetBalance(account)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(balance)
}

/**************************** self function ******************************/

func compileContract(path string) (*CompileResult, error) {
	contract, _ := common.ReadFileAsString(path)
	cr, err := rpc.CompileContract(contract)
	if err != nil {
		logger.Error("can not get compile return, ", err.String())
		return nil, err
	}
	fmt.Println("abi:", cr.Abi[0])
	fmt.Println("bin:", cr.Bin[0])
	fmt.Println("type:", cr.Types[0])

	return cr, err
}

func decode(contractAbi abi.ABI, v interface{}, method string, ret string) (result interface{}) {
	if err := contractAbi.UnpackResult(v, method, ret); err != nil {
		logger.Error(NewSystemError(err).String())
	}
	result = v
	return result
}

func deployContract(bin, abi string, params ...interface{}) (string, StdError) {
	var transaction *Transaction
	var err StdError
	if len(params) == 0 {
		transaction = NewTransaction(guomiKey.GetAddress()).Deploy(bin)
		//transaction, err = NewDeployTransaction(guomiKey.GetAddress(), bin, false, EVM)
	} else {
		transaction = NewTransaction(guomiKey.GetAddress()).Deploy(bin).DeployArgs(abi, params)
		//transaction, err = NewDeployTransactionWithArgs(guomiKey.GetAddress(), bin, false, EVM, abi, params)
	}
	if err != nil {
		return "", err
	}
	transaction.Sign(guomiKey)
	txReceipt, err := rpc.DeployContract(transaction)
	if err != nil {
		logger.Error(err)
	}
	return txReceipt.ContractAddress, nil
}
