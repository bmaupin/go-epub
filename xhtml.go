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

type xhtmlRoot struct {
	XMLName   xml.Name      `xml:"http://www.w3.org/1999/xhtml html"`
	XmlnsEpub string        `xml:"xmlns:epub,attr,omitempty"`
	Title     string        `xml:"head>title"`
	Body      xhtmlInnerxml `xml:"body"`
}

type xhtmlInnerxml struct {
	XML string `xml:",innerxml"`
}

func (x *xhtmlRoot) setBody(body string) {
	x.Body.XML = body
}

func (x *xhtmlRoot) setTitle(title string) {
	x.Title = title
}
