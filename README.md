# Golang Webcam
Camtron is a simple cross platform library written in go to easily have Go code interact with webcams i.e. consume and process a stream from a webcam consistantly across OS's.. It uses Electron and the MediaDevices Web API to access the webcam and allows a variety of consumers to listen for and process the video stream, for example saving the video to a file or sending it on to one or more endpoints. It is supported on Linux, Windows 10 and Macos. Currently the only supported codec is VP9. More will be added shortly.

## To use Camtron
1. Add github.com/vee2xx/camtron to go.mod in your project
2. Configure one or more consumers and pass them to camtron.StartCam as an array
```golang
var consumers map[string]camtron.StreamConsumer = make(map[string]camtron.StreamConsumer)
var vidToFileStream = make(chan []byte, 10)
options := make(map[string]string)
options["filePath"] = "./vids/vid-" + time.Now().Format("2006_01_02_15_04_05") + "." + config.Video.Format
options["maxSize"] = config.Video.MaxSize
streamConsumer := camtron.StreamConsumer{Stream: vidToFileStream, Context: make(chan string), Handler: camtron.StreamToFile, Options: options}
consumers["file"] = streamConsumer
camtron.StartCam(consumers)
```
3. Calling StartCam invokes the Electron app which retrieves a stream from the webcam and sends it to the Camtron over websockets. The first time the code is run it will download the Electron binary for the operating system you are on to the root directory of your project.
4. The StreamToFile handles is included in the library. The configuration options it uses are:
*  filePath: the directory in which to store the video
*  maxSize: the maximum size of the file to save. If the maximum size is exceeded the StreamToFile handler will stop processing the stream.
4. An example project is available here: [camtron-demo](https://github.com/vee2xx/camtron-demo)
5. The Electron code and binaries are available here: [camtron-ui](https://github.com/vee2xx/camtron-ui)

## Next enhancements
1. Support additional codecs
2. Add APIs to start and stop streaming
