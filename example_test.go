package epub_test

import (
	"fmt"
	"log"

	"github.com/bmaupin/go-epub"
)

func ExampleEpub_AddImage() {
	// Create a new EPUB
	e := epub.NewEpub("My title")

	// Add an image from a local file
	img1Path, err := e.AddImage("testdata/gophercolor16x16.png", "go-gopher.png")
	if err != nil {
		log.Fatal(err)
	}

	// Add an image from a URL. The image filename is also optional
	img2Path, err := e.AddImage("https://golang.org/doc/gopher/gophercolor16x16.png", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(img1Path)
	fmt.Println(img2Path)

	// Output:
	// ../img/go-gopher.png
	// ../img/image0002.png
}
