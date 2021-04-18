package camtron

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
)

var vidStream = make(chan []byte, 10)

type StreamConsumer struct {
	Stream  chan []byte
	Context chan string
	Options map[string]string
	Handler func(chan []byte, chan string, map[string]string)
}

func ConsumeStream(vidStream chan []byte, consumers map[string]StreamConsumer) {
	for {
		packet, ok := <-vidStream
		if !ok {
			log.Print("bad packet!")
		}
		for _, consumer := range consumers {
			consumer.Stream <- packet
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func streamVideo(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	defer conn.Close()

	for {

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Print("ERROR:" + err.Error())
			return
		}
		vidStream <- msg
	}
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	var p string

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := base64.StdEncoding.DecodeString(p)
	currentTime := time.Now()
	//TODO: Get file name from request
	var filename = currentTime.Format("2006_01_02_15_04_05_000000") + ".png"
	ioutil.WriteFile(filename, data, 0644)

}

func Shellout(shell string, args ...string) error {
	cmd := exec.Command(shell, args...)
	err := cmd.Start()
	return err
}

func StartElectron() {
	log.Println("INFO: starting electron")

	goos := runtime.GOOS

	_, filename, _, ok := runtime.Caller(0)

	if !ok {
		panic("No caller information")
	}

	goPath := path.Dir(filename)

	var shell string
	var args []string
	switch goos {
	case "windows":
		shell = "cmd"
		args = append(args, "/C")
		args = append(args, "cd "+goPath+"/camtron-win32-x64 && camtron.exe")
	case "darwin":
		shell = "bash"
		args = append(args, "-c")
		args = append(args, "cd "+goPath+"/camtron-darwin-x64 && open camtron")
	case "linux":
		shell = "bash"
		args = append(args, "-c")
		args = append(args, "cd "+goPath+"/camtron-linux-x64 && ./camtron")
	default:
		log.Println("Unsupported OS: %s.\n", goos)
	}

	log.Print("Starting Electron")

	err := Shellout(shell, args...)
	if err != nil {
		log.Print(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		select {
		case sig := <-c:
			if goos == "windows" {
				log.Println("Got %s signal. Its windows so gotta kill Electron\n", sig)
				cmd := exec.Command("cmd", "/C", " taskkill /f /im camtron.exe")
				err := cmd.Run()
				if err != nil {
					log.Println("Couldn't kill camtron")
				}
			} else if goos == "darwin" {
				log.Println("Got %s signal. Its MacOs so gotta kill Electron\n", sig)
				cmd := exec.Command("bash", "-c", " killall camtron")
				err := cmd.Run()
				if err != nil {
					log.Println("Couldn't kill camtron")
				}
			}
			os.Exit(0)
		}
	}()

	if err != nil {
		log.Printf("error: %v\n", err)
	}
}

func StartCam(consumers map[string]StreamConsumer) {
	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	go StartElectron()

	for _, consumer := range consumers {
		go consumer.Handler(consumer.Stream, consumer.Context, consumer.Options)
	}
	go ConsumeStream(vidStream, consumers)

	mux := http.NewServeMux()
	mux.HandleFunc("/streamVideo", streamVideo)
	mux.HandleFunc("/log", handleLogging)
	mux.HandleFunc("/uploadImage", uploadImage)
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
