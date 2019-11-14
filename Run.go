package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mkideal/cli"
	"github.com/theckman/go-ipdata"
	"github.com/thecodeteam/goodbye"
)

type runT struct {
	cli.Helper
}

var ipdataClient *ipdata.Client

var runCMD = &cli.Command{
	Name:    "run",
	Aliases: []string{},
	Desc:    "run the server service",
	Argv:    func() interface{} { return new(runT) },
	Fn: func(ct *cli.Context) error {
		exit, config := checkConfig(configFile)
		if exit {
			os.Exit(1)
			return nil
		}

		ctx := context.Background()
		defer goodbye.Exit(ctx, -1)
		goodbye.Notify(ctx)
		goodbye.Register(func(ctx context.Context, sig os.Signal) {
			if db != nil {
				_ = db.Close()
				LogInfo("DB closed")
			}
		})

		showTimeInLog = config.ShowTimeInLog

		initDB(config)
		useTLS := false
		if len(config.CertFile) > 0 {
			_, err := os.Stat(config.CertFile)
			if err != nil {
				LogError("Certfile not found. HTTP only!")
				useTLS = false
			} else {
				useTLS = true
			}
		}

		if len(config.KeyFile) > 0 {
			_, err := os.Stat(config.KeyFile)
			if err != nil {
				LogError("Keyfile not found. HTTP only!")
				useTLS = false
			}
		}

		router := NewRouter()
		if useTLS {
			go (func() {
				if config.TLSPort < 2 {
					LogError("TLS port must be bigger than 1")
					os.Exit(1)
				}
				if config.TLSPort == config.HTTPPort {
					LogCritical("HTTP port can't be the same as TLS port!")
					os.Exit(1)
				}
				tlsprt := strconv.Itoa(config.TLSPort)
				log.Fatal(http.ListenAndServeTLS(":"+tlsprt, config.CertFile, config.KeyFile, router))
				LogInfo("Server started TLS on port (" + tlsprt + ")")
			})()
		}
		initAutoDeleteTimer(*config)
		if config.HTTPPort < 2 {
			LogError("HTTP port must be bigger than 1")
			os.Exit(1)
			return nil
		}
		httpprt := strconv.Itoa(config.HTTPPort)
		LogInfo("Server started HTTP on port (" + httpprt + ")")
		log.Fatal(http.ListenAndServe(":"+httpprt, router))
		return nil
	},
}

func checkConfig(configFile string) (exit bool, config *Config) {
	_, err := os.Stat(configFile)
	if err != nil {
		err = createConfig(configFile)
		if err != nil {
			LogCritical("Couldn't create config: " + err.Error())
		} else {
			LogInfo("Created config sucessfully")
		}
		return true, nil
	}
	config = readConfig(configFile)
	if len(config.Host) == 0 || len(config.Pass) == 0 {
		LogCritical("You need to fill the config")
		return true, nil
	}
	return false, config
}

func initAutoDeleteTimer(config Config) {
	if config.DeleteLogsAfter == 0 {
		return
	}
	LogInfo("Deleting logs after " + strconv.Itoa(config.DeleteLogsAfter) + "h")
	go (func() {
		timer := time.Tick(1 * time.Hour)
		for {
			minTime := time.Now().Unix() - int64(config.DeleteLogsAfter*3600)
			_, err := db.Exec("DELETE FROM SystemdLog WHERE date < ?", minTime)
			if err != nil {
				LogError("Error deleting old logs: " + err.Error())
			} else {
				LogInfo("Deleted old logs")
			}
			<-timer
		}
	})()
}
