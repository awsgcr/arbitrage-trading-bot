package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SecType int

const (
	SecTypeNone SecType = iota
	SecTypeAPIKey
	SecTypeSigned // if the 'timestamp' parameter is required
)

type params map[string]interface{}

// Request define an API request
type Request struct {
	Method     string
	Endpoint   string
	query      url.Values
	form       url.Values
	recvWindow int64
	SecType    SecType
	header     http.Header
	body       io.Reader
	fullURL    string
}

// addParam add param with key/value to query string
func (r *Request) addParam(key string, value interface{}) *Request {
	if r.query == nil {
		r.query = url.Values{}
	}
	r.query.Add(key, fmt.Sprintf("%v", value))
	return r
}

// SetParam set param with key/value to query string
func (r *Request) SetParam(key string, value interface{}) *Request {
	if r.query == nil {
		r.query = url.Values{}
	}
	r.query.Set(key, fmt.Sprintf("%v", value))
	return r
}

// setParams set params with key/values to query string
func (r *Request) setParams(m params) *Request {
	for k, v := range m {
		r.SetParam(k, v)
	}
	return r
}

// setFormParam set param with key/value to Request form body
func (r *Request) setFormParam(key string, value interface{}) *Request {
	if r.form == nil {
		r.form = url.Values{}
	}
	r.form.Set(key, fmt.Sprintf("%v", value))
	return r
}

// setFormParams set params with key/values to Request form body
func (r *Request) setFormParams(m params) *Request {
	for k, v := range m {
		r.setFormParam(k, v)
	}
	return r
}

func (r *Request) validate() (err error) {
	if r.query == nil {
		r.query = url.Values{}
	}
	if r.form == nil {
		r.form = url.Values{}
	}
	return nil
}

// RequestOption define option type for Request
type RequestOption func(*Request)

// WithRecvWindow set recvWindow param for the Request
func WithRecvWindow(recvWindow int64) RequestOption {
	return func(r *Request) {
		r.recvWindow = recvWindow
	}
}

// WithHeader set or add a header value to the Request
func WithHeader(key, value string, replace bool) RequestOption {
	return func(r *Request) {
		if r.header == nil {
			r.header = http.Header{}
		}
		if replace {
			r.header.Set(key, value)
		} else {
			r.header.Add(key, value)
		}
	}
}

// WithHeaders set or replace the headers of the Request
func WithHeaders(header http.Header) RequestOption {
	return func(r *Request) {
		r.header = header.Clone()
	}
}
