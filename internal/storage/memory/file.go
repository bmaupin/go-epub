package memory

import (
	"bytes"
	"io/fs"
	"time"
)

type file struct {
	name    string
	isDir   bool
	modTime time.Time
	content bytes.Buffer
	perm    fs.FileMode
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

func (f *file) Mode() fs.FileMode {
	return f.perm
}

func (f *file) ModTime() time.Time {
	return f.modTime
}

func (f *file) IsDir() bool {
	return f.isDir
}

func (f *file) Sys() interface{} {
	return nil
}
