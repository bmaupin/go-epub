Work in progress.

See the latest documentation here: https://godoc.org/github.com/bmaupin/go-epub

**Features:**

- Clean API
- Creates valid EPUB 3.0 files
- Adds an additional EPUB 2.0 table of contents ([as seen here](https://github.com/bmaupin/epub-samples)) for maximum compatibility

**Basic usage:**

    // Create a new EPUB
	e:= epub.NewEpub("My title")

    // Set the author
	e.SetAuthor("Hingle McCringleberry")

    // Add a section
	section1Content := `    <h1>Section 1</h1>
    <p>This is a paragraph.</p>`
	e.AddSection("Section 1", section1Content)

	section2Content := `    <h1>Section 2</h1>
    <p>This is a paragraph.</p>`
	e.AddSection("Section 2", section2Content)

    // Write the EPUB
	err = e.Write("My EPUB.epub")
	if err != nil {
		fmt.Println("epub.Write error: %s", err)
	}

**Wishlist:**

- Clean up error handling
- Add support for cover pages
- Add more documentation
- Add tests
- Add support for CSS
- Add functionality to read EPUB files
- Add [examples](https://golang.org/pkg/testing/#hdr-Examples)
