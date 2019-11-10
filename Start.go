package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/JojiiOfficial/SystemdGoService"

	"github.com/mkideal/cli"
)

type startT struct {
	cli.Helper
}

var startCMD = &cli.Command{
	Name:    "start",
	Aliases: []string{"restart", "rest", "sta", "star", "s"},
	Desc:    "(re)starts the server",
	Argv:    func() interface{} { return new(startT) },
	Fn: func(ct *cli.Context) error {
		err := setDeamonStatus(SystemdGoService.Restart)
		if err != nil {
			fmt.Println("Error:", err.Error())
		} else {
			fmt.Println("Restarted successfully")
		}
		return nil
	},
}

func setDeamonStatus(status SystemdGoService.SystemdCommand) error {
	if os.Getuid() != 0 {
		return errors.New("you need to be root")
	}
	if !SystemdGoService.SystemfileExists(SystemdGoService.NameToServiceFile(serviceName)) {
		return errors.New("service doesn't exists")
	}

	err := SystemdGoService.SetServiceStatus(serviceName, status)
	if err != nil {
		return err
	}
	return nil
}
