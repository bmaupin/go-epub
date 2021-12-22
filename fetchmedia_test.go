package epub

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var golangFavicon = strings.Replace(`AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAD///8AVE44//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb/
/uF2/1ROOP////8A////AFROOP/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+
4Xb//uF2//7hdv9UTjj/////AP///wBUTjj//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7h
dv/+4Xb//uF2//7hdv/+4Xb/VE44/////wD///8AVE44//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2
//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2/1ROOP////8A////AFROOP/+4Xb//uF2//7hdv/+4Xb/
/uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv9UTjj/////AP///wBUTjj//uF2//7hdv/+
4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb//uF2//7hdv/+4Xb/VE44/////wD///8AVE44//7h
dv/+4Xb//uF2//7hdv/+4Xb/z7t5/8Kyev/+4Xb//993///dd///3Xf//uF2/1ROOP////8A////
AFROOP/+4Xb//uF2//7hdv//4Hn/dIzD//v8///7/P//dIzD//7hdv//3Xf//913//7hdv9UTjj/
////AP///wBUTjj//uF2///fd//+4Xb//uF2/6ajif90jMP/dIzD/46Zpv/+4Xb//+F1///feP/+
4Xb/VE44/////wD///8AVE44//7hdv/z1XT////////////Is3L/HyAj/x8gI//Is3L/////////
///z1XT//uF2/1ROOP////8A19nd/1ROOP/+4Xb/5+HS//v+//8RExf/Liwn//7hdv/+4Xb/5+HS
//v8//8RExf/Liwn//7hdv9UTjj/19nd/1ROOP94aDT/yKdO/+fh0v//////ERMX/y4sJ//+4Xb/
/uF2/+fh0v//////ERMX/y4sJ//Ip07/dWU3/1ROOP9UTjj/yKdO/6qSSP/Is3L/9fb7//f6///I
s3L//uF2//7hdv/Is3L////////////Is3L/qpJI/8inTv9UTjj/19nd/1ROOP97c07/qpJI/8in
Tv/Ip07//uF2//7hdv/+4Xb//uF2/8zBlv/Kv4//pZJU/3tzTv9UTjj/19nd/////wD///8A4eLl
/6CcjP97c07/e3NO/1dOMf9BOiX/TkUn/2VXLf97c07/e3NO/6CcjP/h4uX/////AP///wD///8A
////AP///wD///8A////AP///wDq6/H/3N/j/9fZ3f/q6/H/////AP///wD///8A////AP///wD/
//8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
AAAAAAAAAAAAAA==`, "\n", "", -1)

func Test_fetchMedia(t *testing.T) {
	t.Run("LocalFS", func(t *testing.T) {
		Use(OsFS)
		testFetchMedia(t)
	})
	t.Run("MemoryFS", func(t *testing.T) {
		Use(MemoryFS)
		testFetchMedia(t)
	})
}

func testFetchMedia(t *testing.T) {
	filename := "gophercolor16x16.png"
	mux := http.NewServeMux()
	mux.HandleFunc("/image.png", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.Open(filepath.Join("testdata", filename))
		if err != nil {
			t.Fatal("cannot open testdata")
		}
		defer data.Close()
		io.Copy(w, data)
	}))
	mux.HandleFunc("/test.css", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "body{}")
	}))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	type args struct {
		mediaSource     string
		mediaFolderPath string
		mediaFilename   string
	}
	tests := []struct {
		name          string
		args          args
		wantMediaType string
		wantErr       bool
	}{
		{
			"URL request with test filename",
			args{
				mediaSource:     ts.URL + "/image.png",
				mediaFolderPath: "/",
				mediaFilename:   "test",
			},
			"image/png",
			false,
		},
		{
			"local file with test filename",
			args{
				mediaSource:     filepath.Join("testdata", filename),
				mediaFolderPath: "/",
				mediaFilename:   "test",
			},
			"image/png",
			false,
		},
		{
			"dataurl media with test filename",
			args{
				mediaSource:     `data:image/vnd.microsoft.icon;name=golang%20favicon;base64,` + golangFavicon,
				mediaFolderPath: "/",
				mediaFilename:   "test",
			},
			"image/x-icon",
			false,
		},
		{
			"bad request",
			args{
				mediaSource:     "badRequest",
				mediaFolderPath: "/",
				mediaFilename:   "test",
			},
			"",
			true,
		},
		{
			"empty filename",
			args{
				mediaSource:     "badRequest",
				mediaFolderPath: "/",
				mediaFilename:   "",
			},
			"",
			true,
		},
		{
			"bad path",
			args{
				mediaSource:     "badRequest",
				mediaFolderPath: "-",
				mediaFilename:   "",
			},
			"",
			true,
		},
		{
			"CSS",
			args{
				mediaSource:     ts.URL + "/test.css",
				mediaFolderPath: "/",
				mediaFilename:   "test.css",
			},
			"text/css",
			false,
		},
		{
			"bad request",
			args{
				mediaSource:     ts.URL + "/nonexistent",
				mediaFolderPath: "/",
				mediaFilename:   "test.css",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &grabber{http.DefaultClient}
			gotMediaType, err := g.fetchMedia(tt.args.mediaSource, tt.args.mediaFolderPath, tt.args.mediaFilename)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchMedia() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMediaType != tt.wantMediaType {
				t.Errorf("fetchMedia() = %v, want %v", gotMediaType, tt.wantMediaType)
			}
			var file fs.File
			if file, err = filesystem.Open(filepath.Join(tt.args.mediaFolderPath, tt.args.mediaFilename)); os.IsNotExist(err) {
				t.Errorf("fetchMedia(): file %v does not exist (source %v): %v", filepath.Join(tt.args.mediaFolderPath, tt.args.mediaFilename), tt.args.mediaSource, err)
			}
			if err == nil {
				file.Close()
			}
		})
	}
}
