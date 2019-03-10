package tcp

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"git.blueboard.it/blueboard/proxy/common"
	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
)

type Handler struct {
	Type string
	tls  bool
}

// ServeHTTP is the HTTP handler function in charge of the proxy logic.
// For each incoming client's request a new request is created, with user-agent set randomly, and transport defined according to the main options (dynamic, tor proxies, http proxies or direct).
// Then the request is made and retried if necessary (response code other than 200).
func (h Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// TODO: proper http error on return

	logFields := log.Fields{"layer": "http"}

	// Headers
	logFields["method"] = req.Method
	if !strings.Contains(req.Host, ":") {
		req.Host += ":80"
	}
	domain, _, err := net.SplitHostPort(req.Host)
	if err != nil {
		logFields["error"] = err
		logFields["error_type"] = "request_host"
		log.WithFields(logFields).Errorln()
		return
	}
	logFields["domain"] = domain
	if h.tls {
		// The URI through the CONNECT tunnel seems not providing the sheme and host
		req.RequestURI = "https://" + domain + req.RequestURI
	}
	logFields["uri"] = req.RequestURI

	// Check Tor Blacklist
	if (h.Type == "dynamic" || h.Type == "tor") && common.CheckTorBlacklist(domain) {
		if h.Type == "tor" {
			err = errors.New("Proxy: Tor blacklisted for this domain")
			logFields["error"] = err.Error()
			logFields["error_type"] = "client_tor_blacklist"
			log.WithFields(logFields).Errorln()
			return
		} else if h.Type == "dynamic" {
			h.Type = "http"
		}
	}

	// Create request
	newReq, err := http.NewRequest(req.Method, req.RequestURI, req.Body)
	if err != nil {
		logFields["error"] = err
		logFields["err_type"] = "response_create_request"
		log.WithFields(logFields).Errorln()
		return
	}
	for name, headers := range req.Header {
		if name == "Proxy-Connection" {
			continue
		}
		for _, header := range headers {
			newReq.Header.Add(name, header)
		}
	}

	var ip string
	var resp *http.Response
	var data []byte
	proxyType := h.Type
	tries := 0
	for i := 0; i < 6; i++ {
		tries++
		logFields["try"] = tries

		// Get ip
		ip, logFields["ip"] = common.GetIp(proxyType)

		// Get user agent
		newReq.Header.Set("User-Agent", common.GetUserAgent())
		logFields["user_agent"] = newReq.Header.Get("User-Agent")

		// Get transport
		var transport *http.Transport
		logFields["type"] = proxyType
		if proxyType == "dynamic" {
			logFields["type"] = "tor"
		}
		switch proxyType {
		case "tor", "dynamic":
			transport, err = torConn(ip)
		case "http":
			transport, err = proxyConn(ip)
		default:
			transport, err = directConn(ip)
		}
		if err != nil {
			logFields["error"] = err
			logFields["error_type"] = "transport"
			log.WithFields(logFields).Errorln()
			return
		}

		// Make request
		client := http.Client{Transport: transport}
		resp, err = client.Do(newReq)
		if err != nil {
			logFields["error"] = err
			logFields["err_type"] = "response_do"
			log.WithFields(logFields).Errorln()
			return
		}
		defer resp.Body.Close()
		logFields["status"] = resp.StatusCode

		// Read response
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logFields["err"] = err
			logFields["error_type"] = "response_read"
			log.WithFields(logFields).Errorln()
			return
		}

		// Switch transport ?
		// On dynamic mode, switch to Http proxies after 4 failures with Tor proxies
		if tries == 4 && proxyType == "dynamic" {
			proxyType = "http"
		}

		// Retry ?
		var retry bool
		logFields["captcha"] = false
		switch resp.StatusCode {
		case 403, 503:
			retry = true
		default:
			// Check Amazon CAPTCHA
			if strings.Contains(domain, "amazon") {
				buf := bytes.NewBuffer(data)
				title := strings.ToLower(common.GetTitle(buf))
				if strings.Contains(title, "captcha") {
					logFields["captcha"] = true
					retry = true
				}
			}
		}
		if retry {
			logFields["error_type"] = "response_status"
			log.WithFields(logFields).Warnln()
		} else {
			break
		}
	}

	// Save response
	//filename := fmt.Sprintf("/var/data/proxy/%d_%s_%s.txt", resp.StatusCode, domain, uuid.NewV4().String())
	filename := fmt.Sprintf("log/%d_%s_%s.txt", resp.StatusCode, domain, uuid.NewV4().String())

	if err = ioutil.WriteFile(filename, data, 0644); err != nil {
		logFields["err"] = err
		logFields["error_type"] = "response_write_file"
		log.WithFields(logFields).Errorln()
		return
	}
	logFields["error_file"] = filename

	// Copy headers response
	for headerName, headerValues := range resp.Header {
		for _, headerValue := range headerValues {
			w.Header().Add(headerName, headerValue)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy body response
	fmt.Fprint(w, string(data))

	// Logs
	switch resp.StatusCode {
	case 403, 503:
		log.WithFields(logFields).Errorln()
	default:
		delete(logFields, "error_type")
		log.WithFields(logFields).Infoln()
	}
}
