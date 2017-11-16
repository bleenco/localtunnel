package vex

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

type createdResp struct {
	ID      string `json:"id,omitempty"`
	URL     string `json:"url,omitempty"`
	Port    int    `json:"port,omitempty"`
	MaxConn int    `json:"max_conn_count,omitempty"`
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
			params := r.URL.Query()
			if _, ok := params["new"]; ok {
				handleNew(w, r)
			} else {
				fmt.Printf("Request %s\n", r.URL)
				proxy := proxies["abc123"]
				proxy.handleRequest(w, r)
			}
		}),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	return server
}

// package vex

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// )

// type server struct {
// 	logger *log.Logger
// 	mux    *http.ServeMux
// }

// type createdResp struct {
// 	ID      string `json:"id,omitempty"`
// 	URL     string `json:"url,omitempty"`
// 	Port    int    `json:"port,omitempty"`
// 	MaxConn int    `json:"max_conn_count,omitempty"`
// }

// func (s *server) indexHandler(c http.ResponseWriter, r *http.Request) {
// 	params := r.URL.Query()
// 	if _, ok := params["new"]; ok {
// 		// id := randID()
// 		id := "abc123"
// 		proxy := NewProxy(id)
// 		proxy.setup()

// 		resp := &createdResp{
// 			ID:      id,
// 			URL:     id + ".jan",
// 			Port:    proxy.port,
// 			MaxConn: 10,
// 		}

// 		go proxy.listen()
// 		io.WriteString(c, toJSON(resp))
// 	} else {
// 		fmt.Printf("Request %s\n", r.URL)

// 		ids := make([]string, len(proxies))
// 		i := 0
// 		for k := range proxies {
// 			ids[i] = k
// 			i++
// 		}

// 		proxy := proxies[ids[0]]
// 		proxy.handleRequest(c, r)
// 	}
// }

// func newServer(options ...func(*server)) *server {
// 	s := &server{
// 		mux: http.NewServeMux(),
// 	}

// 	for _, f := range options {
// 		f(s)
// 	}

// 	if s.logger == nil {
// 		s.logger = log.New(os.Stdout, "", 0)
// 	}

// 	s.mux.HandleFunc("/", s.indexHandler)

// 	return s
// }

// func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	s.mux.ServeHTTP(w, r)
// }

// // SetupServer creates http server and correlated logger
// func SetupServer(addr string) (*http.Server, *log.Logger) {
// 	logger := log.New(os.Stdout, "", 0)
// 	s := newServer(func(s *server) {
// 		s.logger = logger
// 	})

// 	hs := &http.Server{Addr: addr, Handler: s}
// 	return hs, logger
// }
