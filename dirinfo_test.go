package epub

import (
	"bytes"
	"io/fs"
	"testing"
	"time"
)

func Test_fileInfoToDirEntry(t *testing.T) {
	f := file{
		name: "test",
		mode: fs.ModeDir,
	}
	d := fileInfoToDirEntry(&f)
	_, err := d.Info()
	if err != nil {
		t.Fatal(err)
	}
	if !d.IsDir() {
		t.Fail()
	}
	if d.Type()&fs.ModeDir == 0 {
		t.Fail()
	}

	if d.Name() != "test" {
		t.Fail()
	}
	if fileInfoToDirEntry(nil) != nil {
		t.Fail()
	}
}

type file struct {
	name    string
	modTime time.Time
	content bytes.Buffer
	mode    fs.FileMode
}

func (f *file) Info() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) Stat() (fs.FileInfo, error) {
	return f, nil
}

func (f *file) Read(b []byte) (int, error) {
	return f.content.Read(b)
}

func (f *file) Close() error {
	return nil
}

func (f *file) Write(p []byte) (n int, err error) {
	return f.content.Write(p)
}

func (f *file) Name() string {
	return f.name
}

func (f *file) Size() int64 {
	return int64(f.content.Len())
}

func (f *file) Type() fs.FileMode {
	return f.mode & fs.ModeType
}

func (f *file) Mode() fs.FileMode {
	return f.mode
}

func (f *file) ModTime() time.Time {
	return f.modTime
}

func (f *file) IsDir() bool {
	return f.mode&fs.ModeDir != 0
}

func (f *file) Sys() interface{} {
	return nil
}
