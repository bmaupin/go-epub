package epub

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testAuthorTemplate    = `<dc:creator id="creator">%s</dc:creator>`
	testContainerContents = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="EPUB/package.opf" media-type="application/oebps-package+xml" />
  </rootfiles>
</container>`
	testDirPerm            = 0775
	testEpubAuthor         = "Hingle McCringleberry"
	testEpubFilename       = "My EPUB.epub"
	testEpubTitle          = "My title"
	testMimetypeContents   = "application/epub+zip"
	testPkgContentTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="pub-id" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id">urn:uuid:21ed94b4-f2ab-44c8-b99d-4f7792587ad6</dc:identifier>
    <dc:title>%s</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2016-04-28T19:09:26Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"></item>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"></item>
  </manifest>
  <spine toc="ncx"></spine>
</package>`
	testTempDirPrefix = "go-epub"
	testTitleTemplate = `<dc:title>%s</dc:title>`
)

var testTempDir = ""

func TestMain(m *testing.M) {
	// Run the tests
	retCode := m.Run()

	// Cleanup and exit
	//os.Remove(testEpubFilename)
	cleanup(testEpubFilename, testTempDir)
	os.Exit(retCode)
}

func TestEpubWrite(t *testing.T) {
	e := NewEpub(testEpubTitle)

	testTempDir = writeAndExtractEpub(t, e, testEpubFilename)
}

func TestEpubMimetypeContents(t *testing.T) {
	contents, err := ioutil.ReadFile(filepath.Join(testTempDir, mimetypeFilename))
	if err != nil {
		t.Errorf("Unexpected error reading mimetype file: %s", err)
	}
	if trimAllSpace(string(contents)) != trimAllSpace(testMimetypeContents) {
		t.Errorf(
			"Mimetype file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testMimetypeContents)
	}
}

func TestEpubContainerContents(t *testing.T) {
	contents, err := ioutil.ReadFile(filepath.Join(testTempDir, metaInfFolderName, containerFilename))
	if err != nil {
		t.Errorf("Unexpected error reading container file: %s", err)
	}
	if trimAllSpace(string(contents)) != trimAllSpace(testContainerContents) {
		t.Errorf(
			"Container file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testContainerContents)
	}
}

func TestEpubPkgContents(t *testing.T) {
	contents, err := ioutil.ReadFile(filepath.Join(testTempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testPkgContents := fmt.Sprintf(testPkgContentTemplate, testEpubTitle)
	if trimAllSpace(string(contents)) != trimAllSpace(testPkgContents) {
		t.Errorf(
			"Package file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testPkgContents)
	}
}

func TestEpubAuthor(t *testing.T) {
	authorTestEpubFilename := testEpubFilename + "author"

	e := NewEpub(testEpubTitle)
	e.SetAuthor(testEpubAuthor)

	if e.Author() != testEpubAuthor {
		t.Errorf(
			"Author doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Author(),
			testEpubAuthor)
	}

	tempDir := writeAndExtractEpub(t, e, authorTestEpubFilename)

	contents, err := ioutil.ReadFile(filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testAuthorElement := fmt.Sprintf(testAuthorTemplate, testEpubAuthor)
	if !strings.Contains(string(contents), testAuthorElement) {
		t.Errorf(
			"Author doesn't match\n"+
				"Expected: %s",
			testAuthorElement)
	}

	cleanup(authorTestEpubFilename, tempDir)
}

func TestEpubTitle(t *testing.T) {
	// First, test the title we provide when creating the epub
	titleTestEpubFilename := testEpubFilename + "title"

	e := NewEpub(testEpubTitle)
	if e.Title() != testEpubTitle {
		t.Errorf(
			"Title doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Title(),
			testEpubTitle)
	}

	tempDir := writeAndExtractEpub(t, e, titleTestEpubFilename)

	contents, err := ioutil.ReadFile(filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testTitleElement := fmt.Sprintf(testTitleTemplate, testEpubTitle)
	if !strings.Contains(string(contents), testTitleElement) {
		t.Errorf(
			"Title doesn't match\n"+
				"Expected: %s",
			testTitleElement)
	}

	cleanup(titleTestEpubFilename, tempDir)

	// Now test changing the title
	e.SetTitle(testEpubAuthor)

	if e.Title() != testEpubAuthor {
		t.Errorf(
			"Title doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Title(),
			testEpubAuthor)
	}

	tempDir = writeAndExtractEpub(t, e, titleTestEpubFilename)

	contents, err = ioutil.ReadFile(filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testTitleElement = fmt.Sprintf(testTitleTemplate, testEpubAuthor)
	if !strings.Contains(string(contents), testTitleElement) {
		t.Errorf(
			"Title doesn't match\n"+
				"Expected: %s",
			testTitleElement)
	}

	cleanup(titleTestEpubFilename, tempDir)
}

func cleanup(epubFilename string, tempDir string) {
	os.Remove(epubFilename)
	os.RemoveAll(tempDir)
}

// TrimAllSpace trims all space from each line of the string and removes empty
// lines for easier comparison
func trimAllSpace(s string) string {
	trimmedLines := []string{}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			trimmedLines = append(trimmedLines)
		}
	}

	return strings.Join(trimmedLines, "\n")
}

// UnzipFile unzips a file located at sourceFilePath to the provided destination directory
func unzipFile(sourceFilePath string, destDirPath string) error {
	// First, make sure the destination exists and is a directory
	info, err := os.Stat(destDirPath)
	if err != nil {
		return err
	}
	if !info.Mode().IsDir() {
		return errors.New("destination is not a directory")
	}

	r, err := zip.OpenReader(sourceFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	// Iterate through each file in the archive
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		destFilePath := filepath.Join(destDirPath, f.Name)

		// Create destination subdirectories if necessary
		destBaseDirPath, _ := filepath.Split(destFilePath)
		os.MkdirAll(destBaseDirPath, testDirPerm)

		// Create the destination file
		w, err := os.Create(destFilePath)
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil {
				panic(err)
			}
		}()

		// Copy the contents of the source file
		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeAndExtractEpub(t *testing.T, e *Epub, epubFilename string) string {
	tempDir, err := ioutil.TempDir("", tempDirPrefix)
	if err != nil {
		t.Errorf("Unexpected error creating temp dir: %s", err)
	}

	err = e.Write(epubFilename)
	if err != nil {
		t.Errorf("Unexpected error writing EPUB: %s", err)
	}

	err = unzipFile(epubFilename, tempDir)
	if err != nil {
		t.Errorf("Unexpected error extracting EPUB: %s", err)
	}

	return tempDir
}
