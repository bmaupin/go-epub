package epub_test

import (
	"fmt"
	"log"

	"github.com/bmaupin/go-epub"
)

func ExampleEpub_AddCSS() {
	e := epub.NewEpub("My title")

	// Add CSS
	css1Path, err := e.AddCSS("cover.css", "epub.css")
	if err != nil {
		log.Fatal(err)
	}

	// The filename is optional
	css2Path, err := e.AddCSS("cover.css", "")
	if err != nil {
		log.Fatal(err)
	}

	// Use the CSS in a section
	sectionBody := `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	e.AddSection(sectionBody, "Section 1", "", css1Path)

	fmt.Println(css1Path)
	fmt.Println(css2Path)

	// Output:
	// ../css/epub.css
	// ../css/cover.css
}

func ExampleEpub_AddFont() {
	e := epub.NewEpub("My title")

	// Add a font from a local file
	font1Path, err := e.AddFont("testdata/redacted-script-regular.ttf", "font.ttf")
	if err != nil {
		log.Fatal(err)
	}

	// The filename is optional
	font2Path, err := e.AddFont("testdata/redacted-script-regular.ttf", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(font1Path)
	fmt.Println(font2Path)

	// Output:
	// ../fonts/font.ttf
	// ../fonts/redacted-script-regular.ttf
}

func ExampleEpub_AddImage() {
	e := epub.NewEpub("My title")

	// Add an image from a local file
	img1Path, err := e.AddImage("testdata/gophercolor16x16.png", "go-gopher.png")
	if err != nil {
		log.Fatal(err)
	}

	// Add an image from a URL. The filename is optional
	img2Path, err := e.AddImage("https://golang.org/doc/gopher/gophercolor16x16.png", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(img1Path)
	fmt.Println(img2Path)

	// Output:
	// ../images/go-gopher.png
	// ../images/gophercolor16x16.png
}

func ExampleEpub_AddSection() {
	e := epub.NewEpub("My title")

	// Add a section. The CSS path is optional
	section1Body := `    <h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	section1Path, err := e.AddSection(section1Body, "Section 1", "firstsection.xhtml", "")
	if err != nil {
		log.Fatal(err)
	}

	// Link to the first section
	section2Body := fmt.Sprintf(`    <h1>Section 2</h1>
	<a href="%s">Link to section 1</a>`,
		section1Path)
	// The title and filename are also optional
	section2Path, err := e.AddSection(section2Body, "", "", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(section1Path)
	fmt.Println(section2Path)

	// Output:
	// firstsection.xhtml
	// section0002.xhtml
}

func ExampleEpub_SetCover() {
	e := epub.NewEpub("My title")

	// Set the cover. The CSS content is optional
	e.SetCover("testdata/gophercolor16x16.png", "")

	// Update the cover using custom CSS
	coverCSSContent := `h1 {
  text-align: center;
}`
	e.SetCover("testdata/gophercolor16x16.png", coverCSSContent)
}
