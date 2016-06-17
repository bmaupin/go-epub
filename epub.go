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
	e.AddSection(section1Body, "Section 1", "", "")

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
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/satori/go.uuid"
)

// ErrFilenameAlreadyUsed is thrown by AddImage and AddSection if the same
// filename is used more than once
var ErrFilenameAlreadyUsed = errors.New("Filename already used")

// ErrRetrievingFile is thrown by AddCSS, AddFont, AddImage, or SetCover if
// there was a problem retrieving the source file that was provided
var ErrRetrievingFile = errors.New("Error retrieving file from source")

// Folder names used for resources inside the EPUB
const (
	CSSFolderName   = "css"
	FontFolderName  = "fonts"
	ImageFolderName = "images"
)

const (
	cssFileFormat             = "css%04d%s"
	defaultCoverBody          = `<img src="%s" alt="Cover Image" />`
	defaultCoverCSSFilename   = "cover.css"
	defaultCoverCSSSource     = "cover.css"
	defaultCoverImgFormat     = "cover%s"
	defaultCoverXhtmlFilename = "cover.xhtml"
	defaultEpubLang           = "en"
	fontFileFormat            = "font%04d%s"
	imageFileFormat           = "image%04d%s"
	sectionFileFormat         = "section%04d.xhtml"
	urnUUIDPrefix             = "urn:uuid:"
)

// Epub implements an EPUB file.
type Epub struct {
	author string
	cover  *epubCover
	// The key is the css filename, the value is the css source
	css map[string]string
	// The key is the font filename, the value is the font source
	fonts map[string]string
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
	e.fonts = make(map[string]string)
	e.images = make(map[string]string)
	e.pkg = newPackage()
	e.toc = newToc()
	// Set minimal required attributes
	e.SetLang(defaultEpubLang)
	e.SetTitle(title)
	e.SetUUID(uuid.NewV4().String())

	return e
}

// AddCSS adds a CSS file to the EPUB and returns a relative path to the CSS
// file that can be used in EPUB sections in the format:
// ../CSSFolderName/internalFilename
//
// The CSS source should either be a URL or a path to a local file; in either
// case, the CSS file will be retrieved and stored in the EPUB.
//
// The internal filename will be used when storing the CSS file in the EPUB
// and must be unique among all CSS files. If the same filename is used more
// than once, ErrFilenameAlreadyUsed will be returned. The internal filename is
// optional; if no filename is provided, one will be generated.
func (e *Epub) AddCSS(source string, internalFilename string) (string, error) {
	return addMedia(source, internalFilename, cssFileFormat, CSSFolderName, e.css)
}

// AddFont adds a font file to the EPUB and returns a relative path to the font
// file that can be used in EPUB sections in the format:
// ../FontFolderName/internalFilename
//
// The font source should either be a URL or a path to a local file; in either
// case, the font file will be retrieved and stored in the EPUB.
//
// The internal filename will be used when storing the font file in the EPUB
// and must be unique among all font files. If the same filename is used more
// than once, ErrFilenameAlreadyUsed will be returned. The internal filename is
// optional; if no filename is provided, one will be generated.
func (e *Epub) AddFont(source string, internalFilename string) (string, error) {
	return addMedia(source, internalFilename, fontFileFormat, FontFolderName, e.fonts)
}

// AddImage adds an image to the EPUB and returns a relative path to the image
// file that can be used in EPUB sections in the format:
// ../ImageFolderName/internalFilename
//
// The image source should either be a URL or a path to a local file; in either
// case, the image file will be retrieved and stored in the EPUB.
//
// The internal filename will be used when storing the image file in the EPUB
// and must be unique among all image files. If the same filename is used more
// than once, ErrFilenameAlreadyUsed will be returned. The internal filename is
// optional; if no filename is provided, one will be generated.
func (e *Epub) AddImage(source string, imageFilename string) (string, error) {
	return addMedia(source, imageFilename, imageFileFormat, ImageFolderName, e.images)
}

