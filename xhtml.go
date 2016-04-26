package epub

import (
	"encoding/xml"
	"io/ioutil"
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

// xhtml implements an XHTML document
type xhtml struct {
	xml *xhtmlRoot
}

// This holds the actual XHTML content
type xhtmlRoot struct {
	XMLName   xml.Name      `xml:"http://www.w3.org/1999/xhtml html"`
	XmlnsEpub string        `xml:"xmlns:epub,attr,omitempty"`
	Title     string        `xml:"head>title"`
	Body      xhtmlInnerxml `xml:"body"`
}

// This holds the content of the XHTML document between the <body> tags. It is
// implemented as a string because we don't know what it will contain and we
// leave it up to the user of the package to validate the content
type xhtmlInnerxml struct {
	XML string `xml:",innerxml"`
}

// Constructor for xhtml
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

// Constructor for xhtmlRoot
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

func (x *xhtml) Title() string {
	return x.xml.Title
}

// Write the XHTML file to the specified path
func (x *xhtml) write(xhtmlFilePath string) error {
	xhtmlFileContent, err := xml.MarshalIndent(x.xml, "", "  ")
	if err != nil {
		return err
	}
	// Add the doctype declaration to the output
	xhtmlFileContent = append([]byte(xhtmlDoctype), xhtmlFileContent...)
	// Add the xml header to the output
	xhtmlFileContent = append([]byte(xml.Header), xhtmlFileContent...)
	// It's generally nice to have files end with a newline
	xhtmlFileContent = append(xhtmlFileContent, "\n"...)

	if err := ioutil.WriteFile(xhtmlFilePath, []byte(xhtmlFileContent), filePermissions); err != nil {
		return err
	}

	return nil
}
