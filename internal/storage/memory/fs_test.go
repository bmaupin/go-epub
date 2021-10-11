package memory

import (
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
