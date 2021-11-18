package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

func NewScan(w http.ResponseWriter, r *http.Request) {
	logrus.Info(fmt.Sprintf("[%s] %s", r.Method, r.URL.Path))

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Fatal("couldn't read body")
		ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	var newScan Scan
	err = json.Unmarshal(body, &newScan)
	if err != nil {
		logrus.WithError(err).Error("missing value")
		ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	newScan.ID = uuid.New().String()
	path := "/tmp/src/" + newScan.ID
	newScan.Status = StatusProcess
	newScan.CreatedAt = time.Now()

	err = DB.Set(&newScan)
	if err != nil {
		ReturnError(w, http.StatusBadRequest, err.Error())
		return
	}

	go func() {
		// clone the repository
		err := GitClone(newScan.URL, path)
		if err != nil {
			newScan.Error = err.Error()
			newScan.Status = StatusError

			if err = DB.Set(&newScan); err != nil {
				logrus.WithError(err).Error("Could not save to db")
			}

			return
		}

		// create docker service
		docker, err := NewDockerService("opensorcery/bandit", []string{"-f", "json", "-r", "/code"})
		if err != nil {
			newScan.Error = err.Error()
			newScan.Status = StatusError

			if err = DB.Set(&newScan); err != nil {
				logrus.WithError(err).Error("Could not save to db")
			}
			return
		}

		// run the docker service
		result, err := docker.Run(path)
		if err != nil {
			newScan.Error = err.Error()
			newScan.Status = StatusError

			if err = DB.Set(&newScan); err != nil {
				logrus.WithError(err).Error("Could not save to db")
			}

			return
		}

		newScan.Result = *result
		newScan.Status = StatusDone

		// check security criteria
		// if total SEVERITY.HIGH more then 1 result, make unsecure, otherwise no rating or comment

		metrics := newScan.Result["metrics"]
		totals := metrics.(map[string]interface{})["_totals"]
		severity := totals.(map[string]interface{})["SEVERITY.HIGH"]

		if severity.(float64) > 1 {
			newScan.IsSecure = false
		}

		// save results
		if err = DB.Set(&newScan); err != nil {
			logrus.WithError(err).Error("Could not save to db")
		}
	}()

	ReturnResponse(w, http.StatusOK, map[string]string{"id": newScan.ID})
}

func GetScan(w http.ResponseWriter, r *http.Request) {
	logrus.Info(fmt.Sprintf("[%s] %s", r.Method, r.URL.Path))

	id := mux.Vars(r)["id"]

	data, err := DB.Get(id)
	if err != nil || data == nil {
		ReturnError(w, http.StatusNotFound, "not found")
		return
	}

	ReturnResponse(w, http.StatusOK, data)
}

func GetAll(w http.ResponseWriter, r *http.Request) {
	logrus.Info(fmt.Sprintf("[%s] %s", r.Method, r.URL.Path))

	result := DB.Keys()

	ReturnResponse(w, http.StatusOK, result)
}

func ReturnResponse(w http.ResponseWriter, statusCode int, resp interface{}) {
	bytes, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, string(bytes))
}

func ReturnError(w http.ResponseWriter, statusCode int, errMsg string) {
	resp := map[string]interface{}{
		"error": errMsg,
	}

	bytes, _ := json.Marshal(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = fmt.Fprintf(w, string(bytes))
}
