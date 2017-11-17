package localtunnel

import (
	"crypto/tls"
	"io"
	"net/http"
)

type createdResp struct {
	ID      string `json:"id,omitempty"`
	URL     string `json:"url,omitempty"`
	Port    int    `json:"port,omitempty"`
	MaxConn int    `json:"max_conn_count,omitempty"`
}

type statusResponse struct {
	Tunnels int `json:"tunnels"`
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	id := "abc123"
	proxy := NewProxy(id)
	proxy.setup()

	resp := &createdResp{
		ID:      id,
		URL:     id + ".jan",
		Port:    proxy.port,
		MaxConn: 10,
	}

	go proxy.listen()
	io.WriteString(w, toJSON(resp))
}

// SetupServer creates main HTTP server
func SetupServer(addr string) *http.Server {
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.EscapedPath()
			if path == "/api/status" {
				resp := &statusResponse{Tunnels: len(proxies)}
				io.WriteString(w, toJSON(resp))
			} else {
				params := r.URL.Query()
				if _, ok := params["new"]; ok {
					handleNew(w, r)
				} else {
					proxy := proxies["abc123"]
					proxy.handleRequest(w, r)
				}
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	return server
}
