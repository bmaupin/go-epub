package epub

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-shiori/go-epub/internal/storage"
	"github.com/gofrs/uuid"
)

const (
	// Set this to false to not delete the generated test EPUB file
	doCleanup             = true
	testAuthorTemplate    = `<dc:creator id="creator">%s</dc:creator>`
	testContainerContents = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="EPUB/package.opf" media-type="application/oebps-package+xml" />
  </rootfiles>
</container>`
	testCoverCSSFilename     = "cover.css"
	testCoverCSSSource       = "testdata/cover.css"
	testCoverContentTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
  <head>
    <title dir="auto">%s</title>
    <link rel="stylesheet" type="text/css" href="%s"></link>
  </head>
  <body dir="auto">
    <img src="%s" alt="Cover Image" />
  </body>
</html>`
	testCSSLinkTemplate       = `<link rel="stylesheet" type="text/css" href="%s"></link>`
	testDirPerm               = 0775
	testEpubAuthor            = "Hingle McCringleberry"
	testEpubcheckJarfile      = "epubcheck.jar"
	testEpubcheckPrefix       = "epubcheck"
	testEpubFilename          = "My EPUB.epub"
	testEpubIdentifier        = "urn:uuid:51b7c9ea-b2a2-49c6-9d8c-522790786d15"
	testEpubLang              = "fr"
	testEpubPpd               = "rtl"
	testEpubTitle             = "My title"
	testEpubDescription       = "My description"
	testFontCSSFilename       = "font.css"
	testFontCSSSource         = "testdata/font.css"
	testFontFromFileSource    = "testdata/redacted-script-regular.ttf"
	testIdentifierTemplate    = `<dc:identifier id="pub-id">%s</dc:identifier>`
	testImageFromFileFilename = "testfromfile.png"
	testImageFromFileSource   = "testdata/gophercolor16x16.png"
	testNumberFilenameStart   = "01filenametest.png"
	testSpaceInFilename       = "filename with space.png"
	testVideoFromFileFilename = "testfromfile.mp4"
	testVideoFromFileSource   = "testdata/sample_640x360.mp4"
	testAudioFromFileFilename = "sample_audio.wav"
	testAudioFromFileSource   = "testdata/sample_audio.wav"
	testLangTemplate          = `<dc:language>%s</dc:language>`
	testDescTemplate          = `<dc:description>%s</dc:description>`
	testPpdTemplate           = `page-progression-direction="%s"`
	testMimetypeContents      = "application/epub+zip"
	testPkgContentTemplate    = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="pub-id" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id">%s</dc:identifier>
    <dc:title>%s</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">%s</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"></item>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"></item>
  </manifest>
  <spine toc="ncx"></spine>
</package>`
	testSectionBody = `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	testSectionBodyWithnotabledownloadImage = `    <h1>Section 1</h1>
	<p>This is a paragraph.</p><p><img src="https://example.com/fileNotExist.jpg" loading="lazy"/></p>`
	testSectionBodyWithImageEmbed = `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>
	<p><img src="../images/gophercolor16x16.png" loading="lazy"/></p>`
	testSectionContentTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
  <head>
    <title dir="auto">%s</title>
  </head>
  <body dir="auto">
    %s
  </body>
</html>`
	testSectionFilename = "section0001.xhtml"
	testSectionTitle    = "Section 1"
	testTempDirPrefix   = "go-epub"
	testTitleTemplate   = `<dc:title>%s</dc:title>`
)

func TestEpubWrite(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	// Check the contents of the package file
	// NOTE: This is tested first because it contains a timestamp; testing it later may result in a different timestamp
	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testPkgContents := fmt.Sprintf(testPkgContentTemplate, e.Identifier(), testEpubTitle, time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	if trimAllSpace(string(contents)) != trimAllSpace(testPkgContents) {
		t.Errorf(
			"Package file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testPkgContents)
	}

	// Check the contents of the mimetype file
	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, mimetypeFilename))
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

	// Check the contents of the container file
	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, metaInfFolderName, containerFilename))
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

	cleanup(testEpubFilename, tempDir)
}

func TestAddCSS(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testCSS1Path, err := e.AddCSS(testCoverCSSSource, testCoverCSSFilename)
	if err != nil {
		t.Errorf("Error adding CSS: %s", err)
	}

	testCSS2Path, err := e.AddCSS(testCoverCSSSource, "")
	if err != nil {
		t.Errorf("Error adding CSS: %s", err)
	}

	// Add a section with CSS to make sure the stylesheet link for a section is properly created
	testSectionPath, err := e.AddSection(testSectionBody, testSectionTitle, testSectionFilename, testCSS1Path)
	if err != nil {
		t.Errorf("Error adding section with CSS: %s", err)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	// The CSS file path is relative to the XHTML folder
	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testCSS1Path))
	if err != nil {
		t.Errorf("Unexpected error reading CSS file: %s", err)
	}

	testCSSContents, err := os.ReadFile(testCoverCSSSource)
	if err != nil {
		t.Errorf("Unexpected error reading CSS file: %s", err)
	}

	if trimAllSpace(string(contents)) != trimAllSpace(string(testCSSContents)) {
		t.Errorf(
			"CSS file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testCSSContents)
	}

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testCSS2Path))
	if err != nil {
		t.Errorf("Unexpected error reading CSS file: %s", err)
	}

	if trimAllSpace(string(contents)) != trimAllSpace(string(testCSSContents)) {
		t.Errorf(
			"CSS file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testCSSContents)
	}

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSectionPath))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}

	testCSSLinkElement := fmt.Sprintf(testCSSLinkTemplate, testCSS1Path)
	if !strings.Contains(string(contents), testCSSLinkElement) {
		t.Errorf(
			"CSS link doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testCSSLinkElement)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestAddFont(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testFontFromFilePath, err := e.AddFont(testFontFromFileSource, "")
	if err != nil {
		t.Errorf("Error adding font: %s", err)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	// The font path is relative to the XHTML folder
	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testFontFromFilePath))
	if err != nil {
		t.Errorf("Unexpected error reading font file from EPUB: %s", err)
	}

	testFontContents, err := os.ReadFile(testFontFromFileSource)
	if err != nil {
		t.Errorf("Unexpected error reading testdata font file: %s", err)
	}
	if bytes.Compare(contents, testFontContents) != 0 {
		t.Errorf("Font file contents don't match")
	}

	cleanup(testEpubFilename, tempDir)
}

func TestAddImage(t *testing.T) {
	fs := http.FileServer(http.Dir("./testdata/"))

	// start a test server with the file server handler
	server := httptest.NewServer(fs)
	defer server.Close()

	testImageFromURLSource := server.URL + "/gophercolor16x16.png"
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testImageFromFilePath, err := e.AddImage(testImageFromFileSource, testImageFromFileFilename)
	if err != nil {
		t.Errorf("Error adding image: %s", err)
	}

	testImageFromURLPath, err := e.AddImage(testImageFromURLSource, "")
	if err != nil {
		t.Errorf("Error adding image: %s", err)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	// The image path is relative to the XHTML folder
	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testImageFromFilePath))
	if err != nil {
		t.Errorf("Unexpected error reading image file from EPUB: %s", err)
	}

	testImageContents, err := os.ReadFile(testImageFromFileSource)
	if err != nil {
		t.Errorf("Unexpected error reading testdata image file: %s", err)
	}
	if bytes.Compare(contents, testImageContents) != 0 {
		t.Errorf("Image file contents don't match")
	}

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testImageFromURLPath))
	if err != nil {
		t.Errorf("Unexpected error reading image file from EPUB: %s", err)
	}

	resp, err := http.Get(testImageFromURLSource)
	if err != nil {
		t.Errorf("Unexpected error response from test image URL: %s", err)
	}
	testImageContents, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error reading test image file from URL: %s", err)
	}
	if bytes.Compare(contents, testImageContents) != 0 {
		t.Errorf("Image file contents don't match")
	}

	cleanup(testEpubFilename, tempDir)
}

func TestAddVideo(t *testing.T) {
	fs := http.FileServer(http.Dir("./testdata/"))

	// start a test server with the file server handler
	server := httptest.NewServer(fs)
	defer server.Close()

	testVideoFromURLSource := server.URL + "/sample_640x360.mp4"

	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testVideoFromFilePath, err := e.AddVideo(testVideoFromFileSource, testVideoFromFileFilename)
	if err != nil {
		t.Errorf("Error adding video: %s", err)
	}
	fmt.Println(testVideoFromFilePath)

	testVideoFromURLPath, err := e.AddVideo(testVideoFromURLSource, "")
	if err != nil {
		t.Errorf("Error adding video: %s", err)
	}
	fmt.Println(testVideoFromURLPath)

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	// The video path is relative to the XHTML folder
	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testVideoFromFilePath))
	if err != nil {
		t.Errorf("Unexpected error reading video file from EPUB: %s", err)
	}

	testVideoContents, err := os.ReadFile(testVideoFromFileSource)
	if err != nil {
		t.Errorf("Unexpected error reading testdata video file: %s", err)
	}
	if bytes.Compare(contents, testVideoContents) != 0 {
		t.Errorf("Video file contents don't match")
	}

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testVideoFromURLPath))
	if err != nil {
		t.Errorf("Unexpected error reading video file from EPUB: %s", err)
	}

	resp, err := http.Get(testVideoFromURLSource)
	if err != nil {
		t.Errorf("Unexpected error response from test video URL: %s", err)
	}
	testVideoContents, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error reading test video file from URL: %s", err)
	}
	if bytes.Compare(contents, testVideoContents) != 0 {
		t.Errorf("Video file contents don't match")
	}

	cleanup(testEpubFilename, tempDir)
}

func TestAddAudio(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testAudioFromFilePath, err := e.AddAudio(testAudioFromFileSource, testAudioFromFileFilename)
	if err != nil {
		t.Errorf("Error adding audio: %s", err)
	}
	fmt.Println(testAudioFromFilePath)

	fs := http.FileServer(http.Dir("./testdata/"))

	// start a test server with the file server handler
	server := httptest.NewServer(fs)
	defer server.Close()

	testAudioFromURLSource := server.URL + "/sample_audio.wav"
	testAudioFromURLPath, err := e.AddAudio(testAudioFromURLSource, "")
	if err != nil {
		t.Errorf("Error adding audio: %s", err)
	}
	fmt.Println(testAudioFromURLPath)

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	// The audio path is relative to the XHTML folder
	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testAudioFromFilePath))
	if err != nil {
		t.Errorf("Unexpected error reading audio file from EPUB: %s", err)
	}

	testAudioContents, err := os.ReadFile(testAudioFromFileSource)
	if err != nil {
		t.Errorf("Unexpected error reading testdata audio file: %s", err)
	}
	if bytes.Compare(contents, testAudioContents) != 0 {
		t.Errorf("Audio file contents don't match")
	}

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testAudioFromURLPath))
	if err != nil {
		t.Errorf("Unexpected error reading audio file from EPUB: %s", err)
	}

	resp, err := http.Get(testAudioFromURLSource)
	if err != nil {
		t.Errorf("Unexpected error response from test audio URL: %s", err)
	}
	testAudioContents, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error reading test audio file from URL: %s", err)
	}
	if bytes.Compare(contents, testAudioContents) != 0 {
		t.Errorf("Audio file contents don't match")
	}

	cleanup(testEpubFilename, tempDir)
}

func TestAddSection(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testSection1Path, err := e.AddSection(testSectionBody, testSectionTitle, testSectionFilename, "")
	if err != nil {
		t.Errorf("Error adding section: %s", err)
	}

	testSection2Path, err := e.AddSection(testSectionBody, testSectionTitle, "", "")
	if err != nil {
		t.Errorf("Error adding section: %s", err)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSection1Path))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}

	testSectionContents := fmt.Sprintf(testSectionContentTemplate, testSectionTitle, testSectionBody)
	if trimAllSpace(string(contents)) != trimAllSpace(testSectionContents) {
		t.Errorf(
			"Section file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testSectionContents)
	}

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSection2Path))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}

	if trimAllSpace(string(contents)) != trimAllSpace(testSectionContents) {
		t.Errorf(
			"Section file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testSectionContents)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestAddSubSection(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testSection1Path, err := e.AddSection(testSectionBody, testSectionTitle, testSectionFilename, "")
	if err != nil {
		t.Errorf("Error adding section: %s", err)
	}

	testSection2Path, err := e.AddSubSection(testSection1Path, testSectionBody, testSectionTitle, "", "")
	if err != nil {
		t.Errorf("Error adding subsection: %s", err)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSection1Path))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}

	testSectionContents := fmt.Sprintf(testSectionContentTemplate, testSectionTitle, testSectionBody)
	if trimAllSpace(string(contents)) != trimAllSpace(testSectionContents) {
		t.Errorf(
			"Section file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testSectionContents)
	}

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSection2Path))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}

	if trimAllSpace(string(contents)) != trimAllSpace(testSectionContents) {
		t.Errorf(
			"Section file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testSectionContents)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestEpubAuthor(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	e.SetAuthor(testEpubAuthor)

	if e.Author() != testEpubAuthor {
		t.Errorf(
			"Author doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Author(),
			testEpubAuthor)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testAuthorElement := fmt.Sprintf(testAuthorTemplate, testEpubAuthor)
	if !strings.Contains(string(contents), testAuthorElement) {
		t.Errorf(
			"Author doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testAuthorElement)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestEpubLang(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	e.SetLang(testEpubLang)

	if e.Lang() != testEpubLang {
		t.Errorf(
			"Language doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Lang(),
			testEpubLang)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testLangElement := fmt.Sprintf(testLangTemplate, testEpubLang)
	if !strings.Contains(string(contents), testLangElement) {
		t.Errorf(
			"Language doesn't match\n"+
				"Got: %s"+
				"Expected: %s",
			contents,
			testLangElement)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestEpubPpd(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	e.SetPpd(testEpubPpd)

	if e.Ppd() != testEpubPpd {
		t.Errorf(
			"Ppd doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Ppd(),
			testEpubPpd)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testPpdElement := fmt.Sprintf(testPpdTemplate, testEpubPpd)
	if !strings.Contains(string(contents), testPpdElement) {
		t.Errorf(
			"Ppd doesn't match\n"+
				"Got: %s"+
				"Expected: %s",
			contents,
			testPpdElement)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestEpubTitle(t *testing.T) {
	// First, test the title we provide when creating the epub
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	if e.Title() != testEpubTitle {
		t.Errorf(
			"Title doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Title(),
			testEpubTitle)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testTitleElement := fmt.Sprintf(testTitleTemplate, testEpubTitle)
	if !strings.Contains(string(contents), testTitleElement) {
		t.Errorf(
			"Title doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testTitleElement)
	}

	cleanup(testEpubFilename, tempDir)

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

	tempDir = writeAndExtractEpub(t, e, testEpubFilename)

	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testTitleElement = fmt.Sprintf(testTitleTemplate, testEpubAuthor)
	if !strings.Contains(string(contents), testTitleElement) {
		t.Errorf(
			"Title doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testTitleElement)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestEpubDescription(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	e.SetDescription(testEpubDescription)

	if e.Description() != testEpubDescription {
		t.Errorf(
			"Description doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Lang(),
			testEpubDescription)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testLangElement := fmt.Sprintf(testDescTemplate, testEpubDescription)
	if !strings.Contains(string(contents), testLangElement) {
		t.Errorf(
			"Description doesn't match\n"+
				"Got: %s"+
				"Expected: %s",
			contents,
			testLangElement)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestEpubIdentifier(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	e.SetIdentifier(testEpubIdentifier)

	if e.Identifier() != testEpubIdentifier {
		t.Errorf(
			"Identifier doesn't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			e.Identifier(),
			testEpubIdentifier)
	}

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	testIdentifierElement := fmt.Sprintf(testIdentifierTemplate, testEpubIdentifier)
	if !strings.Contains(string(contents), testIdentifierElement) {
		t.Errorf(
			"Identifier doesn't match\n"+
				"Got: %s"+
				"Expected: %s",
			contents,
			testIdentifierElement)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestSetCover(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testImagePath, _ := e.AddImage(testImageFromFileSource, testImageFromFileFilename)
	testCSSPath, _ := e.AddCSS(testCoverCSSSource, testCoverCSSFilename)
	e.SetCover(testImagePath, testCSSPath)

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, defaultCoverXhtmlFilename))
	if err != nil {
		t.Errorf("Unexpected error reading cover XHTML file: %s", err)
	}

	testCoverContents := fmt.Sprintf(testCoverContentTemplate, testEpubTitle, testCSSPath, testImagePath)
	if trimAllSpace(string(contents)) != trimAllSpace(testCoverContents) {
		t.Errorf(
			"Cover file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testCoverContents)
	}

	cleanup(testEpubFilename, tempDir)
}

func TestManifestItems(t *testing.T) {
	fs := http.FileServer(http.Dir("./testdata/"))

	// start a test server with the file server handler
	server := httptest.NewServer(fs)
	defer server.Close()

	testImageFromURLSource := server.URL + "/gophercolor16x16.png"
	testManifestItems := []string{`id="filenamewithspace.png" href="images/filename with space.png" media-type="image/png"></item>`,
		`id="gophercolor16x16.png" href="images/gophercolor16x16.png" media-type="image/png"></item>`,
		`id="id01filenametest.png" href="images/01filenametest.png" media-type="image/png"></item>`,
		`id="image0005.png" href="images/image0005.png" media-type="image/png"></item>`,
		`id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"></item>`,
		`id="testfromfile.png" href="images/testfromfile.png" media-type="image/png"></item>`,
	}

	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	e.AddImage(testImageFromFileSource, testImageFromFileFilename)
	e.AddImage(testImageFromFileSource, "")
	// In particular, we want to test these next two, which will be modified by fixXMLId()
	e.AddImage(testImageFromFileSource, testNumberFilenameStart)
	e.AddImage(testImageFromFileSource, testSpaceInFilename)
	e.AddImage(testImageFromURLSource, "")

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	pkgFileContent, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, pkgFilename))
	if err != nil {
		t.Errorf("Unexpected error reading package file: %s", err)
	}

	// Get just the manifest portion of the package file
	manifestContentFromFile := string(pkgFileContent)[strings.Index(string(pkgFileContent), "<manifest>"):strings.Index(string(pkgFileContent), "</manifest>")]
	// Convert the manifest portion of the package file to a slice
	pkgFileManifestItems := strings.Split(manifestContentFromFile, "<item")
	// Drop the <manifest> and </manifest>
	pkgFileManifestItems = pkgFileManifestItems[1 : len(pkgFileManifestItems)-1]
	// Trim whitespace for each item
	for i := range pkgFileManifestItems {
		pkgFileManifestItems[i] = strings.TrimSpace(pkgFileManifestItems[i])
	}
	// Sort the manifest items from the package file (they will be in a random order)
	sort.Strings(pkgFileManifestItems)

	// Compare the slices by converting them to strings
	if strings.Join(pkgFileManifestItems[:], ",") != strings.Join(testManifestItems[:], ",") {
		t.Errorf(
			"Package file manifest items don't match\n"+
				"Got: \n%s\n\n"+
				"Expected: \n%s\n",
			strings.Join(pkgFileManifestItems[:], "\n"),
			strings.Join(testManifestItems[:], "\n"))
	}

	cleanup(testEpubFilename, tempDir)
}

