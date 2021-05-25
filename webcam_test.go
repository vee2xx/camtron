package camtron

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBinaryDownload(t *testing.T) {
	downloadBinary("camtron-linux-x64")
	assert.FileExists(t, "camtron-linux-x64.zip")
}

func TestUnzip(t *testing.T) {
	UnzipBinary("camtron-linux-x64")
	assert.DirExists(t, "camtron-linux-x64")
}

//Use to test end to end. Requires manual intervention
func TestStreamToFile(t *testing.T) {
	StartStreamToFileConsumer()

	go StartCam()
	time.Sleep(60 * time.Second)
	ShutdownStream()
	assert.DirExists(t, "videos")
	_ = os.RemoveAll("videos")

}

func TestCleanUp(t *testing.T) {
	_ = os.Remove("camtron-linux-x64.zip")
	_ = os.RemoveAll("camtron-linux-x64")
	_ = os.Remove("camtron.log")
}
