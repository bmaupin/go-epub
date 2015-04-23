package epub

import (
    "github.com/satori/go.uuid"
)

type Epub struct {
	lang string
	pkgdoc *Pkgdoc
	title string
	uuid uuid.UUID
}

func NewEpub(title string) *Epub {
	e := &Epub{}
	
	e.lang = "en"
	e.title = title
	e.uuid = uuid.NewV4()
	e.pkgdoc = NewPkgdoc()
	e.pkgdoc.Metadata.Identifier.Data = e.uuid.String()
	
	return e
}

func (e *Epub) Uuid() uuid.UUID {
	return e.uuid
}

func (e *Epub) Lang() string {
	return e.lang
}

func (e *Epub) SetLang(lang string) {
	e.lang = lang
}
