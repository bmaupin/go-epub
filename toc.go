package epub

import (
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
)

const (
	tocNavBodyTemplate = `
    <nav epub:type="toc">
      <h1>Table of Contents</h1>
      <ol>
      </ol>
    </nav>
`
	tocNavFilename = "nav.xhtml"
	tocNavEpubType = "toc"

	tocNcxFilename = "toc.ncx"
	tocNcxTemplate = `
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="" />
  </head>
  <docTitle>
    <text></text>
  </docTitle>
  <navMap>
  </navMap>
</ncx>`

	xmlnsEpub = "http://www.idpf.org/2007/ops"
)

type toc struct {
	navDoc *xhtml
	ncxXml *tocNcxRoot
}

type tocNavBody struct {
	XMLName  xml.Name     `xml:"nav"`
	EpubType string       `xml:"epub:type,attr"`
	H1       string       `xml:"h1"`
	Links    []tocNavLink `xml:"ol>li"`
}

type tocNavLink struct {
	A struct {
		XMLName xml.Name `xml:"a"`
		Href    string   `xml:"href,attr"`
		Data    string   `xml:",chardata"`
	} `xml:a`
}

type tocNcxRoot struct {
	XMLName xml.Name         `xml:"http://www.daisy.org/z3986/2005/ncx/ ncx"`
	Version string           `xml:"version,attr"`
	Meta    tocNcxMeta       `xml:"head>meta"`
	Title   string           `xml:"docTitle>text"`
	NavMap  []tocNcxNavPoint `xml:"navMap>navPoint"`
}

type tocNcxContent struct {
	Src string `xml:"src,attr"`
}

type tocNcxMeta struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type tocNcxNavPoint struct {
	XMLName xml.Name      `xml:"navPoint"`
	Id      string        `xml:"id,attr"`
	Text    string        `xml:"navLabel>text"`
	Content tocNcxContent `xml:"content"`
}

func newToc() (*toc, error) {
	var err error

	t := &toc{}

	t.navDoc, err = newTocNavDoc()
	if err != nil {
		return t, err
	}

	t.ncxXml, err = newTocNcxXml()
	if err != nil {
		return t, err
	}

	return t, nil
}

func newTocNavDoc() (*xhtml, error) {
	var err error
	var n *xhtml

	b := &tocNavBody{
		EpubType: tocNavEpubType,
	}
	err = xml.Unmarshal([]byte(tocNavBodyTemplate), &b)
	if err != nil {
		return n, err
	}

	navBodyContent, err := xml.MarshalIndent(b, "    ", "  ")
	if err != nil {
		return n, err
	}

	n, err = newXhtml(string(navBodyContent))

	n.setXmlnsEpub(xmlnsEpub)

	return n, err
}

func newTocNcxXml() (*tocNcxRoot, error) {
	n := &tocNcxRoot{}

	err := xml.Unmarshal([]byte(tocNcxTemplate), &n)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (t *toc) setTitle(title string) {
	t.navDoc.setTitle(title)
	t.ncxXml.Title = title
}

func (t *toc) setUUID(uuid string) {
	t.ncxXml.Meta.Content = uuid
}

func (t *toc) write(tempDir string) error {
	navFilePath := filepath.Join(tempDir, contentFolderName, tocNavFilename)
	if err := t.navDoc.write(navFilePath); err != nil {
		return err
	}

	err := t.writeNcxDoc(tempDir)
	if err != nil {
		return err
	}

	return nil
}

func (t *toc) writeNcxDoc(tempDir string) error {
	ncxFilePath := filepath.Join(tempDir, contentFolderName, tocNcxFilename)

	ncxFileContent, err := xml.MarshalIndent(t.ncxXml, "", "  ")
	if err != nil {
		return err
	}
	// Add the xml header to the output
	ncxFileContent = append([]byte(xml.Header), ncxFileContent...)
	// It's generally nice to have files end with a newline
	ncxFileContent = append(ncxFileContent, "\n"...)

	if err := ioutil.WriteFile(ncxFilePath, []byte(ncxFileContent), filePermissions); err != nil {
		return err
	}

	return nil
}