func TestFilenameAlreadyUsedError(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	_, err = e.AddCSS(testCoverCSSSource, testCoverCSSFilename)
	if err != nil {
		t.Errorf("Error adding CSS: %s", err)
	}

	_, err = e.AddCSS(testCoverCSSSource, testCoverCSSFilename)
	if _, ok := err.(*FilenameAlreadyUsedError); !ok {
		t.Errorf("Expected error FilenameAlreadyUsedError not returned. Returned instead: %+v", err)
	}
}

func TestFileRetrievalError(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	_, err = e.AddCSS("/sbin/thisShouldFail", testCoverCSSFilename)
	if _, ok := err.(*FileRetrievalError); !ok {
		t.Errorf("Expected error FileRetrievalError not returned. Returned instead: %+v", err)
	}
}

func TestUnableToCreateEpubError(t *testing.T) {
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	err = e.Write("/sbin/thisShouldFail")
	if _, ok := err.(*UnableToCreateEpubError); !ok {
		t.Errorf("Expected error UnableToCreateEpubError not returned. Returned instead: %+v", err)
	}
}

func TestEmbedImage(t *testing.T) {
	fs := http.FileServer(http.Dir("./testdata/"))

	// start a test server with the file server handler
	server := httptest.NewServer(fs)
	defer server.Close()

	testSectionBodyWithImage := `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>
	<p><img src="` + server.URL + `/gophercolor16x16.png" loading="lazy"/></p>`
	testSectionBodyWithImageExpect := `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>
	<p><img src="../images/gophercolor16x16.png" loading="lazy"/></p>`
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testSection1Path, err := e.AddSection(testSectionBody, testSectionTitle, testSectionFilename, "")
	if err != nil {
		t.Errorf("Error adding section: %s", err)
	}

	testSection2Path, err := e.AddSection(testSectionBodyWithnotabledownloadImage, testSectionTitle, "", "")
	if err != nil {
		t.Errorf("Error adding section: %s", err)
	}
	testSection3Path, err := e.AddSection(testSectionBodyWithImage, testSectionTitle, "", "")
	if err != nil {
		t.Errorf("Error adding section: %s", err)
	}
	e.EmbedImages()
	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	contents, err := storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSection1Path))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}
	// test 1
	testSectionContents := fmt.Sprintf(testSectionContentTemplate, testSectionTitle, testSectionBody)
	if trimAllSpace(string(contents)) != trimAllSpace(testSectionContents) {
		t.Errorf(
			"Section file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testSectionContents)
	}
	// test 2
	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSection2Path))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}

	testSection2Contents := fmt.Sprintf(testSectionContentTemplate, testSectionTitle, testSectionBodyWithnotabledownloadImage)
	if trimAllSpace(string(contents)) != trimAllSpace(testSection2Contents) {
		t.Errorf(
			"Section file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testSection2Contents)
	}
	// test 3
	contents, err = storage.ReadFile(filesystem, filepath.Join(tempDir, contentFolderName, xhtmlFolderName, testSection3Path))
	if err != nil {
		t.Errorf("Unexpected error reading section file: %s", err)
	}

	testSection3Contents := fmt.Sprintf(testSectionContentTemplate, testSectionTitle, testSectionBodyWithImageExpect)
	if trimAllSpace(string(contents)) != trimAllSpace(testSection3Contents) {
		t.Errorf(
			"Section file contents don't match\n"+
				"Got: %s\n"+
				"Expected: %s",
			contents,
			testSection3Contents)
	}
	cleanup(testEpubFilename, tempDir)
}

