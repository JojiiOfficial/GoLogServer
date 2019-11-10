package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/JojiiOfficial/SystemdGoService"

	"github.com/mkideal/cli"
)

type installT struct {
	cli.Helper
}

var installCMD = &cli.Command{
	Name:    "install",
	Aliases: []string{},
	Desc:    "installs the deamon",
	Argv:    func() interface{} { return new(installT) },
	Fn: func(ct *cli.Context) error {
		if os.Getuid() != 0 {
			fmt.Println("You neet to be root!")
			return nil
		}

		exit, _ := checkConfig(configFile)
		if exit {
			os.Exit(1)
			return nil
		}

		if SystemdGoService.SystemfileExists(serviceName) {
			fmt.Println("Service already exists")
			return nil
		}

		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		service := SystemdGoService.NewDefaultService(serviceName, "Centralized loggingservice easy and simple", ex+" run")
		service.Service.User = "root"
		service.Service.Group = "root"
		service.Service.Restart = SystemdGoService.Always
		cpath, _ := filepath.Abs(ex)
		cpath, _ = path.Split(cpath)
		service.Service.WorkingDirectory = cpath
		service.Service.RestartSec = "3"
		err = service.Create()
		if err == nil {
			err := service.Enable()
			if err != nil {
				LogCritical("Couldn't enable service: " + err.Error())
				return nil
			}
			err = service.Start()
			if err != nil {
				LogCritical("Couldn't start service: " + err.Error())
				return nil
			}
			fmt.Println("Service installed and started")
		} else {
			fmt.Println("An error occured installitg the service: ", err.Error())
		}

		return nil
	},
}
