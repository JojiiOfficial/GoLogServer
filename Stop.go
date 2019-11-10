package main

import (
	"fmt"

	"github.com/JojiiOfficial/SystemdGoService"

	"github.com/mkideal/cli"
)

var stopCMD = &cli.Command{
	Name:    "stop",
	Aliases: []string{"stop"},
	Desc:    "stops the server",
	Fn: func(ct *cli.Context) error {
		err := setDeamonStatus(SystemdGoService.Stop)
		if err != nil {
			fmt.Println("Error:", err.Error())
		} else {
			fmt.Println("Stopped successfully")
		}
		return nil
	},
}