func testEpubValidity(t testing.TB) {
	fs := http.FileServer(http.Dir("./testdata/"))

	// start a test server with the file server handler
	server := httptest.NewServer(fs)
	defer server.Close()

	testAudioFromURLSource := server.URL + "/sample_audio.wav"
	testImageFromURLSource := server.URL + "/gophercolor16x16.png"
	testVideoFromURLSource := server.URL + "/sample_640x360.mp4"
	e, err := NewEpub(testEpubTitle)
	if err != nil {
		t.Error(err)
	}

	testCoverCSSPath, _ := e.AddCSS(testCoverCSSSource, testCoverCSSFilename)
	e.AddCSS(testCoverCSSSource, "")
	testSectionPath, _ := e.AddSection(testSectionBody, testSectionTitle, testSectionFilename, testCoverCSSPath)
	e.AddSubSection(testSectionPath, testSectionBody, "Test subsection", "subsection.xhtml", "")

	e.AddFont(testFontFromFileSource, "")
	// Add CSS referencing the font in order to validate the font MIME type
	testFontCSSPath, _ := e.AddCSS(testFontCSSSource, testFontCSSFilename)

	e.AddSection(testSectionBody, testSectionTitle, "", testFontCSSPath)

	testImagePath, _ := e.AddImage(testImageFromFileSource, testImageFromFileFilename)
	e.AddImage(testImageFromFileSource, testImageFromFileFilename)
	e.AddImage(testImageFromURLSource, "")
	e.AddImage(testImageFromFileSource, testNumberFilenameStart)
	e.AddVideo(testVideoFromURLSource, testVideoFromFileFilename)
	e.AddAudio(testAudioFromURLSource, testAudioFromFileFilename)
	e.SetAuthor(testEpubAuthor)
	e.SetCover(testImagePath, "")
	e.SetDescription(testEpubDescription)
	e.SetIdentifier(testEpubIdentifier)
	e.SetLang(testEpubLang)
	e.SetPpd(testEpubPpd)
	e.SetTitle(testEpubAuthor)

	tempDir := writeAndExtractEpub(t, e, testEpubFilename)

	output, err := validateEpub(t, testEpubFilename)
	if err != nil {
		t.Errorf("EPUB validation failed")
	}

	// Always print the output so we can see warnings as well
	if output != nil {
		fmt.Println(string(output))
	}
	if doCleanup {
		cleanup(testEpubFilename, tempDir)
	} else {
		// Always remove the files in tempDir; they can still be extracted from the test epub as needed
		filesystem.RemoveAll(tempDir)
	}

}

