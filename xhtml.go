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

func (x *xhtml) setBody(body string) {
	x.xml.Body.XML = body
}

func (x *xhtml) setTitle(title string) {
	x.xml.Title = title
}

func newXhtml(title string, content string) (*xhtml, error) {
	r, err := newXhtmlRoot(content)
	if err != nil {
		return nil, err
	}

	x := &xhtml{
		xml: r,
	}

	xhtmlBodyContent, err := xml.MarshalIndent(content, "    ", "  ")
	if err != nil {
		return x, err
	}

	x.setBody("\n" + string(xhtmlBodyContent) + "\n")
	x.setTitle(title)

	return x, nil
}

func newXhtmlRoot(content string) (*xhtmlRoot, error) {
	r := &xhtmlRoot{}
	err := xml.Unmarshal([]byte(xhtmlTemplate), &r)
	if err != nil {
		return r, err
	}

	return r, nil
}
