package epub

import (
    "github.com/satori/go.uuid"
)

type Epub struct {
	filename string
	pkgdoc *Pkgdoc
	title string
	uuid uuid.UUID
}

func NewEpub(title string) *Epub {
	e := &Epub{}
	
	e.title = title
	e.uuid = uuid.NewV4()
	e.pkgdoc = NewPkgdoc()
	e.pkgdoc.Metadata.Identifier.Data = e.uuid.String()
	
	return e
}

func (e *Epub) Uuid() uuid.UUID {
	return e.uuid
}

