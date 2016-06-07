/*
Package epub generates valid EPUB 3.0 files with additional EPUB 2.0 table of
contents (as seen here: https://github.com/bmaupin/epub-samples) for maximum
compatibility.

Basic usage:

	// Create a new EPUB
	e := epub.NewEpub("My title")

	// Set the author
	e.SetAuthor("Hingle McCringleberry")

	// Add a section
	section1Body := `<h1>Section 1</h1>
	<p>This is a paragraph.</p>`
	e.AddSection("Section 1", section1Body, "", "")

	// Write the EPUB
	err = e.Write("My EPUB.epub")
	if err != nil {
		// handle error
	}

*/
package epub

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/satori/go.uuid"
)

// ErrFilenameAlreadyUsed is thrown by AddImage and AddSection if the same
// filename is used more than once
var ErrFilenameAlreadyUsed = errors.New("Filename already used")

const (
	cssFileFormat          = "css%04d.css"
	defaultCoverBody       = `<img src="%s" alt="Cover Image" />`
	defaultCoverCSSContent = `body {
  background-color: #FFFFFF;
  margin-bottom: 0px;
  margin-left: 0px;
  margin-right: 0px;
  margin-top: 0px;
  text-align: center;
}

img {
  max-height: 100%;
  max-width: 100%;
}`
	defaultCoverCSSFilename   = "cover.css"
	defaultCoverImgFormat     = "cover%s"
	defaultCoverXhtmlFilename = "cover.xhtml"
	defaultEpubLang           = "en"
	imageFileFormat           = "image%04d%s"
	sectionFileFormat         = "section%04d.xhtml"
	urnUUIDPrefix             = "urn:uuid:"
)

// Epub implements an EPUB file.
type Epub struct {
	author string
	cover  *epubCover
	// The key is the css filename, the value is the file content
	css map[string]string
	// The key is the image filename, the value is the image source
	images map[string]string
	// Language
	lang string
	// The package file (package.opf)
	pkg      *pkg
	sections []epubSection
	title    string
	// Table of contents
	toc  *toc
	uuid string
}

type epubCover struct {
	cssFilename   string
	imageFilename string
	xhtmlFilename string
}

type epubSection struct {
	filename string
	xhtml    *xhtml
}

// NewEpub returns a new Epub.
func NewEpub(title string) *Epub {
	e := &Epub{}
	e.cover = &epubCover{
		xhtmlFilename: "",
		imageFilename: "",
	}
	e.css = make(map[string]string)
	e.images = make(map[string]string)
	e.pkg = newPackage()
	e.toc = newToc()
	// Set minimal required attributes
	e.SetLang(defaultEpubLang)
	e.SetTitle(title)
	e.SetUUID(uuid.NewV4().String())

	return e
}

// AddCSS adds a new CSS file to the EPUB and returns a relative path to the
// CSS file that can be used in EPUB sections.
//
// The CSS content is the exact content that will be stored in the CSS file. It
// will not be validated.
//
// The CSS filename will be used when storing the CSS file in the EPUB and must
// be unique among all CSS files. If the same filename is used more than
// once, ErrFilenameAlreadyUsed will be returned. The CSS filename is
// optional; if no filename is provided, one will be generated.
func (e *Epub) AddCSS(cssFileContent string, cssFilename string) (string, error) {
	// Generate a filename if one isn't provided
	if cssFilename == "" {
		cssFilename = fmt.Sprintf(cssFileFormat, len(e.css)+1)
	}

	if _, ok := e.css[cssFilename]; ok {
		return "", ErrFilenameAlreadyUsed
	}

	e.css[cssFilename] = cssFileContent

	return filepath.Join(
		"..",
		cssFolderName,
		cssFilename,
	), nil
}

// AddImage adds an image to the EPUB and returns a relative path that can be
// used in the content of a section.
//
// The image source should either be a URL or a path to a local file; in either
// case, the image will be retrieved and stored in the EPUB.
//
// The image filename will be used when storing the image in the EPUB and must
// be unique among all image files. If the same filename is used more than once,
// ErrFilenameAlreadyUsed will be returned. The image filename is optional; if
// no filename is provided, one will be generated.
func (e *Epub) AddImage(imageSource string, imageFilename string) (string, error) {
	// Generate a filename if one isn't provided
	if imageFilename == "" {
		imageFilename = fmt.Sprintf(imageFileFormat, len(e.images)+1, filepath.Ext(imageSource))
	}

	if _, ok := e.images[imageFilename]; ok {
		return "", ErrFilenameAlreadyUsed
	}

	e.images[imageFilename] = imageSource

	return filepath.Join(
		"..",
		imageFolderName,
		imageFilename,
	), nil
}

