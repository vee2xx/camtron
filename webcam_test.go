package camtron

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// func TestBinaryDownload(t *testing.T) {
// 	downloadBinary("camtron-linux-x64")
// 	assert.FileExists(t, "camtron-linux-x64.zip")
// }

// func TestUnzip(t *testing.T) {
// 	UnzipBinary("camtron-linux-x64")
// 	assert.DirExists(t, "camtron-linux-x64")
// }

//Use to test end to end. Requires manual intervention
func TestRunElectron(t *testing.T) {

	consumers = make(map[string]StreamConsumer)
	var vidToFileStream = make(chan []byte, 10)
	options := make(map[string]string)
	options["filePath"] = "testvid.webm"
	options["maxSize"] = "100000000"
	streamConsumer := StreamConsumer{Stream: vidToFileStream, Context: make(chan string), Handler: StreamToFile, Options: options}
	consumers["file"] = streamConsumer

	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	go StartWebcamUI()

	for _, consumer := range consumers {
		go consumer.Handler(consumer.Stream, consumer.Context, consumer.Options)
	}

	go ConsumeStream()

	go StartServer()

	time.Sleep(30 * time.Second)

	StopWebcamUI()

	assert.FileExists(t, "testvid.webm")
	_ = os.Remove("testvid.webm")

}

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/endStreaming", shutdownStream)
	mux.HandleFunc("/streamVideo", streamVideo)
	mux.HandleFunc("/log", handleLogging)
	mux.HandleFunc("/uploadImage", uploadImage)
	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}

func TestCleanUp(t *testing.T) {
	_ = os.Remove("camtron-linux-x64.zip")
	_ = os.RemoveAll("camtron-linux-x64")
	_ = os.Remove("camtron.log")
}
