package memory

import (
	"fmt"
	"io/fs"
	"testing"
	"time"
)

func Test_file(t *testing.T) {
	name := "test"
	now := time.Now()
	content := "test"
	f := &file{
		name:    name,
		modTime: now,
	}
	fmt.Fprint(f, content)
	if f.Size() != int64(len(content)) {
		t.Fail()
	}
	if f.ModTime() != now {
		t.Fail()
	}
	if f.Name() != name {
		t.Fail()
	}
	if f.Type()&fs.ModeType != 0 {
		t.Fail()
	}
	_ = f.Sys()
	f.Close()
}
