package localtunnel

import (
	"errors"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/apex/log"
	"github.com/dustin/go-humanize"
)

// Socket is a TCP tunnel established between the client and server
type Socket struct {
	id            string
	conn          net.Conn
	inUse         bool
	done          chan bool
	sentBytes     uint64
	receivedBytes uint64
}

// Proxy holds data of the Proxy TCP Server
type Proxy struct {
	id      string
	server  net.Listener
	port    int
	sockets []*Socket
	logger  log.Interface
	mux     sync.Mutex
}

var proxies = make(map[string]*Proxy)

// NewProxy creates new Proxy instance
func NewProxy(id string) *Proxy {
	logContext := log.WithFields(log.Fields{
		"proxyID": id,
	})
	p := &Proxy{id: id, logger: logContext}
	proxies[id] = p
	return p
}

func (p *Proxy) setup() {
	listener, err := net.Listen("tcp4", ":0")
	if err != nil {
		p.logger.Errorf("error starting server on %s", listener.Addr().String())
		return
	}

	p.logger.Infof("tcp server listening on %s", listener.Addr().String())

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
	s := &Socket{id: randID(), conn: conn, done: make(chan bool)}
	p.sockets = append(p.sockets, s)
	p.logger.WithField("socketID", s.id).Infof("new tcp connection %s <> %s", conn.RemoteAddr().String(), conn.LocalAddr().String())

	// defer s.conn.Close()
	// notify := make(chan error)
	// go func() {
	// 	buf := make([]byte, 1024)
	// 	for {
	// 		_, err := conn.Read(buf)
	// 		if err != nil {
	// 			notify <- err
	// 			return
	// 		}
	// 	}
	// }()

	// err := <-notify
	// if io.EOF == err {
	// 	p.cleanUpSocket(s)
	// }
}

func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	socket, err := p.getSocket()
	if err != nil {
		p.logger.Errorf("error finding available tcp connection: %s", err)
		return
	}

	socket.inUse = true
	websocket := isWebSocketRequest(r)

	if websocket {
		p.logger.WithField("socketID", socket.id).Infof("websocket request: %s %s", r.Method, r.URL)
	} else {
		p.logger.WithField("socketID", socket.id).Infof("request: %s %s", r.Method, r.URL)
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Not a hijacker?", 500)
		return
	}
	nc, _, err := hj.Hijack()
	if err != nil {
		p.logger.Errorf("hijack error %s", err)
		return
	}
	defer nc.Close()
	defer socket.conn.Close()

	err = r.Write(socket.conn)
	if err != nil {
		p.logger.WithField("socketID", socket.id).Errorf("error copying request to target: %s", err)
		return
	}

	go p.pipe(socket.conn, nc, socket)
	go p.pipe(nc, socket.conn, socket)

	<-socket.done
}

func (p *Proxy) pipe(src, dst io.ReadWriter, socket *Socket) {
	isOut := dst == socket.conn

	buf := make([]byte, 0xffff)
	for {
		n, err := src.Read(buf)
		if err != nil {
			socket.done <- true
			return
		}
		b := buf[:n]

		n, err = dst.Write(b)
		if err != nil {
			socket.done <- true
			return
		}

		if isOut {
			socket.sentBytes += uint64(n)
		} else {
			socket.receivedBytes += uint64(n)
		}

		p.logger.WithField("socketID", socket.id).Debugf("%d bytes transferred", uint64(n))
	}
}

func (p *Proxy) getSocket() (*Socket, error) {
	for i := range p.sockets {
		if !p.sockets[i].inUse {
			return p.sockets[i], nil
		}
	}

	return nil, errors.New("socket not found")
}

func (p *Proxy) cleanUpSocket(socket *Socket) {
	p.mux.Lock()
	for i, s := range p.sockets {
		if s == socket {
			s.conn.Close()
			p.sockets = append(p.sockets[:i], p.sockets[i+1:]...)
			p.logger.WithField("socketID", s.id).Warnf("socket closed. sent bytes: %d (%s) received bytes: %d (%s)", s.sentBytes, humanize.Bytes(s.sentBytes), s.receivedBytes, humanize.Bytes(s.receivedBytes))
			continue
		}
	}

	if len(p.sockets) == 0 {
		p.server.Close()
		p.logger.Warn("tcp server closed")
		delete(proxies, p.id)
	}
	p.mux.Unlock()
}
