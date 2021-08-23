package epub

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/vincent-petithory/dataurl"
)

// fetchMedia from mediaSource into mediaFolderPath as mediaFilename returning its type.
// the mediaSource can be a URL, a local path or an inline dataurl (as specified in RFC 2397)
func fetchMedia(mediaSource, mediaFolderPath, mediaFilename string) (mediaType string, err error) {

	mediaFilePath := filepath.Join(
		mediaFolderPath,
		mediaFilename,
	)
	// failfast, create the output file handler at the begining, if we cannot write the file, bail out
	w, err := os.Create(mediaFilePath)
	if err != nil {
		return "", fmt.Errorf("unable to create file: %s", err)
	}
	defer w.Close()
	var source io.ReadCloser
	fetchErrors := make([]error, 0)
	for _, f := range []func(string) (io.ReadCloser, error){
		fetchMediaFromURL,
		fetchMediaLocally,
		fetchMediaDataURL,
	} {
		var err error
		source, err = f(mediaSource)
		if err != nil {
			fetchErrors = append(fetchErrors, err)
			continue
		}
		break
	}
	if source == nil {
		return "", &FileRetrievalError{Source: mediaSource, Err: fetchError(fetchErrors)}

	}
	defer source.Close()
	_, err = io.Copy(w, source)
	if err != nil {
		// There shouldn't be any problem with the writer, but the reader
		// might have an issue
		return "", &FileRetrievalError{Source: mediaSource, Err: err}
	}
	// Detect the mediaType
	w.Seek(0, io.SeekStart)
	mime, err := mimetype.DetectReader(w)
	if err != nil {
		panic(err)
	}
	// Is it CSS?
	mtype := mime.String()
	if mime.Is("text/plain") {
		if filepath.Ext(mediaSource) == ".css" || filepath.Ext(mediaFilename) == ".css" {
			mtype = "text/css"
		}
	}
	return mtype, nil
}

func fetchMediaFromURL(mediaSource string) (io.ReadCloser, error) {
	resp, err := http.Get(mediaSource)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 400 {
		return nil, errors.New("cannot get file, bad return code")
	}
	return resp.Body, nil
}

func fetchMediaLocally(mediaSource string) (io.ReadCloser, error) {
	return os.Open(mediaSource)
}

func fetchMediaDataURL(mediaSource string) (io.ReadCloser, error) {
	data, err := dataurl.DecodeString(mediaSource)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(data.Data)), nil
}

type fetchError []error

func (f fetchError) Error() string {
	var message string
	for _, err := range f {
		message = fmt.Sprintf("%v\n %v", message, err.Error())
	}
	return message
}
