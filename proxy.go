package localtunnel

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
)

type connection struct {
	conn  net.Conn
	resp  chan []byte
	inUse bool
}

// Proxy holds data of Proxy instance
type Proxy struct {
	id          string
	server      net.Listener
	port        int
	connections []*connection
}

var proxies = make(map[string]*Proxy)

// NewProxy creates new Proxy instance
func NewProxy(id string) *Proxy {
	p := &Proxy{id: id}
	proxies[id] = p
	return p
}

func (p *Proxy) setup() {
	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		fmt.Printf("Error starting TCP server on %s\n", listener.Addr().String())
		return
	}

	fmt.Printf("[%s] TCP server listening on %s\n", p.id, listener.Addr().String())

	p.server = listener
	p.port = listener.Addr().(*net.TCPAddr).Port
}

func (p *Proxy) listen() {
	for {
		conn, err := p.server.Accept()
		if err != nil {
			break
		}

		go p.handleConnection(conn)
	}
}

func (p *Proxy) handleConnection(conn net.Conn) {
	c := &connection{conn: conn, resp: make(chan []byte)}
	p.connections = append(p.connections, c)
	fmt.Printf("[%s] New connection %s <> %s\n", p.id, conn.RemoteAddr().String(), conn.LocalAddr().String())
}

func (p *Proxy) handleWebSocketRequest(w http.ResponseWriter, r *http.Request, destConn net.Conn) {
	fmt.Printf("[%s] Request (WebSocket) %s\n", p.id, r.URL)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Not a hijacker?", 500)
		return
	}
	nc, _, err := hj.Hijack()
	if err != nil {
		fmt.Printf("[%s] hijack error: %s\n", p.id, err)
		return
	}
	defer nc.Close()
	defer destConn.Close()

	err = r.Write(destConn)
	if err != nil {
		fmt.Printf("[%s] error copying request to target: %s\n", p.id, err)
		return
	}
	errc := make(chan error, 2)
	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}
	go cp(destConn, nc)
	go cp(nc, destConn)
	<-errc
}

func (p *Proxy) handleHTTPRequest(w http.ResponseWriter, r *http.Request, c *connection) {
	fmt.Printf("[%s] Request %s\n", p.id, r.URL)

	go func() {
		for {
			var buf bytes.Buffer
			io.Copy(&buf, c.conn)

			if buf.Len() == 0 {
				fmt.Printf("[%s] Closing connection %s <> %s\n", p.id, c.conn.RemoteAddr().String(), c.conn.LocalAddr().String())
				c.conn.Close()
				c.conn = nil
				p.cleanupConnections()
			}

			c.resp <- buf.Bytes()
		}
	}()

	method := r.Method
	path := r.URL.EscapedPath()

	c.conn.Write([]byte(method + " " + path + "\r\n\r\n"))

	payload := <-c.resp
	splitted := regexp.MustCompile(`\r\n\r\n`).Split(string(payload), 2)
	headers, body := splitted[0], splitted[1]

	splittedHeaders := strings.SplitN(headers, "\r\n", 2)
	_, headers = splittedHeaders[0], splittedHeaders[1]

	for _, header := range strings.Split(headers, "\r\n") {
		split := strings.Split(header, ":")
		w.Header().Set(strings.TrimSpace(split[0]), strings.TrimSpace(split[1]))
	}

	io.WriteString(w, body)
}

func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	c, err := p.getConnection()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	c.inUse = true
	destConn := c.conn

	if isWebSocketRequest(r) {
		p.handleWebSocketRequest(w, r, destConn)
	} else {
		p.handleHTTPRequest(w, r, c)
	}
}

func (p *Proxy) getConnection() (*connection, error) {
	for i := range p.connections {
		if !p.connections[i].inUse {
			return p.connections[i], nil
		}
	}

	return nil, errors.New("connection not found")
}

func (p *Proxy) cleanupConnections() {
	for i, conn := range p.connections {
		if conn.conn == nil {
			p.connections = removeConnection(p.connections, i)
		}
	}

	if len(p.connections) == 0 {
		p.server.Close()
		fmt.Printf("[%s] TCP server closed.\n", p.id)
		delete(proxies, p.id)
	}
}
