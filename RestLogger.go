package main

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

//Logger logs stuff
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info(r.Method + " " + r.RequestURI + " " + name)
		start := time.Now()
		inner.ServeHTTP(w, r)
		dur := time.Since(start)
		if dur > 500*time.Millisecond && dur < 5*time.Second {
			log.Info("Duration: " + dur.String())
		}
	})
}
