package tcp

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

func torConn(ip string) (*http.Transport, error) {
	var transport *http.Transport
	torUrl, err := url.Parse("socks5://" + ip)
	if err != nil {
		return transport, err
	}
	tor, err := proxy.FromURL(torUrl, proxy.Direct)
	if err != nil {
		return transport, err
	}
	transport = &http.Transport{
		Dial: tor.Dial,
		// Disable HTTP/2
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
	return transport, nil
}

func proxyConn(ip string) (*http.Transport, error) {
	var transport *http.Transport
	proxyUrl, err := url.Parse("http://" + ip)
	if err != nil {
		return transport, err
	}
	transport = &http.Transport{
		Dial:  net.Dial,
		Proxy: http.ProxyURL(proxyUrl),
		// Disable HTTP/2
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
	return transport, nil
}

func directConn(ip string) (*http.Transport, error) {
	transport := &http.Transport{
		Dial: net.Dial,
		// Disable HTTP/2
		TLSNextProto: make(map[string]func(string, *tls.Conn) http.RoundTripper),
	}
	return transport, nil
}
