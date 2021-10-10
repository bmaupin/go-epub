package epub

import (
	"bytes"
	"net/http"
	"sync"
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

func TestEpub_writeEpub(t *testing.T) {
	type fields struct {
		Mutex      sync.Mutex
		Client     *http.Client
		author     string
		cover      *epubCover
		css        map[string]string
		fonts      map[string]string
		identifier string
		images     map[string]string
		lang       string
		desc       string
		ppd        string
		pkg        *pkg
		sections   []epubSection
		title      string
		toc        *toc
	}
	type args struct {
		tempDir string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantDst string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Epub{
				Mutex:      tt.fields.Mutex,
				Client:     tt.fields.Client,
				author:     tt.fields.author,
				cover:      tt.fields.cover,
				css:        tt.fields.css,
				fonts:      tt.fields.fonts,
				identifier: tt.fields.identifier,
				images:     tt.fields.images,
				lang:       tt.fields.lang,
				desc:       tt.fields.desc,
				ppd:        tt.fields.ppd,
				pkg:        tt.fields.pkg,
				sections:   tt.fields.sections,
				title:      tt.fields.title,
				toc:        tt.fields.toc,
			}
			dst := &bytes.Buffer{}
			got, err := e.writeEpub(tt.args.tempDir, dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("Epub.writeEpub() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Epub.writeEpub() = %v, want %v", got, tt.want)
			}
			if gotDst := dst.String(); gotDst != tt.wantDst {
				t.Errorf("Epub.writeEpub() = %v, want %v", gotDst, tt.wantDst)
			}
		})
	}
}
