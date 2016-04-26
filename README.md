[![Go Report Card](http://goreportcard.com/badge/bmaupin/go-epub)](http://goreportcard.com/report/bmaupin/go-epub)

[![GoDoc](https://godoc.org/github.com/bmaupin/go-epub?status.svg)](https://godoc.org/github.com/bmaupin/go-epub)

Work in progress.

**Features:**

- [Clean API](https://godoc.org/github.com/bmaupin/go-epub)
- Creates valid EPUB 3.0 files
- Adds an additional EPUB 2.0 table of contents ([as seen here](https://github.com/bmaupin/epub-samples)) for maximum compatibility

**Basic usage:**

    // Create a new EPUB
	e := epub.NewEpub("My title")

    // Set the author
	e.SetAuthor("Hingle McCringleberry")

    // Add a section
	section1Content := `    <h1>Section 1</h1>
    <p>This is a paragraph.</p>`
	err := e.AddSection("Section 1", section1Content)
    if err != nil {
        // handle error
    }

	section2Content := `    <h1>Section 2</h1>
    <p>This is a paragraph.</p>`
	err = e.AddSection("Section 2", section2Content)
    if err != nil {
        // handle error
    }

    // Write the EPUB
	err = e.Write("My EPUB.epub")
	if err != nil {
		// handle error
	}

**Wishlist:**

- Clean up error handling
- Add support for cover pages
- Add tests
- Add support for CSS
- Add functionality to read EPUB files
