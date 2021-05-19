# Go Webcam Simplified
Camtron is a simple cross platform library written in go to easily have Go code interact with webcams i.e. consume and process a stream from a webcam consistantly across OS's without relying on opencv. It uses Electron and the MediaDevices Web API to access the webcam and allows a variety of consumers to listen for and process the video stream, for example recording a video from a webcam and saving the video to a file or sending the video from a webcam on to one or more endpoints. It is supported on Linux, Windows 10 and Macos. Currently the only supported codec is VP9. More will be added shortly.

## Install
There are two ways to install Camtron
1. Download it using 'go get'
```
go get github.com/vee2xx/camtron
```
2. Or add github.com/vee2xx/camtron to go.mod file of your project
```
require (
	github.com/vee2xx/camtron v1.0.8
)
```
3. The first time Camtron runs it will download and unzip the os appropriate camtron-ui package to your project's root directory so that Camtron can find the Electron app binary and execute it.

### Record a video and save it to a file
1. Create a project add camtron to the imports at the top of main.go
```
 "github.com/vee2xx/camtron"
```
2. Camtron comes with a built in stream handler that will save the incoming video to a file. To stream a video to a file add the following two lines to your main function
```
StartStreamToFileConsumer() //start a listener that accepts and processes the stream
go StartCam() //start the Electron app that connects to the webcam and captures the stream
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
	...
	
	for {
		select {
		case packet, ok := <-myStreamChan:
			if !ok {
				log.Print("WARNING: Failed to get packet")
			}

			//code to do something with packet
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
