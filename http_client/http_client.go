package http_client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type HTTPResponse struct {
	Status  int
	Headers http.Header
	Body    []byte
}

type HttpClient struct {
	httpHttpClient *http.Client
}

func getDefaultTransport() *http.Transport {
	trans := http.DefaultTransport.(*http.Transport).Clone()

	trans.MaxIdleConns = 100
	trans.MaxConnsPerHost = 100
	trans.MaxIdleConnsPerHost = 100

	return trans
}

func NewHTTPClient(timeoutSeconds int, c *http.Client) *HttpClient {
	if c == nil {
		c = &http.Client{
			Timeout:   time.Second * time.Duration(timeoutSeconds),
			Transport: getDefaultTransport(),
		}
	}

	return &HttpClient{
		httpHttpClient: c,
	}
}

func (c *HttpClient) Request(reqType, url string, payloads []byte, headers *http.Header) (*HTTPResponse, error) {
	h := http.Header{
		"Content-Type": []string{"application/json"},
		"User-Agent":   []string{"Clockify-Reporter"},
		"X-Api-Key":    []string{viper.GetString("clockify.apiKey")},
	}

	if headers != nil {
		for key, val := range *headers {
			h[key] = val
		}
	}

	reqNeedPayloads := map[string]bool{
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}
	reqType = strings.ToUpper(reqType)
	var reqPayloads io.Reader = nil
	if reqNeedPayloads[reqType] && payloads != nil {
		reqPayloads = bytes.NewBuffer(payloads)
	}

	httpReq, err := http.NewRequest(reqType, url, reqPayloads)
	if err != nil {
		return nil, err
	}

	httpReq.Header = h

	resp, err := c.httpHttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		Status:  resp.StatusCode,
		Headers: resp.Header,
		Body:    body,
	}, nil
}

func (c *HttpClient) Head(url string, headers *http.Header) (*HTTPResponse, error) {
	h := &http.Header{
		"Cache-Control": []string{"no-cache"},
	}

	if headers != nil {
		h = headers
	}

	r, err := c.Request("HEAD", url, nil, h)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *HttpClient) Get(url string, headers *http.Header) (*HTTPResponse, error) {
	h := &http.Header{
		"Cache-Control": []string{"no-cache"},
	}

	if headers != nil {
		h = headers
	}

	r, err := c.Request("GET", url, nil, h)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *HttpClient) Post(url string, payloads []byte, headers *http.Header) (*HTTPResponse, error) {
	h := &http.Header{
		"Content-Type": []string{"application/json"},
	}

	if headers != nil {
		h = headers
	}

	r, err := c.Request("POST", url, payloads, h)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *HttpClient) Put(url string, payloads []byte, headers *http.Header) (*HTTPResponse, error) {
	h := &http.Header{
		"Content-Type": []string{"application/json"},
	}

	if headers != nil {
		h = headers
	}

	r, err := c.Request("PUT", url, payloads, h)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *HttpClient) Patch(url string, payloads []byte, headers *http.Header) (*HTTPResponse, error) {
	h := &http.Header{
		"Content-Type": []string{"application/json"},
	}

	if headers != nil {
		h = headers
	}

	r, err := c.Request("PATCH", url, payloads, h)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *HttpClient) Delete(url string, payloads []byte, headers *http.Header) (*HTTPResponse, error) {
	h := &http.Header{
		"Content-Type": []string{"application/json"},
	}

	if headers != nil {
		h = headers
	}

	r, err := c.Request("DELETE", url, payloads, h)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c HttpClient) AsJSON(body []byte) (map[string]interface{}, error) {
	var jbody map[string]interface{}
	err := json.Unmarshal(body, &jbody)
	if err != nil {
		return nil, err
	}

	return jbody, err
}
