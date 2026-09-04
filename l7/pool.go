package l7

import (
	"net/http/httputil"
	"net/url"

			"github.com/iamonah/loadbalancer/l7/strategy"
)

type Server struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

type ServerPool struct {
	servers  []*Server
	Strategy strategy.Strategy
}
