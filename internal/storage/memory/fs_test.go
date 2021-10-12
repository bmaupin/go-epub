package memory

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestMemory_Mkdir(t *testing.T) {
	fs := NewMemory()

	fs.Mkdir("test", 0666)
	f, _ := fs.Open("test")
	stat, _ := f.Stat()
	if !stat.IsDir() {
		t.Fail()
	}
	info, _ := f.(*file).Info()
	if info.Mode().IsRegular() {
		t.Fatal("unexpected regular file")
	}
}

func TestMemory_WriteFile(t *testing.T) {
	fs := NewMemory()

	fs.WriteFile("test", []byte{}, 0666)
	file, _ := fs.Open("test")
	stat, _ := file.Stat()
	if !stat.Mode().IsRegular() {
		t.Fail()
	}
}

func TestMemory_Create(t *testing.T) {
	fs := NewMemory()

	fs.Create("test")
	file, _ := fs.Open("test")
	stat, _ := file.Stat()
	if !stat.Mode().IsRegular() {
		t.Fail()
	}
}

func TestMemory(t *testing.T) {
	fs := NewMemory()
	err := fs.Mkdir("directory", 0666)
	if err != nil {
		t.Fatalf("mkdir error %v", err)
	}
	err = fs.WriteFile(filepath.Join("directory", "test"), []byte(`content`), 0666)
	if err != nil {
		t.Fatalf("writefile error: %v", err)
	}
	f, err := fs.Open(filepath.Join("directory", "test"))
	if err != nil {
		t.Fatalf("open error: %v", err)
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("readall error: %v", err)
	}
	if string(content) != "content" {
		t.Fatal("unexpected content")
	}
	err = fs.RemoveAll("directory")
	if err != nil {
		t.Fatalf("removeall err: %v", err)
	}
	_, err = fs.Open(filepath.Join("directory", "test"))
	if err == nil {
		t.Fatal("file should be gone")
	}
}
