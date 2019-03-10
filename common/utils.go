package common

import (
	"bytes"
	"errors"
	"math/rand"
	"strings"
	"time"
)

var (
	HTTP_PROXIES []string
	TOR_PROXIES  []string
	RANDOM       = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func EndHeaders(msg []byte) (int, error) {
	var index int
	for _, delem := range [][]byte{[]byte("\r\n\r\n"), []byte("\n\n")} {
		index = bytes.Index(msg, delem)
		if index != -1 {
			index += len(delem)
			break
		}
	}
	if index == -1 {
		return index, errors.New("No headers")
	}
	return index, nil
}

func CheckTorBlacklist(host string) bool {
	for _, domain := range torBlacklist {
		if strings.Contains(host, domain) {
			return true
		}
	}
	return false
}

func GetHttpProxy() string {
	return HTTP_PROXIES[RANDOM.Intn(len(HTTP_PROXIES))]
}

func GetTorProxy() string {
	return TOR_PROXIES[RANDOM.Intn(len(TOR_PROXIES))]
}

func GetIp(proxyType string) (string, string) {
	var ip, ipLog string
	switch proxyType {
	case "tor", "dynamic":
		ip = GetTorProxy()
		ipLog = ip
	case "http":
		ip = GetHttpProxy()
		ipLog = strings.Split(ip, "@")[1]
	}
	return ip, ipLog
}

func GetUserAgent() string {
	return userAgents[RANDOM.Intn(len(userAgents))]
}
