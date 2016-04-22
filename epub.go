package epub

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/satori/go.uuid"
)

const (
	urnUuid = "urn:uuid:"
)

// Epub implements an EPUB file
type Epub struct {
	author string
	images map[string]string
	lang   string
	pkg    *pkg
	//	sections []section
	sections []xhtml
	title    string
	toc      *toc
	uuid     string
}

// NewWriter returns an Epub. A title must be provided as it is a required EPUB
// element. It can be changed using SetTitle.
func NewEpub(title string) (*Epub, error) {
	var err error

	e := &Epub{}
	e.images = make(map[string]string)
	e.pkg = newPackage()
	e.toc, err = newToc()
	if err != nil {
		return e, err
	}
	// Set minimal required attributes
	e.SetLang("en")
	e.SetTitle(title)
	e.SetUUID(urnUuid + uuid.NewV4().String())

	return e, nil
}

func (e *Epub) AddImage(imageSource string, imageFilename string) (string, error) {
	if _, ok := e.images[imageFilename]; ok {
		return "", errors.New(fmt.Sprintf("Image filename %s already used", imageFilename))
	}

	e.images[imageFilename] = imageSource

	return filepath.Join(
		"..",
		imageFolderName,
		imageFilename,
	), nil
}

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

// UUID returns the UUID of the EPUB
func (e *Epub) UUID() string {
	return e.uuid
}
