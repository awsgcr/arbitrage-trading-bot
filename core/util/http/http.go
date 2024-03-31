package http

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/net/context/ctxhttp"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var netTransport = &http.Transport{
	//DialContext: (&net.Dialer{
	//	Timeout:   10 * time.Second,
	//	KeepAlive: 60 * time.Second,
	//}).DialContext,
	TLSHandshakeTimeout:   5 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
}

var NetClient = &http.Client{
	Timeout:   time.Second * 10,
	Transport: netTransport,
}

func Get(uri string, params Params) ([]byte, error) {
	return GetWithMultipleValues(uri, params.ToUrlValues())
}

func GetWithMultipleValues(uri string, params url.Values) ([]byte, error) {
	uri = uri + "?" + params.Encode()
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 == 2 {
		return body, nil
	}
	return nil, fmt.Errorf(string(body))
}

func Post(uri string, params url.Values) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := ctxhttp.Do(context.Background(), NetClient, request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 == 2 {
		return body, nil
	}
	return nil, fmt.Errorf(string(body))
}

func PostJson(url, jsonBody string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(jsonBody)))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")

	resp, err := ctxhttp.Do(context.Background(), NetClient, request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		// flushing the body enables the transport to reuse the same connection
		io.Copy(ioutil.Discard, resp.Body)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 == 2 {
		return body, nil
	}
	return nil, fmt.Errorf(string(body))
}
