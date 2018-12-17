package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/glog"
	"github.com/hyperchain/gosdk/common"
	"github.com/terasum/viper"
)

const (
	// TRANSACTION type
	TRANSACTION = "tx_"
	// CONTRACT type
	CONTRACT = "contract_"
	// BLOCK type
	BLOCK = "block_"
	// ACCOUNT type
	ACCOUNT = "account_"
	// NODE type
	NODE = "node_"
	// CERT type
	CERT = "cert_"
	// SUB type
	SUB = "sub_"
	// ARCHIVE type
	ARCHIVE = "archive_"
	// MQ type
	MQ = "mq_"
)

var (
	logger = common.GetLogger("rpc")
	once   = sync.Once{}
)

// RPC represents rpc apis
type RPC struct {
	hrm                HTTPRequestManager
	namespace          string
	resTime            int64
	firstPollInterval  int64
	firstPollTime      int64
	secondPollInterval int64
	secondPollTime     int64
	reConnTime         int64
}

func (r *RPC) String() string {
	nodes := r.hrm.nodes
	var nodeString string
	nodeString += "["
	for i, v := range nodes {
		nodeString += "{\"index\":" + strconv.Itoa(i) + ", \"url:\"" + v.url + "}"
		if i < len(nodes)-1 {
			nodeString += ", "
		}
	}
	nodeString += "]"
	return "\"namespace\":" + r.namespace + ", \"nodeUrl\":" + nodeString
}

// NewRPC get a RPC instance with default conf directory path "../conf"
func NewRPC() *RPC {
	return NewRPCWithPath(common.DefaultConfRootPath)
}

// post most panic error and resume
func panicHandler() {
	if err := recover(); err != nil {
		ers := "[recover] rpc panic: %v"
		glog.Errorf(ers, err)
	}
}

// NewRPCWithPath get a RPC instance with user defined root conf directory path
// the default conf root file structure should like this:
//
//      conf
//		├── certs
//		│   ├── ecert.cert
//		│   ├── ecert.priv
//		│   ├── sdkcert.cert
//		│   ├── sdkcert.priv
//		│   ├── tls
//		│   │   ├── tls_peer.cert
//		│   │   ├── tls_peer.priv
//		│   │   └── tlsca.ca
//		│   ├── unique.priv
//		│   └── unique.pub
//		└── hpc.toml
func NewRPCWithPath(confRootPath string) *RPC {
	defer panicHandler()

	vip := viper.New()
	vip.SetConfigFile(confRootPath + "/" + common.DefaultConfRelPath)
	err := vip.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("read conf from %s error", confRootPath+"/"+common.DefaultConfRelPath))
	}

	common.InitLog(vip)

	httpRequestManager := newHTTPRequestManager(vip, confRootPath)

	namespace := vip.GetString(common.NamespaceConf)
	logger.Debugf("[CONFIG]: %s = %v", common.NamespaceConf, namespace)

	resTime := vip.GetInt64(common.PollingResendTime)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingResendTime, resTime)

	firstPollInterval := vip.GetInt64(common.PollingFirstPollingInterval)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingFirstPollingInterval, firstPollInterval)

	firstPollTime := vip.GetInt64(common.PollingFirstPollingTimes)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingFirstPollingTimes, firstPollTime)

	secondPollInterval := vip.GetInt64(common.PollingSecondPollingInterval)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingSecondPollingInterval, secondPollInterval)

	secondPollTime := vip.GetInt64(common.PollingSecondPollingTimes)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingSecondPollingTimes, secondPollTime)

	reConnTime := vip.GetInt64(common.ReConnectTime)
	logger.Debugf("[CONFIG]: %s = %v", common.ReConnectTime, reConnTime)

	return &RPC{
		hrm:                *httpRequestManager,
		namespace:          namespace,
		resTime:            resTime,
		firstPollInterval:  firstPollInterval,
		firstPollTime:      firstPollTime,
		secondPollInterval: secondPollInterval,
		secondPollTime:     secondPollTime,
		reConnTime:         reConnTime,
	}
}

// BindNodes generate a new RPC instance that bind with given indexes
func (r *RPC) BindNodes(nodeIndexes ...int) (*RPC, error) {
	if len(nodeIndexes) == 0 {
		return r, nil
	}
	proxy := *r
	proxy.hrm.nodes = make([]*Node, len(nodeIndexes))
	proxy.hrm.nodeIndex = 0

	limit := len(r.hrm.nodes)
	for i := 0; i < len(nodeIndexes); i++ {
		if nodeIndexes[i] > limit {
			return nil, fmt.Errorf("nodeIndex %d is out of range", i)
		}
		proxy.hrm.nodes[i] = r.hrm.nodes[nodeIndexes[i]-1]
	}
	return &proxy, nil
}

