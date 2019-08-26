[![Go Report Card](https://goreportcard.com/badge/github.com/speedata/go-epub)](https://goreportcard.com/report/github.com/speedata/go-epub)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/speedata/go-epub/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/speedata/go-epub?status.svg)](https://godoc.org/github.com/speedata/go-epub)
---

### Features
- [Documented API](https://godoc.org/github.com/speedata/go-epub)
- Creates valid EPUB 3.0 files
- Adds an additional EPUB 2.0 table of contents ([as seen here](https://github.com/bmaupin/epub-samples)) for maximum compatibility
- Includes support for adding CSS, images, and fonts

For an example of actual usage, see https://github.com/bmaupin/go-docs-epub

### Installation

    go get github.com/speedata/go-epub

### Development

    go get github.com/speedata/go-epub
    cd $GOPATH/src/github.com/speedata/go-epub

Dependencies are managed using [Go modules](https://github.com/golang/go/wiki/Modules)

### Testing

1. (Optional) Install EpubCheck

       wget https://github.com/w3c/epubcheck/releases/download/v4.2.2/epubcheck-4.2.2.zip
       unzip epubcheck-4.2.2.zip

2. Run tests

       go test

## Credits

This is a fork of https://github.com/bmaupin/go-epub. So all credits and kudos should go there and any bugs are most likeliy introduced by us.

This fork allows subsections to be added to the navigation.
