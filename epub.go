package epub

import (
    "github.com/satori/go.uuid"
)

type Epub struct {
	Filename string
	Pkgdoc *Pkgdoc
	Title string
	Uuid uuid.UUID
}

func NewEpub(title string) *Epub {
	e := &Epub{}
	
	e.Title = title
	e.Uuid = uuid.NewV4()
	e.Pkgdoc = NewPkgdoc()
	e.Pkgdoc.Metadata.Identifier.Data = e.Uuid.String()
	
	return e
}

func (e *Epub) Write() {
	Write(e, e.Title + ".epub")
}