func BenchmarkEpubValidity(b *testing.B) {
	b.Run("LocalFS", func(b *testing.B) {
		Use(OsFS)
		for i := 0; i < b.N; i++ {
			testEpubValidity(b)
		}
	})
	b.Run("MemoryFS", func(b *testing.B) {
		Use(MemoryFS)
		for i := 0; i < b.N; i++ {
			testEpubValidity(b)
		}
	})

}

func TestEpubValidity(t *testing.T) {
	t.Run("LocalFS", func(t *testing.T) {
		Use(OsFS)
		testEpubValidity(t)
	})
	t.Run("MemoryFS", func(t *testing.T) {
		Use(MemoryFS)
		testEpubValidity(t)
	})
}

func cleanup(epubFilename string, tempDir string) {
	os.Remove(epubFilename)
	filesystem.RemoveAll(tempDir)
}

// TrimAllSpace trims all space from each line of the string and removes empty
// lines for easier comparison
func trimAllSpace(s string) string {
	trimmedLines := []string{}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			trimmedLines = append(trimmedLines, line)
		}
	}

	return strings.Join(trimmedLines, "\n")
}

// UnzipFile unzips a file located at sourceFilePath to the provided destination directory
func unzipFile(sourceFilePath string, destDirPath string) error {
	var t *testing.T
	// First, make sure the destination exists and is a directory
	f, err := filesystem.Open(destDirPath)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("destination is not a directory")
	}

	r, err := zip.OpenReader(sourceFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Error(err)
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
				t.Error(err)
			}
		}()

		destFilePath := filepath.Join(destDirPath, strings.TrimLeft(f.Name, filepath.Dir(sourceFilePath)))

		// Create destination subdirectories if necessary
		destBaseDirPath, _ := filepath.Split(destFilePath)
		err = storage.MkdirAll(filesystem, destBaseDirPath, testDirPerm)
		if err != nil {
			return err
		}

		// Create the destination file
		w, err := filesystem.Create(destFilePath)
		if err != nil {
			return err
		}
		defer func() {
			if err := w.Close(); err != nil {
				t.Error(err)
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

// This function requires EPUBCheck to work; see README.md for more information
func validateEpub(t testing.TB, epubFilename string) ([]byte, error) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Error("Error getting working directory")
	}

	items, err := ioutil.ReadDir(cwd)
	if err != nil {
		t.Error("Error getting contents of working directory")
	}

	pathToEpubcheck := ""
	for _, i := range items {
		if i.Name() == testEpubcheckJarfile {
			pathToEpubcheck = i.Name()
			break

		} else if strings.HasPrefix(i.Name(), testEpubcheckPrefix) {
			if i.Mode().IsDir() {
				pathToEpubcheck = filepath.Join(i.Name(), testEpubcheckJarfile)
				if _, err := os.Stat(pathToEpubcheck); err == nil {
					break
				} else {
					pathToEpubcheck = ""
				}
			}
		}
	}

	if pathToEpubcheck == "" {
		if testing.Verbose() {
			fmt.Println("Epubcheck tool not installed, skipping EPUB validation.")
		}
		return nil, nil
	}

	cmd := exec.Command("java", "-jar", pathToEpubcheck, epubFilename)
	return cmd.CombinedOutput()
}

func writeAndExtractEpub(t testing.TB, e *Epub, epubFilename string) string {
	tempDir := uuid.Must(uuid.NewV4()).String()
	err := filesystem.Mkdir(tempDir, 0777)
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
