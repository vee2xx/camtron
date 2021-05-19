package camtron

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

var logfile string = "camtron.log"

type LogEntry struct {
	LogLevel string
	Message  string
}

func handleLogging(w http.ResponseWriter, r *http.Request) {

	var logEntry LogEntry
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &logEntry)

	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	log.Print(logEntry.LogLevel + " " + logEntry.Message)
}

func RegisterStream(stream chan []byte) {
	streams = append(streams, stream)
}
