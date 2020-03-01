package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/mkideal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/theckman/go-ipdata"
	"github.com/thecodeteam/goodbye"
)

type runT struct {
	cli.Helper
}

var ipdataClient *ipdata.Client

func startServer(configFile string) {
	log.Info("Init config")
	config, exit := InitConfig(configFile)
	if exit {
		os.Exit(1)
		return
	}

	ctx := context.Background()
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(func(ctx context.Context, sig os.Signal) {
		if db != nil {
			_ = db.Close()
			log.Info("DB closed")
		}
	})

	log.Info("Init DB")
	initDB(config)

	useTLS := false
	if len(config.WebserverConfig.CertFile) > 0 {
		_, err := os.Stat(config.WebserverConfig.CertFile)
		if err != nil {
			log.Error("Certfile not found. HTTP only!")
			useTLS = false
		} else {
			useTLS = true
		}
	}

	if len(config.WebserverConfig.KeyFile) > 0 {
		_, err := os.Stat(config.WebserverConfig.KeyFile)
		if err != nil {
			log.Error("Keyfile not found. HTTP only!")
			useTLS = false
		}
	}

	router := NewRouter()
	if useTLS {
		go (func() {
			if config.WebserverConfig.TLSPort < 2 {
				log.Error("TLS port must be bigger than 1")
				os.Exit(1)
			}
			if config.WebserverConfig.TLSPort == config.WebserverConfig.HTTPPort {
				log.Fatalln("HTTP port can't be the same as TLS port!")
				os.Exit(1)
			}
			tlsprt := strconv.Itoa(config.WebserverConfig.TLSPort)
			log.Fatal(http.ListenAndServeTLS(":"+tlsprt, config.WebserverConfig.CertFile, config.WebserverConfig.KeyFile, router))
			log.Info("Server started TLS on port (" + tlsprt + ")")
		})()
	}

	initAutoDeleteTimer(*config)
	if config.WebserverConfig.HTTPPort < 2 {
		log.Error("HTTP port must be bigger than 1")
		os.Exit(1)
		return
	}

	httpprt := strconv.Itoa(config.WebserverConfig.HTTPPort)

	log.Info("Server started HTTP on port (" + httpprt + ")")
	log.Fatal(http.ListenAndServe(":"+httpprt, router))
	return
}

func initAutoDeleteTimer(config Config) {
	if config.DeleteLogInterval == 0 {
		return
	}

	log.Info("Deleting logs after " + strconv.Itoa(config.DeleteLogInterval) + "h")

	go (func() {
		timer := time.Tick(1 * time.Hour)
		for {
			<-timer
			minTime := time.Now().Unix() - int64(config.DeleteLogInterval*3600)
			_, err := db.Exec("DELETE FROM SystemdLog WHERE date < ?", minTime)
			if err != nil {
				log.Error("Error deleting old systemdlogs: " + err.Error())
				continue
			}
			_, err = db.Exec("DELETE FROM CustomLog WHERE date < ?", minTime)
			if err != nil {
				log.Error("Error deleting old logs: " + err.Error())
			} else {
				log.Info("Deleted old logs")
			}
			_, err = db.Exec("DELETE FROM Message WHERE pk_id not in (SELECT message FROM SystemdLog) and pk_id not in (SELECT message FROM CustomLog)")
			if err != nil {
				log.Error("Error deleting unused messages: " + err.Error())
			}
		}
	})()
}
