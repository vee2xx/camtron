# Go Webcam Simplified
Camtron is a simple cross platform library written in go to easily have Go code interact with webcams i.e. consume and process a stream from a webcam consistantly across OS's without relying on opencv. It uses Electron and the MediaDevices Web API to access the webcam and allows a variety of consumers to listen for and process the video stream, for example recording a video from a webcam and saving the video to a file or sending the video from a webcam on to one or more endpoints. It is supported on Linux, Windows 10, Macos and Raspberry Pi 4. Currently the only supported codec is VP9. More will be added shortly.

## Install
There are two ways to install Camtron
### Download the module
For Go 1.16 and up turn modules off first

```
go version
output:  go version go1.16.4 linux/amd64
export GO111MODULE=off //Linux or macOS
go env -w GO111MODULE=off //Windows
```
    
Install using 'go get'
```
go get github.com/vee2xx/camtron
```
### Or use Go Modules[https://blog.golang.org/using-go-modules]

Initialize your project as a module
   
```
go mod init yourproject.com/yourmodule
```
    
Add Camtron to the resulting go.mod file
    
```
require (
    github.com/vee2xx/camtron v1.0.8
)
```
    
The first time Camtron runs it will download and unzip the os appropriate camtron-ui package to your project's root directory so that Camtron can find the Electron app binary and execute it.

#### Connecting to the webcam on a Raspberry Pi
If you are trying to run Camtron on a Raspberry Pi you will need to in install the GNOME configuarion database system as it is missing.
```
sudo apt-get install libgconf-2-4
```

You may also have trouble connecting to the webcam. Raspberry Pis can be temperamental and a reboot might do the trick. If it, there is more information here:

https://www.raspberrypi.org/forums/viewtopic.php?t=173181

https://www.raspberrypi.org/forums/viewtopic.php?t=220261

### Record a video and save it to a file with Golang
Create a project add the following code to main.go

```
import (
 "github.com/vee2xx/camtron"
)
StartStreamToFileConsumer() //start a listener that accepts and processes the stream
StartCam() //start the Electron app that connects to the webcam and captures the stream
```

Run main.go in a terminal

```
go run main.go
```

This starts a listener function that accepts the stream and processes it and the Electron app itself which connects to the webcam and captures the stream. The video file is saved to the videos directory in the project root.

### Create a custom handler
1. Register a channel that will recieve the incoming stream
```
myStreamChan := make(chan []byte, 10)
RegisterStream(myStreamChan)
```
2. Create a function with a loop that will check the channel for data and then handle it in some way
```
func MyStreamHandler(myStreamChan chan []byte) {
	var data []byte
	var myFile = "some/file.webm"
	for {
		select {
		case packet, ok := <-myStreamChan:
			if !ok {
				log.Print("WARNING: Failed to get packet")
			}

			//code to do something with packet
			if len(data) > 1000 {
				// the StreamToFile handler is included by default (see StreamToFile.go but you can write your own
				// or one to transform the stream or forward it to other clients. Anything, really!
				vidFile, fileOpenErr := os.OpenFile(myFile,
					os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if fileOpenErr != nil {
					log.Println(fileOpenErr)
				}
				defer vidFile.Close()
				fileStat, statErr := vidFile.Stat()

				if statErr != nil {
					log.Println(statErr)
				}

				_, writeErr := vidFile.Write(video)
				if writeErr == nil {
					data = nil
				}
			}
			data = append(data, packet...)
		case val, _ := <-context: //check the Camtron's global context channel for the signal to shut down
			if val == "stop" {
				close(myStreamChan)
				//do any other cleanup here
				return
			}
		}
	}
}
```
5. Call the function as a separate process to make it non-blocking
```
go MyStreamHandler(myStreamChan)
```

# Additional information
1. On Macos the Electron app should pop up a message asking for permission to use the camera. If it does not and the screen is black you may need to go into System > Security and allow it from there.
2. The Electron app uses localhost:8080 to send the stream to the Go library. Make sure this port is not blocked by the firewall.
3. If more than one webcam is attached a dropdown will be displayed allowing you to select the desired webcam

### Example project
[camtron-demo](https://github.com/vee2xx/camtron-demo)

### Source code for the Electron app
[camtron-ui](https://github.com/vee2xx/camtron-ui)
