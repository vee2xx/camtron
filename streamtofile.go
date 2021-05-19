package camtron

import (
	"log"
	"os"
	"time"
)

var filePath = "./videos/"
var fileName string = filePath + "vid-" + time.Now().Format("2006_01_02_15_04_05") + ".webm"
var maxSize int = 1000000000

func StreamToFile(vidStream chan []byte) {
	if _, err := os.Stat("videos"); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("videos", os.ModePerm)
		}
	}
	var data []byte
	for {
		select {
		case packet, ok := <-vidStream:
			if !ok {
				log.Print("WARNING: Failed to get packet")
			}
			if len(data) > 1000 {
				if !writeVideoToFile(fileName, data, maxSize) {
					return
				}
				data = nil
			}
			data = append(data, packet...)
		case val, _ := <-context:
			log.Println("got signal " + val)
			if val == "stop" {
				writeVideoToFile(fileName, data, maxSize)
				close(vidStream)
				log.Println("INFO: Shutting stream to file")
				data = nil
				return
			}
		}
	}
}
func writeVideoToFile(fileName string, video []byte, maxSize int) bool {
	vidFile, fileOpenErr := os.OpenFile(fileName,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if fileOpenErr != nil {
		log.Println(fileOpenErr)
	}
	defer vidFile.Close()
	fileStat, statErr := vidFile.Stat()

	if statErr != nil {
		log.Println(statErr)
	}

	if fileStat.Size() > int64(maxSize) {
		log.Println("Maximum file size reached")
		return false
	}

	_, writeErr := vidFile.Write(video)
	if writeErr != nil {
		log.Println(writeErr)
		return false
	}

	return true
}

func StartStreamToFileConsumer() {
	vidStream := make(chan []byte, 10)
	RegisterStream(vidStream)
	go StreamToFile(vidStream)
}
