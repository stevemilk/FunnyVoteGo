package rpc

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperchain/gosdk/common"
	"github.com/terasum/viper"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Node is used to contain node info
type Node struct {
	url    string
	wsURL  string
	status bool
}

func newNode(url string, rpcPort string, wsPort string, isHTTPS bool) (node *Node) {
	var scheme string

	if isHTTPS {
		scheme = "https://"
	} else {
		scheme = "http://"
	}

	node = &Node{
		url:    scheme + url + ":" + rpcPort,
		wsURL:  "ws://" + url + ":" + wsPort,
		status: true,
	}
	return node
}

// HTTPRequestManager is used to manager node and http request
type HTTPRequestManager struct {
	nodes     []*Node
	nodeIndex int
	client    *http.Client
	namespace string
	sendTcert bool
	tcm       *TCertManager
}

// newHTTPRequestManager is used to construct HTTPRequestManager
func newHTTPRequestManager(vip *viper.Viper, confRootPath string) (hrm *HTTPRequestManager) {
	var (
		namespace string
		urls      []string
		rpcPorts  []string
		wsPorts   []string
		isHTTPS   bool
		client    *http.Client
		tcm       *TCertManager
	)

	namespace = vip.GetString(common.NamespaceConf)

	urls = vip.GetStringSlice(common.JSONRPCNodes)
	logger.Debugf("[CONFIG]: %s = %v", common.JSONRPCNodes, urls)

	rpcPorts = vip.GetStringSlice(common.JSONRPCPorts)
	logger.Debugf("[CONFIG]: %s = %v", common.JSONRPCPorts, rpcPorts)

	wsPorts = vip.GetStringSlice(common.WebSocketPorts)
	logger.Debugf("[CONFIG]: %s = %v", common.WebSocketPorts, wsPorts)

	isHTTPS = vip.GetBool(common.SecurityHttps)
	logger.Debugf("[CONFIG]: %s = %v", common.SecurityHttps, isHTTPS)

	var nodes = make([]*Node, len(urls))

	for i, url := range urls {
		nodes[i] = newNode(url, rpcPorts[i], wsPorts[i], isHTTPS)
	}

	if isHTTPS {
		pool := x509.NewCertPool()

		tlscaPath := strings.Join([]string{confRootPath, vip.GetString(common.SecurityTlsca)}, "/")
		tlspeerCertPath := strings.Join([]string{confRootPath, vip.GetString(common.SecurityTlspeerCert)}, "/")
		tlspeerCertPrivPath := strings.Join([]string{confRootPath, vip.GetString(common.SecurityTlspeerPriv)}, "/")

		caCrt, err := ioutil.ReadFile(tlscaPath)
		if err != nil {
			panic(fmt.Sprintf("read tlsCA from %s failed", tlscaPath))
		}

		pool.AppendCertsFromPEM(caCrt)

		cliCrt, err := tls.LoadX509KeyPair(tlspeerCertPath, tlspeerCertPrivPath)
		if err != nil {
			panic(fmt.Sprintf("read tlspeerCert from %s and %s failed", tlspeerCertPath, tlspeerCertPrivPath))
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{cliCrt},
				ServerName:   "hyperchain.cn",
			},
		}

		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	sendTcert := vip.GetBool(common.PrivacySendTcert)
	logger.Debugf("[CONFIG]: sendTcert = %v", sendTcert)
	if sendTcert {
		tcm, _ = NewTCertManager(vip, confRootPath)
	}

	httpRequestManager := &HTTPRequestManager{
		nodes:     nodes,
		nodeIndex: 0,
		client:    client,
		namespace: namespace,
		sendTcert: vip.GetBool(common.PrivacySendTcert),
		tcm:       tcm,
	}

	if sendTcert && !vip.GetBool(common.Cfca) {
		tcm.tcertPool = make(map[string]TCert)
		for _, node := range nodes {
			tcert, err := httpRequestManager.getTCert(node.url)
			if err != nil {
				logger.Error("can not get tcert from ", node.url, err)
				return
			}
			tcm.tcertPool[node.url] = TCert(tcert)
		}
	}

	return httpRequestManager
}

func post(url string, body []byte) (*http.Request, StdError) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	return req, NewGetResponseError(err)
}

// SyncRequest function is used to send http request
func (hrm *HTTPRequestManager) SyncRequest(body []byte) ([]byte, StdError) {
	randomURL, stdErr := hrm.randomURL()
	if stdErr != nil {
		return nil, stdErr
	}

	return hrm.SyncRequestSpecificURL(body, randomURL)
}

