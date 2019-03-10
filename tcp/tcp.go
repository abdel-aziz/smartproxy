package tcp

import (
	"fmt"
	"io"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

const BUF_SIZE = 54 * 1024 // 64 ko

const (
	RL            = "\r\n"
	EOL           = RL + RL
	HTTP_VERSION  = "HTTP/1.1 "
	CONNECT_OK    = HTTP_VERSION + "200 Connection established" + EOL
	CONNECT_ERROR = HTTP_VERSION + "502 Bad Gateway" + EOL
	CLIENT_ERROR  = HTTP_VERSION + "400 Bad Request" + EOL
)

type Proxy struct {
	Type             string
	ConnClt, connSrv net.Conn
	received         int
	sent             int
	err              error
	logFields        log.Fields
}

// Start read the first bytes of an HTTP request at the TCP level, and identify if the request is with TLS (method CONNECT) or not (other methods).
// Then the request is upgraded at the HTTP level and done through the 2 internal servers (with and without TLS) accordingy.
// Read/writes between the TCP and the HTTP level is done by the upgrader through bi-directional pipes on those streams.
func (p *Proxy) Start() {
	p.logFields = log.Fields{"layer": "tcp"}
	defer p.ConnClt.Close()

	// Get request
	msg := make([]byte, BUF_SIZE)
	nb, err := p.ConnClt.Read(msg)
	if err != nil {
		// If there is an error at this level the client request is empty
		p.logFields["error"] = err
		p.logFields["error_type"] = "request_first"
		log.WithFields(p.logFields).Debugln()
		return
	}
	req, err := getRequest(msg[:nb])
	if err != nil {
		p.logFields["error"] = err
		p.logFields["error_type"] = "request_to_http"
		log.WithFields(p.logFields).Errorln()
		return
	}

	// Handle request
	if err = p.upgrade(req, msg[:nb], req.Method == "CONNECT"); err != nil {
		p.logFields["error"] = err
		log.WithFields(p.logFields).Errorln()
		if req.Method == "CONNECT" {
			_, err = p.ConnClt.Write([]byte(CONNECT_ERROR))
		}
		return
	}
}

type TcpResponseWriter struct {
	header http.Header
	conn   io.Writer
}

func (w *TcpResponseWriter) Header() http.Header {
	return w.header
}

func (w *TcpResponseWriter) WriteHeader(code int) {
	w.Write([]byte(fmt.Sprintf("HTTP/1.1 %d%s", code, RL))) // TODO status msg
	for headerName, headerValues := range w.header {
		for _, headerValue := range headerValues {
			w.Write([]byte(fmt.Sprintf("%s: %s%s", headerName, headerValue, RL)))
		}
	}
	w.Write([]byte("Transfer-Encoding: chunked" + RL))
	w.Write([]byte(RL))
}

func (w *TcpResponseWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}
