package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/thecodeteam/goodbye"
)

var showTimeInLog = false
var logPrefix = ""
var config Config
var configFile = "config.json"

func main() {
	ctx := context.Background()
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(func(ctx context.Context, sig os.Signal) {
		if db != nil {
			_ = db.Close()
			LogInfo("DB closed")
		}
	})

	_, err := os.Stat(configFile)
	if err != nil {
		err = createConfig(configFile)
		if err != nil {
			LogCritical("Couldn't create config: " + err.Error())
		} else {
			LogInfo("Created config sucessfully")
		}
		return
	}

	config = readConfig(configFile)
	showTimeInLog = config.ShowTimeInLog

	initDB(config)
	useTLS := false
	if len(config.CertFile) > 0 {
		_, err = os.Stat(config.CertFile)
		if err != nil {
			LogError("Certfile not found. HTTP only!")
			useTLS = false
		} else {
			useTLS = true
		}
	}

	if len(config.KeyFile) > 0 {
		_, err = os.Stat(config.KeyFile)
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

	if config.HTTPPort < 2 {
		LogError("HTTP port must be bigger than 1")
		return
	}
	httpprt := strconv.Itoa(config.HTTPPort)
	LogInfo("Server started HTTP on port (" + httpprt + ")")
	log.Fatal(http.ListenAndServe(":"+httpprt, router))
}
