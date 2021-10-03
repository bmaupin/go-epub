package epub

import (
	"bytes"
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
		t.Fail()
	}
}
