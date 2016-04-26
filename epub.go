package epub

import (
	"fmt"
	"path/filepath"

	"github.com/satori/go.uuid"
)

const (
	urnUUID = "urn:uuid:"
)

// Epub implements an EPUB file.
type Epub struct {
	author string
	images map[string]string // Images added to the EPUB
	lang   string            // Language
	pkg    *pkg              // The package file (package.opf)
	//	sections []section
	sections []xhtml // Sections (chapters)
	title    string
	toc      *toc // Table of contents
	uuid     string
}

// NewEpub returns a new Epub.
func NewEpub(title string) *Epub {
	e := &Epub{}
	e.images = make(map[string]string)
	e.pkg = newPackage()
	e.toc = newToc()
	// Set minimal required attributes
	e.SetLang("en")
	e.SetTitle(title)
	e.SetUUID(urnUUID + uuid.NewV4().String())

	return e
}

// AddImage adds an image to the EPUB and returns a relative path that can be
// used in the content of a section. The image source should either be a URL or
// a path to a local file; in either case, the image will be retrieved and
// stored in the EPUB. The image filename will be used when storing the image in
// the EPUB and must be unique.
func (e *Epub) AddImage(imageSource string, imageFilename string) (string, error) {
	if _, ok := e.images[imageFilename]; ok {
		return "", fmt.Errorf("Image filename %s already used", imageFilename)
	}

	e.images[imageFilename] = imageSource

	return filepath.Join(
		"..",
		imageFolderName,
		imageFilename,
	), nil
}

// AddSection adds a new section (chapter, etc) to the EPUB. The title will be
// used for the table of contents. The content must be valid XHTML that will go
// between the <body> tags. The content will not be validated.
func (e *Epub) AddSection(title string, content string) error {
	x, err := newXhtml(content)
	if err != nil {
		return err
	}
	x.setTitle(title)

	e.sections = append(e.sections, *x)

	return nil
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
	e.pkg.setUUID(uuid)
	e.toc.setUUID(uuid)
}

// Title returns the title of the EPUB.
func (e *Epub) Title() string {
	return e.title
}

// UUID returns the UUID of the EPUB.
func (e *Epub) UUID() string {
	return e.uuid
}
