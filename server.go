package localtunnel

import (
	"io"
	"net/http"
	"strings"
)

var domain string
var secure bool

type createdResp struct {
	ID      string `json:"id,omitempty"`
	URL     string `json:"url,omitempty"`
	Port    int    `json:"port,omitempty"`
	MaxConn int    `json:"max_conn_count,omitempty"`
}

type statusResponse struct {
	Tunnels int `json:"tunnels"`
}

type infoResponse struct {
	Info string `json:"info"`
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	host := getProto() + domain
	info := &infoResponse{Info: "localtunnel server running on " + host}
	io.WriteString(w, toJSON(info))
}

func handleNew(w http.ResponseWriter, r *http.Request) {
	id := randID()
	proxy := NewProxy(id)
	proxy.setup()

	resp := &createdResp{
		ID:      id,
		URL:     getProto() + id + "." + domain,
		Port:    proxy.port,
		MaxConn: 10,
	}

	go proxy.listen()
	io.WriteString(w, toJSON(resp))
}

// SetupServer creates main HTTP server
func SetupServer(port, serverDomain string, isSecure bool) *http.Server {
	domain = serverDomain
	secure = isSecure

	server := &http.Server{
		Addr: ":" + port,
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
					hostname := r.Host
					id := strings.Split(hostname, ".")[0]

					proxy, ok := proxies[id]
					if !ok {
						handleInfo(w, r)
					} else {
						proxy.askNewClientConnection(w, r)
					}
				}
			}
		}),
	}

	return server
}

func getProto() string {
	if secure {
		return "https://"
	}

	return "http://"
}
