package log

import (
	"testing"
	"time"
)

var tl *logger = NewLogger("./", "test.log")

func Test_rotate(t *testing.T) {
	for i := 0; i < 10; i++ {
		tl.rotate()
	}
}

func Test_checkFile_1(t *testing.T) {
	t.Logf("Test_checkFile is called.")
	tl.checkFile()
}
func Test_checkFile_2(t *testing.T) {
	t.Logf("Test_checkFile 2 is called.")
	tl.date = time.Now().Add((-24) * time.Hour)
	tl.checkFile()
}