// AddSection adds a new section (chapter, etc) to the EPUB and returns a
// relative path to the section that can be used from another section (for
// links).
//
// The body must be valid XHTML that will go between the <body> tags of the
// section XHTML file. The content will not be validated.
//
// The title will be used for the table of contents. The section will be shown
// in the table of contents in the same order it was added to the EPUB. The
// title is optional; if no title is provided, the section will not be added to
// the table of contents.
//
// The internal filename will be used when storing the section file in the EPUB
// and must be unique among all section files. If the same filename is used more
// than once, ErrFilenameAlreadyUsed will be returned. The internal filename is
// optional; if no filename is provided, one will be generated.
//
// The internal path to an already-added CSS file (as returned by AddCSS) to be
// used for the section is optional.
func (e *Epub) AddSection(body string, sectionTitle string, internalFilename string, internalCSSPath string) (string, error) {
	// Generate a filename if one isn't provided
	if internalFilename == "" {
		internalFilename = fmt.Sprintf(sectionFileFormat, len(e.sections)+1)
	}

	for _, section := range e.sections {
		if section.filename == internalFilename {
			return "", ErrFilenameAlreadyUsed
		}
	}

	x := newXhtml(body)
	x.setTitle(sectionTitle)

	if internalCSSPath != "" {
		x.setCSS(internalCSSPath)
	}

	s := epubSection{
		filename: internalFilename,
		xhtml:    x,
	}
	e.sections = append(e.sections, s)

	return internalFilename, nil
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
func (e *Epub) SetCover(imageSource string, cssSource string) error {
	var err error

	// Make sure the source files are valid before proceeding
	if isFileSourceValid(imageSource) == false {
		return ErrRetrievingFile
	} else if isFileSourceValid(cssSource) == false {
		return ErrRetrievingFile
	}

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

	defaultImageFilename := fmt.Sprintf(
		defaultCoverImgFormat,
		strings.ToLower(filepath.Ext(imageSource)),
	)
	imagePath, err := e.AddImage(imageSource, defaultImageFilename)
	// If that doesn't work, generate a filename
	if err == ErrFilenameAlreadyUsed {
		imagePath, err = e.AddImage(imageSource, "")
		if err == ErrFilenameAlreadyUsed {
			// This shouldn't cause an error since we're not specifying a filename
			panic(fmt.Sprintf("Error adding default cover image file: %s", err))
		}
	}
	if err != nil && err != ErrFilenameAlreadyUsed {
		return err
	}
	e.cover.imageFilename = filepath.Base(imagePath)

	// Use default cover stylesheet if one isn't provided
	if cssSource == "" {
		cssSource = defaultCoverCSSSource
	}
	cssPath, err := e.AddCSS(cssSource, defaultCoverCSSFilename)
	// If that doesn't work, generate a filename
	if err == ErrFilenameAlreadyUsed {
		cssPath, err = e.AddCSS(cssSource, "")
		if err == ErrFilenameAlreadyUsed {
			// This shouldn't cause an error since we're not specifying a filename
			panic(fmt.Sprintf("Error adding default cover CSS file: %s", err))
		}
	}
	if err != nil && err != ErrFilenameAlreadyUsed {
		return err
	}
	e.cover.cssFilename = filepath.Base(cssPath)

	coverBody := fmt.Sprintf(defaultCoverBody, imagePath)
	// Title won't be used since the cover won't be added to the TOC
	// First try to use the default cover filename
	coverPath, err := e.AddSection(coverBody, "", defaultCoverXhtmlFilename, cssPath)
	// If that doesn't work, generate a filename
	if err == ErrFilenameAlreadyUsed {
		coverPath, err = e.AddSection(coverBody, "", "", cssPath)
		if err == ErrFilenameAlreadyUsed {
			// This shouldn't cause an error since we're not specifying a filename
			panic(fmt.Sprintf("Error adding default cover XHTML file: %s", err))
		}
	}
	if err != nil && err != ErrFilenameAlreadyUsed {
		return err
	}
	e.cover.xhtmlFilename = filepath.Base(coverPath)

	return nil
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

// Add a media file to the EPUB and return the path relative to the EPUB section
// files
func addMedia(source string, internalFilename string, mediaFileFormat string, mediaFolderName string, mediaMap map[string]string) (string, error) {
	// Make sure the source file is valid before proceeding
	if isFileSourceValid(source) == false {
		return "", ErrRetrievingFile
	}

	if internalFilename == "" {
		// If a filename isn't provided, use the filename from the source
		internalFilename = filepath.Base(source)
		// If that's already used, try to generate a unique filename
		if _, ok := mediaMap[internalFilename]; ok {
			internalFilename = fmt.Sprintf(
				mediaFileFormat,
				len(mediaMap)+1,
				strings.ToLower(filepath.Ext(source)),
			)
		}
	}

	if _, ok := mediaMap[internalFilename]; ok {
		return "", ErrFilenameAlreadyUsed
	}

	mediaMap[internalFilename] = source

	return filepath.Join(
		"..",
		mediaFolderName,
		internalFilename,
	), nil
}

func isFileSourceValid(source string) bool {
	u, err := url.Parse(source)
	if err != nil {
		return false
	}

	var r io.ReadCloser
	var resp *http.Response
	// If it's a URL
	if u.Scheme == "http" || u.Scheme == "https" {
		resp, err = http.Get(source)
		if err != nil {
			return false
		}
		r = resp.Body

		// Otherwise, assume it's a local file
	} else {
		r, err = os.Open(source)
	}
	if err != nil {
		return false
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	return true
}