// package method name and params to JsonRequest
func (r *RPC) jsonRPC(method string, params ...interface{}) *JSONRequest {
	req := &JSONRequest{
		Method:    method,
		Version:   JSONRPCVersion,
		ID:        1,
		Namespace: r.namespace,
		Params:    params,
	}
	return req
}

// call is a function to get response result commodiously
func (r *RPC) call(method string, params ...interface{}) (json.RawMessage, StdError) {
	req := r.jsonRPC(method, params...)
	return r.callWithReq(req)
}

// callWithReq is a function to get response origin data
func (r *RPC) callWithReq(req *JSONRequest) (json.RawMessage, StdError) {
	body, sysErr := json.Marshal(req)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	data, err := r.hrm.SyncRequest(body)
	if err != nil {
		return nil, err
	}

	var resp *JSONResponse
	if sysErr = json.Unmarshal(data, &resp); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	if resp.Code != SuccessCode {
		return nil, NewServerError(resp.Code, resp.Message)
	}

	return resp.Result, nil
}

// callWithSpecificUrl is a function to get response form specific url
func (r *RPC) callWithSpecificURL(method string, url string, params ...interface{}) (json.RawMessage, StdError) {
	req := r.jsonRPC(method, params...)

	body, sysErr := json.Marshal(req)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	data, err := r.hrm.SyncRequestSpecificURL(body, url)
	if err != nil {
		return nil, err
	}

	var resp *JSONResponse
	if sysErr = json.Unmarshal(data, &resp); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	if resp.Code != SuccessCode {
		return nil, NewServerError(resp.Code, resp.Message)
	}

	return resp.Result, nil
}

// Call call and get tx receipt directly without polling
func (r *RPC) Call(method string, param interface{}) (*TxReceipt, StdError) {
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}
	var receipt TxReceipt
	if sysErr := json.Unmarshal(data, &receipt); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &receipt, nil
}

// CallByPolling call and get tx receipt by polling
func (r *RPC) CallByPolling(method string, param interface{}) (*TxReceipt, StdError) {
	var (
		req    *JSONRequest
		data   json.RawMessage
		hash   string
		err    StdError
		sysErr error
	)
	// if simulate is false, transaction need to resend
	req = r.jsonRPC(method, param)

	for i := int64(0); i < r.resTime; i++ {
		if data, err = r.callWithReq(req); err != nil {
			if err.Code() == DuplicateTransactionsCode {
				// -32007: 交易重复
				s := strings.Split(string(data), " ")
				if len(s) >= 3 {
					hash = s[2]
				}
				txReceipt, innErr, success := r.GetTxReceiptByPolling(hash)
				err = innErr
				if success {
					return txReceipt, err
				}
				continue
			} else if err.Code() == GetResponseErrorCode || err.Code() == SystemErrorCode {
				// resend
			} else if err.Code() != SystemBusyCode && err.Code() != DataNotExistCode {
				// -9999: 获取响应失败
				// -32001: 查询的数据不存在
				// -32006: 系统繁忙
				return nil, err
			}
		} else {
			if sysErr = json.Unmarshal(data, &hash); sysErr != nil {
				return nil, NewSystemError(sysErr)
			}
			txReceipt, innErr, success := r.GetTxReceiptByPolling(hash)
			err = innErr
			if success {
				return txReceipt, err
			}
			continue
		}
		//if code is -9999 -32001 and -32006, we should sleep then resend
		time.Sleep(time.Millisecond * time.Duration(r.firstPollInterval+r.secondPollInterval))
	}
	return nil, NewRequestTimeoutError(errors.New("request time out"))
}

