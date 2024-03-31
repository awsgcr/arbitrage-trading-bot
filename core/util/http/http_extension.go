package http

import "net/url"

type Params map[string]string

func (p Params) ToUrlValues() url.Values {
	values := url.Values{}
	for k, v := range p {
		values.Set(k, v)
	}
	return values
}
