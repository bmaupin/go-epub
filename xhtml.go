package epub

import (
	"encoding/xml"
)

const (
	xhtmlDoctype = `<!DOCTYPE html>
`
	xhtmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title></title>
  </head>
  <body></body>
</html>
`
)

type xhtml struct {
	xml *xhtmlRoot
}

type xhtmlRoot struct {
	XMLName   xml.Name      `xml:"http://www.w3.org/1999/xhtml html"`
	XmlnsEpub string        `xml:"xmlns:epub,attr,omitempty"`
	Title     string        `xml:"head>title"`
	Body      xhtmlInnerxml `xml:"body"`
}

type xhtmlInnerxml struct {
	XML string `xml:",innerxml"`
}

func newXhtml(content string) (*xhtml, error) {
	var x *xhtml

	r, err := newXhtmlRoot()
	if err != nil {
		return x, err
	}

	x = &xhtml{
		xml: r,
	}

	x.setBody(content)

	return x, nil
}

func newXhtmlRoot() (*xhtmlRoot, error) {
	r := &xhtmlRoot{}
	err := xml.Unmarshal([]byte(xhtmlTemplate), &r)
	if err != nil {
		return r, err
	}

	return r, nil
}

func (x *xhtml) setBody(body string) {
	x.xml.Body.XML = "\n" + body + "\n"
}

func (x *xhtml) setTitle(title string) {
	x.xml.Title = title
}

func (x *xhtml) setXmlnsEpub(xmlns string) {
	x.xml.XmlnsEpub = xmlns
}