// SyncRequestSpecificURL is used to post request to specific url
func (hrm *HTTPRequestManager) SyncRequestSpecificURL(body []byte, url string) ([]byte, StdError) {
	req, stdErr := post(url, body)
	if stdErr != nil {
		return nil, stdErr
	}

	if hrm.sendTcert {
		if hrm.tcm.cfca {
			signature, sysErr := hrm.tcm.sdkCert.Sign(body)
			if sysErr != nil {
				logger.Error("sign error", sysErr)
				return nil, NewSystemError(sysErr)
			}
			req.Header.Add("tcert", hrm.tcm.ecert)
			req.Header.Add("signature", common.Bytes2Hex(signature))
			req.Header.Add("msg", common.Bytes2Hex(body))
		} else {
			signature, err := hrm.tcm.uniqueCert.Sign(body)
			if err != nil {
				logger.Error("signature body error,", err)
				return nil, NewSystemError(err)
			}
			req.Header.Add("tcert", string(hrm.tcm.tcertPool[url]))
			req.Header.Add("signature", common.Bytes2Hex(signature))
			req.Header.Add("msg", common.Bytes2Hex(body))
		}
	}

	//logger.Info("[URL]:", url)
	//logger.Info("[REQUEST]:", string(body))

	resp, sysErr := hrm.client.Do(req)
	if sysErr != nil {
		return nil, NewGetResponseError(sysErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		ret, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, NewSystemError(err)
		}
		//logger.Info("[RESPONSE]:", string(ret))
		return ret, nil
	}

	// 请求异常返回，重连节点
	hrm.ReConnectNode(hrm.nodeIndex)

	return nil, NewGetResponseError(errors.New("http failed " + resp.Status))
}

func (hrm *HTTPRequestManager) getTCert(url string) (string, StdError) {
	rawReq := &JSONRequest{
		Method:    "cert_getTCert",
		Version:   JSONRPCVersion,
		ID:        1,
		Namespace: hrm.namespace,
	}
	uniqPub, sysErr := ioutil.ReadFile(hrm.tcm.uniquePubPath)
	if sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	param := NewMapParam("pubkey", common.Bytes2Hex(uniqPub)).Serialize()
	rawReq.Params = []interface{}{param}

	body, sysErr := json.Marshal(rawReq)
	if sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	req, stdErr := post(url, body)
	if stdErr != nil {
		return "", stdErr
	}

	signature, sysErr := hrm.tcm.sdkCert.Sign(body)
	if sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	req.Header.Add("tcert", hrm.tcm.ecert)
	req.Header.Add("signature", common.Bytes2Hex(signature))
	req.Header.Add("msg", common.Bytes2Hex(body))

	logger.Info("[URL]:", url)
	logger.Info("[REQUEST]:", string(body))

	resp, sysErr := hrm.client.Do(req)
	if sysErr != nil {
		return "", NewGetResponseError(sysErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		ret, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", NewSystemError(err)
		}
		logger.Info("[RESPONSE]:", string(ret))

		var resp *JSONResponse
		if sysErr = json.Unmarshal(ret, &resp); sysErr != nil {
			return "", NewSystemError(sysErr)
		}

		if resp.Code != SuccessCode {
			return "", NewServerError(resp.Code, resp.Message)
		}

		var tcert TCertResponse
		if err := json.Unmarshal(resp.Result, &tcert); err != nil {
			return "", NewSystemError(err)
		}
		return tcert.TCert, nil
	}

	return "", NewGetResponseError(errors.New("http failed " + resp.Status))
}

func (hrm *HTTPRequestManager) randomURL() (url string, err StdError) {
	nodeNum := len(hrm.nodes)
	randomNum := nodeNum * 2
	for randomNum > 0 {
		hrm.nodeIndex = common.RandInt(nodeNum)
		if hrm.nodes[hrm.nodeIndex].status {
			return hrm.nodes[hrm.nodeIndex].url, nil
		}
		randomNum--
	}
	logger.Error("All nodes are bad, please check it! Now retry to connect all nodes.")

	//if random fail, try round
	for i := 0; i < nodeNum; i++ {
		hrm.nodeIndex = (hrm.nodeIndex + 1) % nodeNum
		if hrm.nodes[hrm.nodeIndex].status {
			return hrm.nodes[hrm.nodeIndex].url, nil
		}
	}

	return "", NewGetResponseError(errors.New("all nodes are bad, please check it"))
}

// ReConnectNode is used to reconnect the node by index
func (hrm *HTTPRequestManager) ReConnectNode(nodeIndex int) {
	url := hrm.nodes[nodeIndex].url
	req := &JSONRequest{
		Method:    "node_getNodes",
		Version:   JSONRPCVersion,
		ID:        1,
		Namespace: hrm.namespace,
	}
	body, err := json.Marshal(req)
	if err != nil {
		logger.Error(NewSystemError(err).String())
	}

	go func() {
		request, err := post(url, body)
		if err != nil {
			logger.Error(err.String())
		}

		for true {
			response, err := hrm.client.Do(request)
			if err != nil {
				logger.Error(NewSystemError(err).String())
			}

			if response != nil && response.StatusCode == http.StatusOK {
				b, _ := ioutil.ReadAll(response.Body)
				logger.Debug("reconnection node body: ", string(b))
				response.Body.Close()
				hrm.nodes[nodeIndex].status = true
				logger.Info("node " + hrm.nodes[nodeIndex].url + " Reconnect Success!")
				return
			}
			response.Body.Close()
			logger.Info("node " + hrm.nodes[nodeIndex].url + " Reconnect failed, will try one second later")
			time.Sleep(time.Second)
		}
	}()

}
