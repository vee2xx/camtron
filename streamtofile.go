package camtron

import (
	"log"
	"os"
	"strconv"
	"time"
)

func StreamToFile(streamChan chan []byte, context chan string, options map[string]string) {

	filePath := options["filePath"]
	if filePath == "" {
		filePath = "vid-" + time.Now().Format("2006_01_02_15_04_05") + "." + videoMetaData.Container
	}

	maxSize, err := strconv.Atoi(options["maxSize"])
	if err != nil {
		maxSize = 1000000000
	}

	var data []byte
	for {
		select {
		case packet, ok := <-streamChan:
			if !ok {
				log.Print("WARNING: Failed to get packet")
			}
			if len(data) > 1000 {
				if !writeVideoToFile(filePath, data, maxSize) {
					return
				}
				data = nil
			}
			data = append(data, packet...)
		case val, _ := <-context:
			log.Println("got signal " + val)
			if val == "stop" {
				writeVideoToFile(filePath, data, maxSize)
				close(streamChan)
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