// GetTxReceiptByPolling get tx receipt by polling
func (r *RPC) GetTxReceiptByPolling(txHash string) (*TxReceipt, StdError, bool) {
	var (
		err     StdError
		receipt *TxReceipt
	)
	txHash = chPrefix(txHash)

	for j := int64(0); j < r.firstPollTime; j++ {
		receipt, err = r.GetTxReceipt(txHash)
		if err != nil {
			if err.Code() == BalanceInsufficientCode {
				return nil, err, true
			} else if err.Code() != DataNotExistCode && err.Code() != SystemBusyCode {
				return nil, err, true
			}
			time.Sleep(time.Millisecond * time.Duration(r.firstPollInterval))
		} else {
			return receipt, nil, true
		}
	}
	for j := int64(0); j < r.secondPollTime; j++ {
		receipt, err = r.GetTxReceipt(txHash)
		if err != nil {
			if err.Code() == BalanceInsufficientCode {
				return nil, err, true
			} else if err.Code() != DataNotExistCode && err.Code() != SystemBusyCode {
				return nil, err, true
			}
			time.Sleep(time.Millisecond * time.Duration(r.firstPollInterval))
		} else {
			return receipt, nil, true
		}
	}
	return nil, NewGetResponseError(errors.New("polling failure")), false
}

/*---------------------------------- node ----------------------------------*/

// GetNodes 获取区块链节点信息
func (r *RPC) GetNodes() ([]NodeInfo, StdError) {
	data, err := r.call(NODE + "getNodes")
	if err != nil {
		return nil, err
	}
	var nodeInfo []NodeInfo
	if sysErr := json.Unmarshal(data, &nodeInfo); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return nodeInfo, nil
}

// GetNodeHash 获取随机节点hash
func (r *RPC) GetNodeHash() (string, StdError) {
	data, err := r.call(NODE + "getNodeHash")
	if err != nil {
		return "", err
	}
	hash := []byte(data)
	return string(hash), nil
}

// GetNodeHashByID 从指定节点获取hash
func (r *RPC) GetNodeHashByID(id int) (string, StdError) {
	url := r.hrm.nodes[id-1].url
	data, err := r.callWithSpecificURL(NODE+"getNodeHash", url)
	if err != nil {
		return "", err
	}

	var hash string
	if sysErr := json.Unmarshal(data, &hash); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return hash, nil
}

