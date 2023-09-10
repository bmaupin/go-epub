package epub

import (
	"encoding/xml"
	"fmt"
)

const (
	xhtmlDoctype = `<!DOCTYPE html>
`
	xhtmlLinkRel  = "stylesheet"
	xhtmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title dir="auto"></title>
  </head>
  <body dir="auto"></body>
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
	Head      xhtmlHead     `xml:"head"`
	Body      xhtmlInnerxml `xml:"body"`
}

type xhtmlHead struct {
	Title xhtmlTitle `xml:"title"`
	Link  *xhtmlLink
}

type xhtmlTitle struct {
	XMLName xml.Name `xml:"title,omitempty"`
	Dir     string   `xml:"dir,attr,omitempty"`
	Value   string   `xml:",chardata"`
}

// The <link> element, used to link to stylesheets
// Ex: <link rel="stylesheet" type="text/css" href="../css/epub.css" />
type xhtmlLink struct {
	XMLName xml.Name `xml:"link,omitempty"`
	Rel     string   `xml:"rel,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	Href    string   `xml:"href,attr,omitempty"`
}

// This holds the content of the XHTML document between the <body> tags. It is
// implemented as a string because we don't know what it will contain and we
// leave it up to the user of the package to validate the content
type xhtmlInnerxml struct {
	XML string `xml:",innerxml"`
	Dir string `xml:"dir,attr,omitempty"`
}

// Constructor for xhtml
func newXhtml(body string) (*xhtml, error) {
	xmlroot, err := newXhtmlRoot()
	if err != nil {
		return nil, fmt.Errorf("can't create newXhtmlRoot because of: %w", err)
	}
	x := &xhtml{
		xml: xmlroot,
	}
	x.setBody(body)

	return x, nil
}

// Constructor for xhtmlRoot
func newXhtmlRoot() (*xhtmlRoot, error) {
	r := &xhtmlRoot{
		Body: xhtmlInnerxml{Dir: "auto"},
	}
	err := xml.Unmarshal([]byte(xhtmlTemplate), &r)

	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling xhtmlRoot: %w\n"+"\txhtmlRoot=%#v\n"+"\txhtmlTemplate=%s", err, *r, xhtmlTemplate)
	}
	return r, nil
}

func (x *xhtml) setBody(body string) {
	x.xml.Body.XML = "\n" + body + "\n"
	x.xml.Body.Dir = "auto"
}

func (x *xhtml) setCSS(path string) {
	x.xml.Head.Link = &xhtmlLink{
		Rel:  xhtmlLinkRel,
		Type: mediaTypeCSS,
		Href: path,
	}
}

func (x *xhtml) setTitle(title string) {
	x.xml.Head.Title = xhtmlTitle{
		Dir:   "auto",
		Value: title,
	}
}

func (x *xhtml) setXmlnsEpub(xmlns string) {
	x.xml.XmlnsEpub = xmlns
}

func (x *xhtml) Title() string {
	return x.xml.Head.Title.Value
}

// Write the XHTML file to the specified path
func (x *xhtml) write(xhtmlFilePath string) error {
	xhtmlFileContent, err := xml.MarshalIndent(x.xml, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshalling XML for XHTML file: %w\n"+"\tXML=%v", err, x.xml)
	}

	// Add the doctype declaration to the output
	xhtmlFileContent = append([]byte(xhtmlDoctype), xhtmlFileContent...)
	// Add the xml header to the output
	xhtmlFileContent = append([]byte(xml.Header), xhtmlFileContent...)
	// It's generally nice to have files end with a newline
	xhtmlFileContent = append(xhtmlFileContent, "\n"...)

	if err := filesystem.WriteFile(xhtmlFilePath, []byte(xhtmlFileContent), filePermissions); err != nil {
		return fmt.Errorf("Error writing XHTML file: %w", err)
	}
	return nil
}
