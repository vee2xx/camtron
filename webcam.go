package camtron

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var vidStream chan []byte

var videoMetaData VideoMetaData

var consumers map[string]StreamConsumer

var continueStreaming bool = false

type VideoMetaData struct {
	Container string
	Codec     string
}

type StreamConsumer struct {
	Stream  chan []byte
	Context chan string
	Options map[string]string
	Handler func(chan []byte, chan string, map[string]string)
}

func ConsumeStream() {
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

	for continueStreaming {

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
	var filename = currentTime.Format("2006_01_02_15_04_05_000000") + ".png"
	ioutil.WriteFile(filename, data, 0644)

}

func initiatializeStream(w http.ResponseWriter, r *http.Request) {
	log.Println("got init signal")
	err := json.NewDecoder(r.Body).Decode(&videoMetaData)
	if err != nil {
		videoMetaData = VideoMetaData{Container: "webm", Codec: "v9"}
	}
	handleStreamStart()
}

func handleStreamStart() {
	vidStream = make(chan []byte, 10)
	continueStreaming = true
}

func shutdownStream(w http.ResponseWriter, r *http.Request) {
	log.Println("got shutdown signal")
	handleStreamShutdown()
}

func handleStreamShutdown() {
	for _, consumer := range consumers {
		consumer.Context <- "stop"
	}
	continueStreaming = false
}

func Shellout(shell string, args ...string) error {
	cmd := exec.Command(shell, args...)
	err := cmd.Start()
	return err
}

func getLatestUIVersion() (string, error) {
	url := "https://api.github.com/repos/vee2xx/camtron-ui/releases"

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err.Error())
	}

	var releases []map[string]interface{}
	if err := json.Unmarshal(body, &releases); err != nil {
		log.Fatal(err)
	}

	if len(releases) == 0 {
		return "", errors.New("unable to find any versions")
	}
	return fmt.Sprintf("%v", releases[0]["tag_name"]), nil
}

func downloadBinary(electronBinary string) {
	latest, err := getLatestUIVersion()
	if err != nil {
		log.Fatal(err.Error())
	}
	filename := electronBinary + ".zip"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		url := fmt.Sprintf("https://github.com/vee2xx/camtron-ui/releases/download/%s/%s", latest, filename)

		file, err := http.Get(url)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer file.Body.Close()

		out, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}

		defer out.Close()

		_, err = io.Copy(out, file.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func UnzipBinary(source string) {

	r, err := zip.OpenReader(source + ".zip")

	if err != nil {
		log.Fatal(err)
	}

	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(".", f.Name)

		if fpath == source {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if !strings.HasPrefix(fpath, filepath.Clean(source)+string(os.PathSeparator)) {
			log.Fatalf("%s is an illegal filepath", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			log.Fatal(err)
		}
		outFile, err := os.OpenFile(fpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			f.Mode())
		if err != nil {
			log.Fatal(err)
		}

		rc, err := f.Open()

		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

}

func StartElectron() {
	log.Println("INFO: starting electron")

	goos := runtime.GOOS

	var shell string
	var args []string
	var electronBinary string
	switch goos {
	case "windows":
		shell = "cmd"
		args = append(args, "/C")
		args = append(args, "cd camtron-win32-x64 && camtron.exe")
		electronBinary = "camtron-win32-x64"
	case "darwin":
		shell = "bash"
		args = append(args, "-c")
		args = append(args, "cd camtron-darwin-x64 && open camtron.app")
		electronBinary = "camtron-darwin-x64"
	case "linux":
		shell = "bash"
		args = append(args, "-c")
		args = append(args, "cd camtron-linux-x64 && ./camtron")
		electronBinary = "camtron-linux-x64"
	default:
		log.Fatalf("Unsupported OS: %s.\n", goos)
	}

	if _, err := os.Stat(electronBinary); os.IsNotExist(err) {
		downloadBinary(electronBinary)
		UnzipBinary(electronBinary)
	}

	log.Print("Starting Electron")

	err := Shellout(shell, args...)
	if err != nil {
		log.Print(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		defer os.Exit(0)
		select {
		case sig := <-c:
			if goos == "windows" {
				log.Printf("Got %s signal. Its windows so gotta kill Electron\n", sig)
				cmd := exec.Command("cmd", "/C", " taskkill /f /im camtron.exe")
				err := cmd.Run()
				if err != nil {
					log.Println("Couldn't kill camtron")
				}
			} else if goos == "darwin" {
				log.Printf("Got %s signal. Its MacOs so gotta kill Electron\n", sig)
				cmd := exec.Command("bash", "-c", " killall camtron")
				err := cmd.Run()
				if err != nil {
					log.Println("Couldn't kill camtron")
				}
			}
			handleStreamShutdown()
		}
	}()

	if err != nil {
		log.Printf("error: %v\n", err)
	}
}

func StartCam(consumerList map[string]StreamConsumer) {
	consumers = consumerList
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
	go ConsumeStream()

	mux := http.NewServeMux()
	mux.HandleFunc("/initializeStream", initiatializeStream)

	mux.HandleFunc("/streamVideo", streamVideo)
	mux.HandleFunc("/log", handleLogging)
	mux.HandleFunc("/uploadImage", uploadImage)
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
