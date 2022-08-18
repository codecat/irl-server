package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codecat/go-libs/log"
)

type Response struct {
	Result string `json:"result"`
}

func writeError(w http.ResponseWriter, code int, format string, args ...interface{}) {
	log.Error(format, args...)

	res, _ := json.Marshal(Response{
		Result: fmt.Sprintf(format, args...),
	})
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(res)
}

func writeResponse(w http.ResponseWriter, obj interface{}) {
	res, err := json.Marshal(obj)
	if err != nil {
		writeError(w, 500, "Unable to marshal response object: %s", err.Error())
		return
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	w.Write(res)
}

func writeOK(w http.ResponseWriter) {
	writeResponse(w, Response{
		Result: "OK",
	})
}
