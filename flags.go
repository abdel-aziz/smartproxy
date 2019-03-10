package main

import "github.com/codegangsta/cli"

// Main flags

var typeFlag = cli.StringFlag{
	Name:  "type, t",
	Value: "dynamic",
	Usage: "Proxy type (dynamic, tor, http, direct)",
}

var addrFlag = cli.StringFlag{
	Name:  "addr",
	Value: "127.0.0.1:9999",
	Usage: "Listening address (hostname:port)",
}

var torAddr = cli.StringSliceFlag{
	Name:  "tor",
	Value: &cli.StringSlice{"127.0.0.1:9050"},
	Usage: "List of tor address (hostname:port)",
}

var proxiesFlag = cli.StringFlag{
	Name:  "proxies",
	Value: "proxies.txt",
	Usage: "File list of http proxies",
}

var debugFlag = cli.BoolFlag{
	Name:  "D",
	Usage: "Debug mode",
}

var envFlag = cli.StringFlag{
	Name:  "E",
	Value: "dev",
	Usage: "Execution environment (dev, test, prod)",
}

// Logs flags

var logAddrFlag = cli.StringFlag{
	Name:  "logaddr",
	Value: "",
	Usage: "Address (host:port) of logs service",
}

var logUserFlag = cli.StringFlag{
	Name:  "loguser",
	Value: "",
	Usage: "User of logs service",
}

var logPwdFlag = cli.StringFlag{
	Name:  "logpwd",
	Value: "",
	Usage: "Password of logs service",
}
