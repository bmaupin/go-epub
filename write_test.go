package epub

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestEpubWriteTo(t *testing.T) {
	e := NewEpub(testEpubTitle)
	var b bytes.Buffer
	n, err := e.WriteTo(&b)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(b.Bytes())) != n {
		t.Fatalf("Expected size %v, got %v", len(b.Bytes()), n)
	}
}

func TestWriteToErrors(t *testing.T) {
	t.Run("CSS", func(t *testing.T) {
		e := NewEpub(testEpubTitle)
		testWriteToErrors(t, e, e.AddCSS, "cover.css")
	})
	t.Run("Font", func(t *testing.T) {
		e := NewEpub(testEpubTitle)
		testWriteToErrors(t, e, e.AddFont, "redacted-script-regular.ttf")
	})
	t.Run("Image", func(t *testing.T) {
		e := NewEpub(testEpubTitle)
		testWriteToErrors(t, e, e.AddImage, "gophercolor16x16.png")
	})
	t.Run("Video", func(t *testing.T) {
		e := NewEpub(testEpubTitle)
		testWriteToErrors(t, e, e.AddVideo, "sample_640x360.mp4")
	})
}

func testWriteToErrors(t *testing.T, e *Epub, adder func(string, string) (string, error), name string) {
	// Copy testdata to temp file
	data, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("cannot open testdata: %v", err)
	}
	defer data.Close()
	temp, err := ioutil.TempFile("", "temp")
	if err != nil {
		t.Fatalf("unable to create temp file: %v", err)
	}
	io.Copy(temp, data)
	temp.Close()
	// Add temp file to epub
	if _, err := adder(temp.Name(), ""); err != nil {
		t.Fatalf("unable to add temp file: %v", err)
	}
	// Delete temp file
	if err := os.Remove(temp.Name()); err != nil {
		t.Fatalf("unable to delete temp file: %v", err)
	}
	// Write epub to buffer
	var b bytes.Buffer
	if _, err := e.WriteTo(&b); err == nil {
		t.Fatal("Expected error")
	}
}
