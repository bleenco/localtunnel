package vex

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
)

type connection struct {
	conn  net.Conn
	resp  chan []byte
	inUse bool
}

// Proxy holds data of Proxy instance
type Proxy struct {
	id          string
	started     bool
	server      net.Listener
	port        int
	connections []*connection
}

var proxies = make(map[string]*Proxy)

// NewProxy creates new Proxy instance
func NewProxy(id string) *Proxy {
	p := &Proxy{
		id:      id,
		started: false,
	}

	proxies[id] = p
	return p
}

func (p *Proxy) setup() {
	if p.started {
		fmt.Printf("Proxy %s already started\n", p.id)
	}

	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		fmt.Printf("Error starting TCP server on %s\n", listener.Addr().String())
		return
	}

	fmt.Printf("[%s] TCP server listening on %s\n", p.id, listener.Addr().String())

	p.started = true
	p.server = listener
	p.port = listener.Addr().(*net.TCPAddr).Port
}

func (p *Proxy) listen() {
	defer func() {
		p.server.Close()
		fmt.Printf("[%s] TCP server closed.", p.id)
	}()

	for {
		conn, err := p.server.Accept()
		if err != nil {
			fmt.Printf("[%s] Error accepting socket connection: %s", p.id, err)
		}

		go p.handleConnection(conn)
	}
}

func (p *Proxy) handleConnection(conn net.Conn) {
	c := &connection{conn: conn, resp: make(chan []byte)}
	p.connections = append(p.connections, c)
	fmt.Printf("[%s] New connection from %s\n", p.id, conn.LocalAddr().String())

	go func() {
		defer c.conn.Close()

		for {
			res, err := ioutil.ReadAll(conn)
			if err != nil {
				fmt.Printf("Error: %s", err)
			}
			c.resp <- res
		}
	}()
}

func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	c, err := p.getConnection()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	c.inUse = true
	destConn := c.conn

	method := r.Method
	path := r.URL.EscapedPath()

	destConn.Write([]byte(method + " " + path + "\r\n\r\n"))

	payload := <-c.resp
	splitted := regexp.MustCompile(`\r\n\r\n`).Split(string(payload), 2)
	_, body := splitted[0], splitted[1]

	// splittedHeaders := strings.SplitN(headers, "\r\n", 2)
	// statusStr, headers := splittedHeaders[0], splittedHeaders[1]

	// for _, header := range strings.Split(headers, "\r\n") {
	// 	split := strings.Split(header, ":")
	// 	w.Header()[split[0]] = strings.Split(split[1], "")
	// }

	io.WriteString(w, body)
}

func (p *Proxy) getConnection() (*connection, error) {
	for i := range p.connections {
		if !p.connections[i].inUse {
			return p.connections[i], nil
		}
	}

	return nil, errors.New("connection not found")
}
