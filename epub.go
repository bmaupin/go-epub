package epub

import (
    "github.com/satori/go.uuid"
)

type epub struct {
	lang string
	pkgdoc *pkgdoc
	title string
	toc *toc
	uuid string
}

func NewEpub(title string) (*epub, error) {
	var err error

	e := &epub{}
	e.pkgdoc = newPkgdoc()
	e.toc, err = newToc()
	// Set minimal required attributes
	e.SetLang("en")
	e.SetTitle(title)
	e.SetUUID(uuid.NewV4().String())
	
	return e, err
}

func (e *epub) Lang() string {
	return e.lang
}

func (e *epub) SetLang(lang string) {
	e.lang = lang
	e.pkgdoc.setLang(lang)
}

func (e *epub) SetTitle(title string) {
	e.title = title
	e.pkgdoc.setTitle(title)
}

func (e *epub) SetUUID(uuid string) {
	e.uuid = uuid
	e.pkgdoc.setUUID(uuid)
}

func (e *epub) Title() string {
	return e.title
}

func (e *epub) Uuid() string {
	return e.uuid
}
