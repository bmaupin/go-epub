package epub_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-shiori/go-epub"
)

func ExampleEpub_AddCSS() {
	var t *testing.T
	e, err := epub.NewEpub("My title")
	if err != nil {
		t.Error(err)
	}

	// Add CSS
	css1Path, err := e.AddCSS("testdata/cover.css", "epub.css")
	if err != nil {
		log.Println(err)
		return
	}

	// The filename is optional
	css2Path, err := e.AddCSS("testdata/cover.css", "")
	if err != nil {
		log.Println(err)
		return
	}

	// Use the CSS in a section
	sectionBody := `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	_, err = e.AddSection(sectionBody, "Section 1", "", css1Path)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(css1Path)
	fmt.Println(css2Path)

	// Output:
	// ../css/epub.css
	// ../css/cover.css
}

func ExampleEpub_AddFont() {
	var t *testing.T
	e, err := epub.NewEpub("My title")
	if err != nil {
		t.Error(err)
	}

	// Add a font from a local file
	font1Path, err := e.AddFont("testdata/redacted-script-regular.ttf", "font.ttf")
	if err != nil {
		log.Println(err)
	}

	// The filename is optional
	font2Path, err := e.AddFont("testdata/redacted-script-regular.ttf", "")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(font1Path)
	fmt.Println(font2Path)

	// Output:
	// ../fonts/font.ttf
	// ../fonts/redacted-script-regular.ttf
}

func ExampleEpub_AddImage() {
	fs := http.FileServer(http.Dir("./testdata/"))

	// start a test server with the file server handler
	server := httptest.NewServer(fs)
	defer server.Close()

	testImageFromURLSource := server.URL + "/gophercolor16x16.png"

	var t *testing.T
	e, err := epub.NewEpub("My title")
	if err != nil {
		t.Error(err)
	}

	// Add an image from a local file
	img1Path, err := e.AddImage("testdata/gophercolor16x16.png", "go-gopher.png")
	if err != nil {
		log.Println(err)
	}

	// Add an image from a URL. The filename is optional
	img2Path, err := e.AddImage(testImageFromURLSource, "")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(img1Path)
	fmt.Println(img2Path)

	// Output:
	// ../images/go-gopher.png
	// ../images/gophercolor16x16.png
}

func ExampleEpub_AddSection() {
	var t *testing.T
	e, err := epub.NewEpub("My title")
	if err != nil {
		t.Error(err)
	}

	// Add a section. The CSS path is optional
	section1Body := `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	section1Path, err := e.AddSection(section1Body, "Section 1", "firstsection.xhtml", "")
	if err != nil {
		log.Println(err)
	}

	// Link to the first section
	section2Body := fmt.Sprintf(`    <h1>Section 2</h1>
	<a href="%s">Link to section 1</a>`,
		section1Path)
	// The title and filename are also optional
	section2Path, err := e.AddSection(section2Body, "", "", "")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(section1Path)
	fmt.Println(section2Path)

	// Output:
	// firstsection.xhtml
	// section0001.xhtml
}

func ExampleEpub_AddSubSection() {
	var t *testing.T
	e, err := epub.NewEpub("My title")
	if err != nil {
		t.Error(err)
	}

	// Add a section. The CSS path is optional
	section1Body := `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	section1Path, err := e.AddSection(section1Body, "Section 1", "firstsection.xhtml", "")
	if err != nil {
		log.Println(err)
	}

	// Link to the first section
	section2Body := fmt.Sprintf(`    <h1>Section 2</h1>
	<a href="%s">Link to section 1</a>`,
		section1Path)
	// The title and filename are also optional
	section2Path, err := e.AddSubSection(section1Path, section2Body, "", "", "")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(section1Path)
	fmt.Println(section2Path)

	// Output:
	// firstsection.xhtml
	// section0001.xhtml
}

func ExampleEpub_SetCover() {
	var t *testing.T
	e, err := epub.NewEpub("My title")
	if err != nil {
		t.Error(err)
	}

	// Set the cover. The CSS file is optional
	coverImagePath, _ := e.AddImage("testdata/gophercolor16x16.png", "cover.png")
	err = e.SetCover(coverImagePath, "")
	if err != nil {
		t.Error(err)
	}

	// Update the cover using custom CSS
	coverCSSPath, _ := e.AddCSS("testdata/cover.css", "")
	err = e.SetCover(coverImagePath, coverCSSPath)
	if err != nil {
		t.Error(err)
	}
}

func ExampleEpub_SetIdentifier() {
	var t *testing.T
	e, err := epub.NewEpub("My title")
	if err != nil {
		t.Error(err)
	}

	// Set the identifier to a UUID
	e.SetIdentifier("urn:uuid:a1b0d67e-2e81-4df5-9e67-a64cbe366809")

	// Set the identifier to an ISBN
	e.SetIdentifier("urn:isbn:9780101010101")
}
