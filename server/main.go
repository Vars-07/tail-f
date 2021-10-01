package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Infof("Booting up logger server")

	m := NewManager()
	router := m.GetRouter()

	// Can we served from flag
	filePath := "logger.log"
	go m.fileWatch(filePath)

	log.Fatal(http.ListenAndServe(":8080", router))
}
