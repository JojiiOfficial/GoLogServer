package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mkideal/cli"
	log "github.com/sirupsen/logrus"
	"github.com/theckman/go-ipdata"
)

type runT struct {
	cli.Helper
}

var ipdataClient *ipdata.Client

func startServer(config *Config) {
	useTLS := false
	if len(config.WebserverConfig.CertFile) > 0 {
		_, err := os.Stat(config.WebserverConfig.CertFile)
		if err != nil {
			log.Error("Certfile not found")
			return
		}
		useTLS = true
	}

	if len(config.WebserverConfig.KeyFile) > 0 {
		_, err := os.Stat(config.WebserverConfig.KeyFile)
		if err != nil {
			log.Error("Keyfile not found. HTTP only!")
			useTLS = false
		}
	}

	router := NewRouter()
	var httpServer, httpsServer *http.Server

	//Init http server
	if useTLS {
		httpsServer = &http.Server{
			Handler:      router,
			Addr:         ":" + strconv.Itoa(config.WebserverConfig.TLSPort),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
	}

	httpServer = &http.Server{
		Handler:      router,
		Addr:         ":" + strconv.Itoa(config.WebserverConfig.HTTPPort),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	//Start TLS
	if useTLS {
		if config.WebserverConfig.TLSPort < 2 {
			log.Error("TLS port must be bigger than 1")
			os.Exit(1)
		}
		if config.WebserverConfig.TLSPort == config.WebserverConfig.HTTPPort {
			log.Fatalln("HTTP port can't be the same as TLS port!")
			os.Exit(1)
		}

		go (func() {
			log.Fatal(httpsServer.ListenAndServeTLS(config.WebserverConfig.CertFile, config.WebserverConfig.KeyFile))
		})()
	}

	//Start cleaner only if specified
	if *appAutoclean {
		initAutoDeleteTimer(config)
	}

	//Start HTTP
	if config.WebserverConfig.HTTPPort > 2 {
		log.Info("Server started HTTP on port (" + httpServer.Addr + ")")
		go (func() {
			log.Fatal(httpServer.ListenAndServe())
		})()
	}

	awaitExit(db, httpServer, httpsServer)
}

//Shutdown server gracefully
func awaitExit(db *sqlx.DB, httpServer, httpTLSserver *http.Server) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, os.Interrupt, syscall.SIGKILL, syscall.SIGTERM)

	// await os signal
	<-signalChan

	// Create a deadline for the await
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	log.Info("Shutting down server")

	if httpServer != nil {
		httpServer.Shutdown(ctx)
		log.Info("HTTP server shutdown complete")
	}

	if httpTLSserver != nil {
		httpTLSserver.Shutdown(ctx)
		log.Info("HTTPs server shutdown complete")
	}

	if db != nil && db.DB != nil {
		db.Close()
		log.Info("Database shutdown complete")
	}

	log.Info("Shutting down complete")
	os.Exit(0)
}

func initAutoDeleteTimer(config *Config) {
	if config.DeleteLogInterval == 0 {
		return
	}

	log.Info("Deleting logs after " + config.DeleteLogInterval.String() + "h")

	go (func() {
		timer := time.Tick(config.DeleteLogInterval)
		for {
			cleanUp(config)
			<-timer
		}
	})()
}

//Cleanup old logs
func cleanUp(config *Config) {
	minTime := time.Now().Unix() - int64(config.DeleteLogInterval.Seconds())
	_, err := db.Exec("DELETE FROM SystemdLog WHERE date < ?", minTime)
	if err != nil {
		log.Error("Error deleting old systemdlogs: " + err.Error())
		return
	}

	_, err = db.Exec("DELETE FROM CustomLog WHERE date < ?", minTime)
	if err != nil {
		log.Error("Error deleting old logs: " + err.Error())
		return
	}

	log.Info("Deleted old logs")

	_, err = db.Exec("DELETE FROM Message WHERE pk_id not in (SELECT message FROM SystemdLog) and pk_id not in (SELECT message FROM CustomLog)")
	if err != nil {
		log.Error("Error deleting unused messages: " + err.Error())
		return
	}

	log.Info("Deleted unused messages")
}