// AddSection adds a new section (chapter, etc) to the EPUB and returns a
// relative path to the section that can be used from another section (for
// links).
//
// The title will be used for the table of contents.
//
// The body must be valid XHTML that will go between the <body> tags of the
// section XHTML file. The content will not be validated.
//
// The section filename will be used when storing the image in the EPUB and must
// be unique among all section files. If the same filename is used more than
// once, ErrFilenameAlreadyUsed will be returned. The section filename is
// optional; if no filename is provided, one will be generated.
//
// The path to the CSS file to be used for the section is optional.
//
// The section will be shown in the table of contents in the same order it was
// added to the EPUB.
func (e *Epub) AddSection(sectionTitle string, sectionBody string, sectionFilename string, cssPath string) (string, error) {
	// Generate a filename if one isn't provided
	if sectionFilename == "" {
		sectionFilename = fmt.Sprintf(sectionFileFormat, len(e.sections)+1)
	}

	for _, section := range e.sections {
		if section.filename == sectionFilename {
			return "", ErrFilenameAlreadyUsed
		}
	}

	x := newXhtml(sectionBody)
	x.setTitle(sectionTitle)

	if cssPath != "" {
		x.setCSS(cssPath)
	}

	s := epubSection{
		filename: sectionFilename,
		xhtml:    x,
	}
	e.sections = append(e.sections, s)

	return sectionFilename, nil
}

// Author returns the author of the EPUB.
func (e *Epub) Author() string {
	return e.author
}

// Lang returns the language of the EPUB.
func (e *Epub) Lang() string {
	return e.lang
}

// SetAuthor sets the author of the EPUB.
func (e *Epub) SetAuthor(author string) {
	e.author = author
	e.pkg.setAuthor(author)
}

// SetCover sets the cover page for the EPUB using the provided image source and
// optional CSS.
//
// The image source should either be a URL or a path to a local file; in either
// case, the image will be retrieved and stored in the EPUB.
//
// The CSS content is the exact content that will be stored in the CSS file. It
// will not be validated. If the CSS content isn't provided, default content
// will be used.
func (e *Epub) SetCover(imageSource string, cssFileContent string) {
	var err error

	// If a cover already exists
	if e.cover.xhtmlFilename != "" {
		// Remove the xhtml file
		for i, section := range e.sections {
			if section.filename == e.cover.xhtmlFilename {
				e.sections = append(e.sections[:i], e.sections[i+1:]...)
				break
			}
		}

		// Remove the image
		delete(e.images, e.cover.imageFilename)

		// Remove the CSS
		delete(e.css, e.cover.cssFilename)
	}

	defaultImageFilename := fmt.Sprintf(defaultCoverImgFormat, filepath.Ext(imageSource))
	imagePath, err := e.AddImage(imageSource, defaultImageFilename)
	// If that doesn't work, generate a filename
	if err != nil {
		imagePath, err = e.AddImage(imageSource, "")
		if err != nil {
			// This shouldn't cause an error since we're not specifying a filename
			panic(fmt.Sprintf("Error adding default cover image file: %s", err))
		}
	}
	e.cover.imageFilename = filepath.Base(imagePath)

	// Use default cover stylesheet if one isn't provided
	if cssFileContent == "" {
		cssFileContent = defaultCoverCSSContent
	}
	cssPath, err := e.AddCSS(cssFileContent, defaultCoverCSSFilename)
	// If that doesn't work, generate a filename
	if err != nil {
		cssPath, err = e.AddCSS(cssFileContent, "")
		if err != nil {
			// This shouldn't cause an error since we're not specifying a filename
			panic(fmt.Sprintf("Error adding default cover CSS file: %s", err))
		}
	}
	e.cover.cssFilename = filepath.Base(cssPath)

	coverBody := fmt.Sprintf(defaultCoverBody, imagePath)
	// Title won't be used since the cover won't be added to the TOC
	// First try to use the default cover filename
	coverPath, err := e.AddSection("", coverBody, defaultCoverXhtmlFilename, cssPath)
	// If that doesn't work, generate a filename
	if err != nil {
		coverPath, err = e.AddSection("", coverBody, "", cssPath)
		if err != nil {
			// This shouldn't cause an error since we're not specifying a filename
			panic(fmt.Sprintf("Error adding default cover XHTML file: %s", err))
		}
	}
	e.cover.xhtmlFilename = filepath.Base(coverPath)
}

// SetLang sets the language of the EPUB.
func (e *Epub) SetLang(lang string) {
	e.lang = lang
	e.pkg.setLang(lang)
}

// SetTitle sets the title of the EPUB.
func (e *Epub) SetTitle(title string) {
	e.title = title
	e.pkg.setTitle(title)
	e.toc.setTitle(title)
}

// SetUUID sets the UUID of the EPUB. A UUID will be automatically be generated
// for you when the NewEpub method is run.
func (e *Epub) SetUUID(uuid string) {
	e.uuid = uuid
	e.pkg.setUUID(urnUUIDPrefix + uuid)
	e.toc.setUUID(urnUUIDPrefix + uuid)
}

// Title returns the title of the EPUB.
func (e *Epub) Title() string {
	return e.title
}

// UUID returns the UUID of the EPUB.
func (e *Epub) UUID() string {
	return e.uuid
}
