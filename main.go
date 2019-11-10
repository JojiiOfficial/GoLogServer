package main

import (
	"fmt"
	"os"

	"github.com/mkideal/cli"
)

var showTimeInLog = false
var logPrefix = ""
var config Config
var configFile = "config.json"
var serviceName = "gologgerserver.service"

type argT struct {
	cli.Helper
}

var root = &cli.Command{
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Usage: gologservec <install/start/stop/run> []")
		return nil
	},
}

func main() {
	if err := cli.Root(root,
		cli.Tree(runCMD),
		cli.Tree(installCMD),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
