package memory

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"testing"
)

func TestMemory_Mkdir(t *testing.T) {
	fs := NewMemory()

	err := fs.Mkdir("test", 0666)
	if err != nil {
		t.Fatal(err)
	}
	f, _ := fs.Open("test")
	stat, _ := f.Stat()
	if !stat.IsDir() {
		t.Fail()
	}
	info, _ := f.(*file).Info()
	if info.Mode().IsRegular() {
		t.Fatal("unexpected regular file")
	}
	// bad path
	err = fs.Mkdir("./..", 0666)
	if err == nil {
		t.Fatal(err)
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
	err := fs.WriteFile("./..", []byte{}, 0666)
	if err == nil {
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
	_, err := fs.Create("./..")
	if err == nil {
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
	assertPrefix(t, fs, "c")
	f, err := fs.Open(filepath.Join("directory", "test"))
	if err != nil {
		t.Fatalf("open error: %v", err)
	}
	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("readall error: %v", err)
	}
	if string(content) != "content" {
		t.Fatalf("unexpected content: unexpected 'content', got '%s'", string(content))
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

func TestMemory_ReadDir(t *testing.T) {
	dir := "test"

	fs := NewMemory()
	err := fs.Mkdir(dir, 0777)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := fs.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !stat.IsDir() {
		t.Fail()
	}
	_, err = fs.Create(path.Join(dir, "test.test"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Create(path.Join(dir, "test2.test"))
	if err != nil {
		t.Fatal(err)
	}
	dirs, err := fs.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 2 {
		t.Fail()
	}
}

func TestMemory_Stat(t *testing.T) {
	fs := NewMemory()
	_, err := fs.Stat("BADFILE")
	if err == nil {
		t.Fail()
	}
}

func assertPrefix(t *testing.T, fs *Memory, prefix string) {
	t.Helper()

	// Open file
	f, err := fs.Open(filepath.Join("directory", "test"))
	if err != nil {
		t.Fatalf("open error: %v", err)
	}
	defer f.Close()

	// Read length bytes
	length := len(prefix)
	b := make([]byte, length)
	n, err := f.Read(b)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if n != length {
		t.Fatal("incorrect read count")
	}
	if string(b) != prefix {
		t.Fatalf("unexpected content: unexpected '%s', got '%s'", prefix, string(b))
	}
}
