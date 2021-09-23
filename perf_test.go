package epub

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkAddImage_http(b *testing.B) {
	filename := "gophercolor16x16.png"
	mux := http.NewServeMux()
	mux.HandleFunc("/image.png", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			data, err := os.Open(filepath.Join("testdata", filename))
			if err != nil {
				b.Fatal("cannot open testdata")
			}
			defer data.Close()
			io.Copy(w, data)
		case "HEAD":
			w.WriteHeader(http.StatusOK)
		}
	}))
	ts := httptest.NewServer(mux)
	defer ts.Close()
	e := NewEpub("test")
	for i := 0; i < b.N; i++ {
		_, err := e.AddImage(ts.URL+"/image.png", "")
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkAddImage_file(b *testing.B) {
	e := NewEpub("test")
	for i := 0; i < b.N; i++ {
		_, err := e.AddImage("testdata/gophercolor16x16.png", "")
		if err != nil {
			b.Fatal(err)
		}
	}
}
