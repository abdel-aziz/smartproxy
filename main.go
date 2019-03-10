package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"git.blueboard.it/blueboard/proxy/common"
	"git.blueboard.it/blueboard/proxy/tcp"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

const VERSION = "0.2.3"

var AUTHORS = []cli.Author{
	{
		Name:  "Aziz Mounir",
		Email: "aziz@blueboard.it",
	},
}

func main() {

	app := cli.NewApp()
	app.Authors = AUTHORS
	app.Version = VERSION
	app.Usage = "BlueBoard Proxy"

	proxyType := "dynamic"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "type, t",
			Value: "dynamic",
			Usage: "Proxy type (dynamic, tor, http, direct)",
		},
		cli.StringFlag{
			Name:  "addr",
			Value: "127.0.0.1:9999",
			Usage: "Listening address (hostname:port)",
		},
		cli.StringSliceFlag{
			Name:  "tor",
			Value: &cli.StringSlice{},
			Usage: "List of tor address (hostname:port)",
		},
		cli.StringFlag{
			Name:  "proxies",
			Value: "proxies.txt",
			Usage: "File list of http proxies",
		},
		cli.BoolFlag{
			Name:  "D",
			Usage: "Debug mode",
		},
	}

	app.Before = func(c *cli.Context) error {

		// Debug
		if c.Bool("D") {
			log.SetLevel(log.DebugLevel)
			log.Debugln("Debug Mode")
		}

		// Proxy type
		proxyType = c.String("type")
		var valid bool
		for _, value := range []string{"dynamic", "tor", "http", "direct"} {
			if proxyType == value {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("Wrong proxy type: %s", proxyType)
		}

		// Listening address
		_, err := net.ResolveTCPAddr("tcp", c.String("addr"))
		if err != nil {
			log.Error("Failed to resolve remote address: %s", err)
			return err
		}

		// Tor proxies
		if proxyType == "dynamic" || proxyType == "tor" {
			for _, addr := range c.StringSlice("tor") {
				_, err := net.ResolveTCPAddr("tcp", addr)
				if err != nil {
					return fmt.Errorf("Failed to resolve tor address: %s", err)
				}
				add := true
				for _, torAddr := range common.TOR_PROXIES {
					if torAddr == addr {
						add = false
						break
					}
				}
				if add {
					common.TOR_PROXIES = append(common.TOR_PROXIES, addr)
				}
			}
			if len(common.TOR_PROXIES) == 0 {
				common.TOR_PROXIES = append(common.TOR_PROXIES, "127.0.0.1:9050")
			}
			log.Infoln("Tor proxies:", common.TOR_PROXIES)
		}

		// Http proxies
		if proxyType == "dynamic" || proxyType == "http" {
			content, err := ioutil.ReadFile(c.String("proxies"))
			if err != nil {
				return err
			}
			scanner := bufio.NewScanner(bytes.NewBuffer(content))
			for scanner.Scan() {
				parts := strings.Split(scanner.Text(), ":")
				addr := fmt.Sprintf("%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1])
				common.HTTP_PROXIES = append(common.HTTP_PROXIES, addr)
			}
			if scanner.Err() != nil {
				return scanner.Err()
			}
		}

		return nil
	}

	app.Action = func(c *cli.Context) {

		// Listen TCP
		// Our main listener, a TCP server
		listener, err := net.Listen("tcp", c.String("addr"))
		if err != nil {
			log.Fatalln(err)
		}
		defer listener.Close()

		// Http servers
		// In parall we launch 2 http servers
		// one for clear requests, the other for tls requests
		go tcp.Server(proxyType)
		go tcp.TlsServer(proxyType)

		// Accept TCP
		for {
			conn, err := listener.Accept()
			if err != nil {
				logFields := log.Fields{"err": err, "error_type": "tcp_accept"}
				log.WithFields(logFields).Errorln()
			}
			p := &tcp.Proxy{ConnClt: conn, Type: proxyType}
			go p.Start()
		}
	}

	app.Run(os.Args)

}