// DeleteNodeVP 删除VP节点
func (r *RPC) DeleteNodeVP(hash string) (bool, StdError) {
	method := NODE + "deleteVP"
	param := NewMapParam("nodehash", hash)
	_, err := r.call(method, param.Serialize())
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteNodeNVP 删除NVP节点
func (r *RPC) DeleteNodeNVP(hash string) (bool, StdError) {
	method := NODE + "deleteNVP"
	param := NewMapParam("nodehash", hash)
	_, err := r.call(method, param.Serialize())
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetNodeStates 获取节点状态信息
func (r *RPC) GetNodeStates() ([]NodeStateInfo, StdError) {
	method := NODE + "getNodeStates"
	data, err := r.call(method)
	if err != nil {
		return nil, err
	}

	var list []NodeStateInfo
	if sysErr := json.Unmarshal(data, &list); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return list, nil
}

/*---------------------------------- block ----------------------------------*/

// GetLatestBlock returns information about the latest block
func (r *RPC) GetLatestBlock() (*Block, StdError) {
	method := BLOCK + "latestBlock"

	data, stdErr := r.call(method)
	if stdErr != nil {
		return nil, stdErr
	}

	blockRaw := BlockRaw{}

	sysErr := json.Unmarshal(data, &blockRaw)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	block, stdErr := blockRaw.ToBlock()
	if stdErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return block, nil
}

// GetBlocks returns a list of blocks from start block number to end block number
// isPlain indicates if the result includes transaction information. if false, includes, otherwise not.
func (r *RPC) GetBlocks(from, to uint64, isPlain bool) ([]*Block, StdError) {
	if from <= 0 || to <= 0 || to < from {
		return nil, NewSystemError(errors.New("参数必须为非0正整数，且to应该大于from"))
	}

	method := BLOCK + "getBlocks"

	mp := NewMapParam("from", from)
	mp.addKV("to", to)
	mp.addKV("isPlain", isPlain)

	data, stdErr := r.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaws []BlockRaw

	sysErr := json.Unmarshal(data, &blockRaws)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blocks := make([]*Block, 0, len(blockRaws))

	for _, v := range blockRaws {
		block, stdErr := v.ToBlock()
		if stdErr != nil {
			return nil, stdErr
		}

		blocks = append(blocks, block)
	}

	return blocks, nil

}

// GetBlockByHash returns information about a block by hash.
// If the param isPlain value is true, it returns block excluding transactions. If false,
// it returns block including transactions.
func (r *RPC) GetBlockByHash(blockHash string, isPlain bool) (*Block, StdError) {
	method := BLOCK + "getBlockByHash"

	data, stdErr := r.call(method, blockHash, isPlain)
	if stdErr != nil {
		return nil, stdErr
	}

	blockRaw := BlockRaw{}
	if sysErr := json.Unmarshal(data, &blockRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	block, stdErr := blockRaw.ToBlock()
	if stdErr != nil {
		return nil, stdErr
	}

	return block, nil
}

// GetBatchBlocksByHash returns a list of blocks by a list of specific block hash.
func (r *RPC) GetBatchBlocksByHash(blockHashes []string, isPlain bool) ([]*Block, StdError) {
	method := BLOCK + "getBatchBlocksByHash"

	mp := NewMapParam("hashes", blockHashes)
	mp.addKV("isPlain", isPlain)

	data, stdErr := r.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaws []BlockRaw

	sysErr := json.Unmarshal(data, &blockRaws)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blocks := make([]*Block, 0, len(blockRaws))

	for _, v := range blockRaws {
		block, stdErr := v.ToBlock()
		if stdErr != nil {
			return nil, stdErr
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// GetBlockByNumber returns information about a block by number. If the param isPlain
// value is true, it returns block excluding transactions. If false, it returns block
// including transactions.
// blockNum can use `latest`, means get latest block
func (r *RPC) GetBlockByNumber(blockNum interface{}, isPlain bool) (*Block, StdError) {
	method := BLOCK + "getBlockByNumber"

	data, stdErr := r.call(method, blockNum, isPlain)
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaw BlockRaw

	sysErr := json.Unmarshal(data, &blockRaw)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	block, stdErr := blockRaw.ToBlock()
	if stdErr != nil {
		return nil, stdErr
	}

	return block, nil
}

// GetBatchBlocksByNumber returns a list of blocks by a list of specific block number.
func (r *RPC) GetBatchBlocksByNumber(blockNums []uint64, isPlain bool) ([]*Block, StdError) {
	method := BLOCK + "getBatchBlocksByNumber"

	mp := NewMapParam("numbers", blockNums)
	mp.addKV("isPlain", isPlain)

	data, stdErr := r.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaws []BlockRaw

	sysErr := json.Unmarshal(data, &blockRaws)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blocks := make([]*Block, 0, len(blockRaws))

	for _, v := range blockRaws {
		block, stdErr := v.ToBlock()
		if stdErr != nil {
			return nil, stdErr
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// GetAvgGenTimeByBlockNum calculates the average generation time of all blocks
// for the given block number.
func (r *RPC) GetAvgGenTimeByBlockNum(from, to uint64) (int64, StdError) {
	if from <= 0 || to <= 0 || to < from {
		return -1, NewSystemError(errors.New("参数必须为非0正整数，且to应该大于from"))
	}

	method := BLOCK + "getAvgGenerateTimeByBlockNumber"

	mp := NewMapParam("from", from)
	mp.addKV("to", to)

	data, stdErr := r.call(method, mp.Serialize())
	if stdErr != nil {
		return -1, stdErr
	}

	str := strings.Replace(string(data), "\"", "", 2)

	if strings.Index(str, "0x") == 0 || strings.Index(str, "-0x") == 0 {
		str = strings.Replace(str, "0x", "", 1)
	}

	avgTime, sysErr := strconv.ParseInt(str, 16, 64)
	if sysErr != nil {
		return -1, NewSystemError(sysErr)
	}

	return avgTime, nil
}

// GetBlocksByTime returns the number of blocks, starting block and ending block
// at specific time periods.
// startTime and endTime are timestamps
func (r *RPC) GetBlocksByTime(startTime, endTime uint64) (*BlockInterval, StdError) {
	if endTime < startTime {
		return nil, NewSystemError(errors.New("startTime必须小于endTime"))
	}

	method := BLOCK + "getBlocksByTime"

	mp := NewMapParam("startTime", startTime)
	mp.addKV("endTime", endTime)

	data, stdErr := r.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockNumRaw BlockIntervalRaw

	sysErr := json.Unmarshal(data, &blockNumRaw)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blockNum, stdErr := blockNumRaw.ToBlockInterval()
	if stdErr != nil {
		return nil, stdErr
	}

	return blockNum, nil
}

// QueryTPS queries the block generation speed and tps within a given time range.
func (r *RPC) QueryTPS(startTime, endTime uint64) (*TPSInfo, StdError) {
	if endTime < startTime {
		return nil, NewSystemError(errors.New("startTime必须小于endTime"))
	}

	method := BLOCK + "queryTPS"

	mp := NewMapParam("startTime", startTime)
	mp.addKV("endTime", endTime)

	data, stdErr := r.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	items := strings.Split(string(data), ";")

	startTimeStr := items[0][12:]
	endTimeStr := items[1][9:]
	totalBlock, sysErr := strconv.ParseUint(strings.Split(items[2], ":")[1], 0, 64)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	blockPerSec, sysErr := strconv.ParseFloat(strings.Split(items[3], ":")[1], 64)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	tps, sysErr := strconv.ParseFloat(strings.Split(items[4], ":")[1], 64)

	return &TPSInfo{
		StartTime:     startTimeStr,
		EndTime:       endTimeStr,
		TotalBlockNum: totalBlock,
		BlocksPerSec:  blockPerSec,
		Tps:           tps,
	}, nil
}

// GetGenesisBlock returns current genesis block number.
// result is hex string
func (r *RPC) GetGenesisBlock() (string, StdError) {
	method := BLOCK + "getGenesisBlock"

	data, stdErr := r.call(method)
	if stdErr != nil {
		return "", stdErr
	}

	var result string
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

// GetChainHeight returns the current chain height.
// result is hex string
func (r *RPC) GetChainHeight() (string, StdError) {
	method := BLOCK + "getChainHeight"

	data, stdErr := r.call(method)
	if stdErr != nil {
		return "", stdErr
	}

	var result string
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

/*---------------------------------- transaction ----------------------------------*/

// GetTransactionsByBlkNum 根据区块号查询范围内的交易
func (r *RPC) GetTransactionsByBlkNum(start, end uint64) ([]TransactionInfo, StdError) {
	qtr := &QueryTxRange{
		From: start,
		To:   end,
	}
	method := TRANSACTION + "getTransactions"
	param := qtr.Serialize()
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetDiscardTx 获取所有非法交易
func (r *RPC) GetDiscardTx() ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getDiscardTransactions"
	data, err := r.call(method)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetTransactionByHash 通过交易hash获取交易
// 参数txHash应该是"0x...."的形式
func (r *RPC) GetTransactionByHash(txHash string) (*TransactionInfo, StdError) {
	method := TRANSACTION + "getTransactionByHash"
	param := txHash
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var tx TransactionRaw
	if sysErr := json.Unmarshal(data, &tx); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return tx.ToTransaction()
}

// GetBatchTxByHash 批量获取交易
func (r *RPC) GetBatchTxByHash(hashes []string) ([]TransactionInfo, StdError) {
	mp := NewMapParam("hashes", hashes)
	method := TRANSACTION + "getBatchTransactions"
	param := mp.Serialize()
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetTxByBlkHashAndIdx 通过区块hash和交易序号返回交易信息
func (r *RPC) GetTxByBlkHashAndIdx(blkHash string, index uint64) (*TransactionInfo, StdError) {
	method := TRANSACTION + "getTransactionByBlockHashAndIndex"
	data, err := r.call(method, blkHash, index)
	if err != nil {
		return nil, err
	}

	var tx TransactionRaw
	if sysErr := json.Unmarshal(data, &tx); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return tx.ToTransaction()
}

// GetTxByBlkNumAndIdx 通过区块号和交易序号查询交易
func (r *RPC) GetTxByBlkNumAndIdx(blkNum, index uint64) (*TransactionInfo, StdError) {
	method := TRANSACTION + "getTransactionByBlockNumberAndIndex"
	data, err := r.call(method, strconv.FormatUint(blkNum, 10), index)
	if err != nil {
		return nil, err
	}

	var tx TransactionRaw
	if sysErr := json.Unmarshal(data, &tx); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return tx.ToTransaction()
}

// GetTxAvgTimeByBlockNumber 通过区块号区间获取交易平均处理时间
func (r *RPC) GetTxAvgTimeByBlockNumber(from, to uint64) (uint64, StdError) {
	mp := NewMapParam("from", from)
	mp.addKV("to", to)
	method := TRANSACTION + "getTxAvgTimeByBlockNumber"
	param := mp.Serialize()
	data, err := r.call(method, param)
	if err != nil {
		return 0, err
	}

	var avgTime string
	if sysErr := json.Unmarshal(data, &avgTime); sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	result, sysErr := strconv.ParseUint(avgTime, 0, 64)
	if err != nil {
		return 0, NewSystemError(sysErr)
	}
	return result, nil
}

// GetBatchReceipt 批量获取回执
func (r *RPC) GetBatchReceipt(hashes []string) ([]TxReceipt, StdError) {
	mp := NewMapParam("hashes", hashes)
	method := TRANSACTION + "getBatchReceipt"
	param := mp.Serialize()
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var txs []TxReceipt
	if sysErr := json.Unmarshal(data, &txs); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return txs, nil
}

// GetBlkTxCountByHash 通过区块hash获取区块上交易数
func (r *RPC) GetBlkTxCountByHash(blkHash string) (uint64, StdError) {
	method := TRANSACTION + "getBlockTransactionCountByHash"
	param := blkHash
	data, err := r.call(method, param)
	if err != nil {
		return 0, err
	}

	var hexCount string
	if sysError := json.Unmarshal(data, &hexCount); sysError != nil {
		return 0, NewSystemError(err)
	}
	count, sysErr := strconv.ParseUint(hexCount, 0, 64)
	if sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	return count, nil
}

// GetTxCount 获取链上所有交易数量
func (r *RPC) GetTxCount() (*TransactionsCount, StdError) {
	mehtod := TRANSACTION + "getTransactionsCount"
	data, err := r.call(mehtod)
	if err != nil {
		return nil, err
	}

	var txRaw TransactionsCountRaw
	if sysErr := json.Unmarshal(data, &txRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	txCount, sysErr := txRaw.ToTransactionsCount()
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return txCount, nil
}

// GetTxCountByContractAddr 查询区块间指定合约的交易量 txExtra过滤是否带有额外字段
func (r *RPC) GetTxCountByContractAddr(from, to uint64, address string, txExtra bool) (*TransactionsCountByContract, StdError) {
	mp := NewMapParam("from", from).addKV("to", to).addKV("address", address).addKV("txExtra", txExtra)
	method := TRANSACTION + "getTransactionsCountByContractAddr"
	param := mp.Serialize()
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var countRaw *TransactionsCountByContractRaw
	if sysErr := json.Unmarshal(data, &countRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	count, sysErr := countRaw.ToTransactionsCountByContract()
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return count, nil
}

// GetTxByTime 根据范围时间戳查询交易信息
func (r *RPC) GetTxByTime(start, end uint64) ([]TransactionInfo, StdError) {
	mp := NewMapParam("startTime", start).addKV("endTime", end)
	method := TRANSACTION + "getTransactionsByTime"
	param := mp.Serialize()
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetNextPageTxs 获取下一页的交易
func (r *RPC) GetNextPageTxs(blkNumber, txIndex, minBlkNumber, maxBlkNumber, separated, pageSize uint64, containCurrent bool, contractAddr string) ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getNextPageTransactions"
	param := &TransactionPageArg{
		strconv.FormatUint(blkNumber, 10),
		strconv.FormatUint(maxBlkNumber, 10),
		strconv.FormatUint(minBlkNumber, 10),
		txIndex, separated, pageSize, containCurrent, contractAddr}
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetPrevPageTxs 获取上一页的交易
func (r *RPC) GetPrevPageTxs(blkNumber, txIndex, minBlkNumber, maxBlkNumber, separated, pageSize uint64, containCurrent bool, contractAddr string) ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getPrevPageTransactions"
	param := &TransactionPageArg{
		strconv.FormatUint(blkNumber, 10),
		strconv.FormatUint(maxBlkNumber, 10),
		strconv.FormatUint(minBlkNumber, 10),
		txIndex, separated, pageSize, containCurrent, contractAddr}
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetTxReceipt 通过交易hash获取交易回执
// 参数txHash应该是"0x...."的形式
func (r *RPC) GetTxReceipt(txHash string) (*TxReceipt, StdError) {
	txHash = chPrefix(txHash)
	method := TRANSACTION + "getTransactionReceipt"
	param := txHash
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}

	var txr TxReceipt
	if sysErr := json.Unmarshal(data, &txr); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &txr, nil
}

// SendTx 同步发送交易
func (r *RPC) SendTx(transaction *Transaction) (*TxReceipt, StdError) {
	method := TRANSACTION + "sendTransaction"
	param := transaction.Serialize()
	if transaction.simulate {
		return r.Call(method, param)
	}
	return r.CallByPolling(method, param)
}

// SendTxAsync 异步发送交易
func (r *RPC) SendTxAsync(transaction *Transaction, handler AsyncHandler) {
	asyncResult := Asyncify(r.SendTx)(transaction)
	go func() {
		res, err := asyncResult.GetResult()
		if err != nil {
			handler.OnFailure(err)
		} else {
			handler.OnSuccess(res)
		}
	}()
}

/*---------------------------------- contract ----------------------------------*/

// CompileContract Compile contract rpc
func (r *RPC) CompileContract(code string) (*CompileResult, StdError) {
	data, err := r.call(CONTRACT+"compileContract", code)
	if err != nil {
		return nil, err
	}

	var cr CompileResult
	if sysErr := json.Unmarshal(data, &cr); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &cr, nil
}

// DeployContract Deploy contract rpc
func (r *RPC) DeployContract(transaction *Transaction) (*TxReceipt, StdError) {
	method := CONTRACT + "deployContract"
	transaction.isDeploy = true
	param := transaction.Serialize()
	if transaction.simulate {
		return r.Call(method, param)
	}
	return r.CallByPolling(method, param)
}

// DeployContractAsync deploy contract async rpc
func (r *RPC) DeployContractAsync(transaction *Transaction, handler AsyncHandler) {
	asyncResult := Asyncify(r.DeployContract)(transaction)
	go func() {
		res, err := asyncResult.GetResult()
		if err != nil {
			handler.OnFailure(err)
		} else {
			handler.OnSuccess(res)
		}
	}()
}

// InvokeContract invoke contract rpc
func (r *RPC) InvokeContract(transaction *Transaction) (*TxReceipt, StdError) {
	method := CONTRACT + "invokeContract"
	transaction.isInvoke = true
	param := transaction.Serialize()
	if transaction.simulate {
		return r.Call(method, param)
	}
	return r.CallByPolling(method, param)
}

// InvokeContractAsync invoke contract async rpc
func (r *RPC) InvokeContractAsync(transaction *Transaction, handler AsyncHandler) {
	asyncResult := Asyncify(r.InvokeContract)(transaction)
	go func() {
		res, err := asyncResult.GetResult()
		if err != nil {
			handler.OnFailure(err)
		} else {
			handler.OnSuccess(res)
		}
	}()
}

// MaintainContract 管理合约 opcode
// 1.升级合约
// 2.冻结
// 3.解冻
func (r *RPC) MaintainContract(transaction *Transaction) (*TxReceipt, StdError) {
	method := CONTRACT + "maintainContract"
	transaction.isMaintain = true
	param := transaction.Serialize()
	return r.CallByPolling(method, param)
}

// MaintainContractAsync maintain contract async
func (r *RPC) MaintainContractAsync(transaction *Transaction, handler AsyncHandler) {
	asyncResult := Asyncify(r.MaintainContract)(transaction)
	go func() {
		res, err := asyncResult.GetResult()
		if err != nil {
			handler.OnFailure(err)
		} else {
			handler.OnSuccess(res)
		}
	}()
}

// GetContractStatus 获取合约状态
func (r *RPC) GetContractStatus(contractAddress string) (string, StdError) {
	method := CONTRACT + "getStatus"
	param := contractAddress
	data, err := r.call(method, param)
	if err != nil {
		return "", err
	}
	result := string([]byte(data))
	return result, nil
}

// GetDeployedList 获取已部署的合约列表
func (r *RPC) GetDeployedList(address string) ([]string, StdError) {
	method := CONTRACT + "getDeployedList"
	param := address
	data, err := r.call(method, param)
	if err != nil {
		return nil, err
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, NewSystemError(err)
	}
	return result, nil
}

// InvokeContractReturnHash for pressure test
// Deprecated:
func (r *RPC) InvokeContractReturnHash(transaction *Transaction) (string, StdError) {
	method := CONTRACT + "invokeContract"
	param := transaction.Serialize()
	data, err := r.call(method, param)
	if err != nil {
		return "", err
	}

	var hash string
	if sysErr := json.Unmarshal(data, &hash); err != nil {
		return "", NewSystemError(sysErr)
	}

	return hash, nil
}

/*---------------------------------- sub ----------------------------------*/

// GetWebSocketClient 获取WebSocket客户端
func (r *RPC) GetWebSocketClient() *WebSocketClient {
	once.Do(func() {
		globalWebSocketClient = &WebSocketClient{
			conns:   make(map[int]*connectionWrapper, len(r.hrm.nodes)),
			hrm:     &r.hrm,
			rwMutex: sync.RWMutex{},
		}
	})

	return globalWebSocketClient
}

/*---------------------------------- mq ----------------------------------*/

// GetMqClient 获取mq客户端
func (r *RPC) GetMqClient() *MqClient {
	once.Do(func() {
		mqClient = &MqClient{
			mqConns: make(map[uint]*mqWrapper, len(r.hrm.nodes)),
			hrm:     &r.hrm,
		}
	})

	return mqClient
}

/*---------------------------------- archive ----------------------------------*/

// Snapshot makes the snapshot for given the future block number or current the latest block number.
// It returns the snapshot id for the client to query.
// blockHeight can use `latest`, means make snapshot now
func (r *RPC) Snapshot(blockHeight interface{}) (string, StdError) {
	method := ARCHIVE + "snapshot"

	data, stdErr := r.call(method, blockHeight)
	if stdErr != nil {
		return "", stdErr
	}

	var result string

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

// QuerySnapshotExist checks if the given snapshot existed, so you can confirm that
// the last step Archive.Snapshot is successful.
func (r *RPC) QuerySnapshotExist(filterID string) (bool, StdError) {
	method := ARCHIVE + "querySnapshotExist"

	data, stdErr := r.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// CheckSnapshot will check that the snapshot is correct. If correct, returns true.
// Otherwise, returns false.
func (r *RPC) CheckSnapshot(filterID string) (bool, StdError) {
	method := ARCHIVE + "checkSnapshot"

	data, stdErr := r.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// DeleteSnapshot delete snapshot by id
func (r *RPC) DeleteSnapshot(filterID string) (bool, StdError) {
	method := ARCHIVE + "deleteSnapshot"

	data, stdErr := r.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// ListSnapshot returns all the existed snapshot information.
func (r *RPC) ListSnapshot() (Manifests, StdError) {
	method := ARCHIVE + "listSnapshot"

	data, stdErr := r.call(method)
	if stdErr != nil {
		return nil, stdErr
	}

	var result Manifests
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return result, nil
}

// ReadSnapshot returns the snapshot information for the given snapshot ID.
func (r *RPC) ReadSnapshot(filterID string) (*Manifest, StdError) {
	method := ARCHIVE + "readSnapshot"

	data, stdErr := r.call(method, filterID)
	if stdErr != nil {
		return nil, stdErr
	}

	var result Manifest
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return &result, nil
}

// Archive will archive data of the given snapshot. If successful, returns true.
func (r *RPC) Archive(filterID string, sync bool) (bool, StdError) {
	method := ARCHIVE + "archive"

	data, stdErr := r.call(method, filterID, sync)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// Restore restores datas that have been archived for given snapshot. If successful, returns true.
func (r *RPC) Restore(filterID string, sync bool) (bool, StdError) {
	method := ARCHIVE + "restore"

	data, stdErr := r.call(method, filterID, sync)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// RestoreAll restores all datas that have been archived. If successful, returns true.
func (r *RPC) RestoreAll(sync bool) (bool, StdError) {
	method := ARCHIVE + "restoreAll"

	data, stdErr := r.call(method, sync)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// QueryArchiveExist checks if the given snapshot has been archived.
func (r *RPC) QueryArchiveExist(filterID string) (bool, StdError) {
	method := ARCHIVE + "queryArchiveExist"

	data, stdErr := r.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// Pending returns all pending snapshot requests in ascend sort.
func (r *RPC) Pending() ([]SnapshotEvent, StdError) {
	method := ARCHIVE + "pending"

	data, stdErr := r.call(method)
	if stdErr != nil {
		return nil, stdErr
	}

	var result []SnapshotEvent
	if sysErr := json.Unmarshal(data, result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return result, nil
}

/*---------------------------------- cert ----------------------------------*/

// GetTCert 获取TCert
// Deprecated:
func (r *RPC) GetTCert(index uint) (string, StdError) {
	return r.hrm.getTCert(r.hrm.nodes[index].url)
}

/*---------------------------------- account ----------------------------------*/

// GetBalance 获取账户余额
func (r *RPC) GetBalance(account string) (string, StdError) {
	account = chPrefix(account)
	method := ACCOUNT + "getBalance"
	param := account
	data, err := r.call(method, param)
	if err != nil {
		return "", err
	}

	var balance string
	if sysErr := json.Unmarshal(data, &balance); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return balance, nil
}
