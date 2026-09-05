package l7

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/iamonah/loadbalancer/config"
	"github.com/iamonah/loadbalancer/strategy"
	"github.com/rs/zerolog/log"
)

type Server struct {
	url     *url.URL
	proxy   *httputil.ReverseProxy
	matcher string
}

func (s *Server) forward(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

type ServerPool struct {
	serviceName string
	servers     []*Server
	Strategy    strategy.Strategy
}

func NewServerPool(svcCfg *config.Service) (*ServerPool, error) {
	servers := make([]*Server, 0, len(svcCfg.Replicas))

	if len(svcCfg.Replicas) == 0 {
		return nil, fmt.Errorf("No replicas defined for service %s", svcCfg.Name)
	}
	for _, replica := range svcCfg.Replicas {
		parsedURL, err := url.Parse(replica)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse URL: %w", err)
		}
		proxy := &httputil.ReverseProxy{
			Rewrite: func(pr *httputil.ProxyRequest) {
				pr.SetURL(parsedURL)
				req := pr.Out

				req.Header.Del("X-Forwarded-For")
				req.Header.Del("X-Forwarded-Host")
				req.Header.Del("X-Forwarded-Proto")
				req.Header.Del("X-Internal-Secret")
				req.Header.Del("Server")

				clientIP, _, err := net.SplitHostPort(pr.In.RemoteAddr)

				if err == nil {
					req.Header.Set("X-Forwarded-For", clientIP)
				} else {
					req.Header.Set("X-Forwarded-For", pr.In.RemoteAddr)
				}

				req.Header.Set("X-Forwarded-Host", pr.In.Host)

				if pr.In.TLS != nil {
					req.Header.Set("X-Forwarded-Proto", "https")
				} else {
					req.Header.Set("X-Forwarded-Proto", "http")
				}

				log.Info().Msgf(
					"L7 Proxy routing request: %s %s to %s",
					req.URL.Path,
					req.URL.Host,
					req.URL.Scheme,
				)
			},
		}

		server := &Server{
			url:     parsedURL,
			proxy:   proxy,
			matcher: svcCfg.Matcher,
		}
		servers = append(servers, server)
	}

	strategy, err := strategy.NewStrategy(
		strategy.StrategyConfig{Type: strategy.StrategyType(svcCfg.Strategy), Weights: svcCfg.Weights},
		uint32(len(svcCfg.Replicas)),
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to create strategy: %w", err)
	}
	return &ServerPool{serviceName: svcCfg.Name, servers: servers, Strategy: strategy}, nil
}

func (sp *ServerPool) GetNextServer() *Server {
	nextIndex := sp.Strategy.NextServer()
	return sp.servers[nextIndex]
}
