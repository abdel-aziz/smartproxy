package tcp

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func getRequest(msg []byte) (*http.Request, error) {
	buf := bytes.NewBuffer(msg)
	req, err := http.ReadRequest(bufio.NewReader(buf))
	if err != nil {
		return req, err
	}
	if !strings.Contains(req.Host, ":") {
		req.Host += ":80"
	}
	return req, nil
}

func (p *Proxy) upgrade(req *http.Request, msg []byte, tls bool) error {
	var err error

	// Dial server
	port := "1080"
	if tls {
		port = "10443"
	}
	p.connSrv, err = net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		p.logFields["err_type"] = "response_dial"
		return err
	}
	defer p.connSrv.Close()

	// First bytes
	if tls {
		// If tls don't send headers to internal server,
		// but initial body if some
		if err = p.initialBody(msg); err != nil {
			p.logFields["err_type"] = "response_initial_body"
			return err
		}
	} else {
		// If not tls send first bytes to internal server,
		// including original headers and initial body if some
		if _, err = p.connSrv.Write(msg); err != nil {
			p.logFields["err_type"] = "response_headers"
			return err
		}
	}

	// Pipes
	errs := make(chan error)
	go p.pipe(p.connSrv, p.ConnClt, errs)
	if tls {
		_, err = p.ConnClt.Write([]byte(CONNECT_OK))
		if err != nil {
			p.logFields["err_type"] = "request_tunnel_ok"
			return err
		}
	}
	go p.pipe(p.ConnClt, p.connSrv, errs)

	// Wait for connections to end
	err = <-errs
	if err != nil && err != io.EOF {
		p.logFields["err_type"] = "tunnel_pipe"
		return err
	}

	return nil
}

func Server(proxyType string) {
	handler := Handler{Type: proxyType}
	server := &http.Server{
		Addr:    ":1080",
		Handler: handler,
		// Disable HTTP/2
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func TlsServer(proxyType string) {
	handler := Handler{Type: proxyType, tls: true}
	server := &http.Server{
		Addr:    ":10443",
		Handler: handler,
		// Disable HTTP/2
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	err := server.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
}
