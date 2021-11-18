package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	// load config.conf
	config, err := LoadConfig("config.conf")
	if err != nil {
		logrus.WithError(err).Fatal("config error")
		os.Exit(1)
	}

	//create database
	err = StorageConnect(&config.Storage)
	if err != nil {
		logrus.Error(err)
		logrus.WithError(err).Fatal("storage error")
		os.Exit(1)
	}

	// init routes
	router := GenerateRouter()

	logrus.Info(fmt.Sprintf("server listening on port: %d", config.Port))

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", config.Host, config.Port), router)
	if err != nil {
		logrus.WithError(err).Fatal("server error")
		os.Exit(1)
	}
}

func GenerateRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/scan", NewScan).Methods(http.MethodPost)
	r.HandleFunc("/scan/{id}", GetScan).Methods(http.MethodGet)
	r.HandleFunc("/scan", GetAll).Methods(http.MethodGet)

	return r
}
