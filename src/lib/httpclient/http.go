package httpclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/glog"
)

// HTTPPost do http post by url and param
func HTTPPost(url string, json []byte) (*http.Response, error) {
	// glog.Info(url)
	// glog.Info(bytes.NewBuffer(json))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("client.do err: %v", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		glog.Errorf("statusCode != 200: %v", resp)
		return nil, err
	}
	return resp, nil
}

// Client defines the interface for go client that wants to connect to a hyperchain RPC endpoint
type client interface {
	Send(header http.Header, body []byte) ([]byte, error)
	Close()
}

// httpClient connects to a hyperchain RPC server over HTTP.
type httpClient struct {
	endpoint *url.URL     // HTTP-RPC server endpoint
	client   *http.Client // reuse connection
}

// NewHTTPClient create a new RPC client that connection to
// a hyperchain RPC server over HTTP.
func newHTTPClient(endpoint string, timeout time.Duration) (client, error) {
	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: timeout,
	}
	return &httpClient{endpoint: url, client: client}, nil
}

// Send will serialize the given req to JSON and sends it to the RPC server.
// If receive response with statusOK(200), return []byte of response body.
func (c *httpClient) Send(header http.Header, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", c.endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = header

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return ioutil.ReadAll(resp.Body)
	}

	return nil, fmt.Errorf("http failed: %s", resp.Status)
}

// Close is not necessary for httpClient
func (c *httpClient) Close() {
}
