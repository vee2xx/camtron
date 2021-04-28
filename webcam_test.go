package camtron

import (
	"os"
	"testing"

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

func TestCleanUp(t *testing.T) {
	_ = os.Remove("camtron-linux-x64.zip")
	_ = os.RemoveAll("camtron-linux-x64")
}
