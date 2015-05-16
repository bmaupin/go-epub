package epub

import (
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
)

const (
	navDocBodyTemplate = `
    <nav epub:type="toc">
      <h1>Table of Contents</h1>
      <ol>
      </ol>
    </nav>
`
	navDocFilename = "nav.xhtml"
	navDocEpubType = "toc"

	ncxDocFilename = "toc.ncx"
	ncxDocTemplate = `
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
	ncxDoc *tocNcx
}

type tocNav struct {
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

type tocNcx struct {
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

	t.ncxDoc, err = newTocNcxDoc()
	if err != nil {
		return t, err
	}

	return t, nil
}

func newTocNavDoc() (*xhtml, error) {
	navDoc := &xhtml{
		XmlnsEpub: xmlnsEpub,
	}
	err := xml.Unmarshal([]byte(xhtmlTemplate), &navDoc)
	if err != nil {
		return navDoc, err
	}

	n := &tocNav{
		EpubType: navDocEpubType,
	}
	err = xml.Unmarshal([]byte(navDocBodyTemplate), &n)
	if err != nil {
		return navDoc, err
	}

	navDocBodyContent, err := xml.MarshalIndent(n, "    ", "  ")
	if err != nil {
		return navDoc, err
	}

	navDoc.setBody("\n" + string(navDocBodyContent) + "\n")

	return navDoc, nil
}

func newTocNcxDoc() (*tocNcx, error) {
	ncxDoc := &tocNcx{}

	err := xml.Unmarshal([]byte(ncxDocTemplate), &ncxDoc)
	if err != nil {
		return ncxDoc, err
	}

	return ncxDoc, nil
}

func (t *toc) setTitle(title string) {
	t.navDoc.setTitle(title)
	t.ncxDoc.Title = title
}

func (t *toc) setUUID(uuid string) {
	t.ncxDoc.Meta.Content = uuid
}

func (t *toc) write(tempDir string) error {
	err := t.writeNavDoc(tempDir)
	if err != nil {
		return err
	}
	err = t.writeNcxDoc(tempDir)
	if err != nil {
		return err
	}

	return nil
}

func (t *toc) writeNavDoc(tempDir string) error {
	navDocFilePath := filepath.Join(tempDir, contentFolderName, navDocFilename)

	navDocFileContent, err := xml.MarshalIndent(t.navDoc, "", "  ")
	if err != nil {
		return err
	}
	// Add the doctype declaration to the output
	navDocFileContent = append([]byte(xhtmlDoctype), navDocFileContent...)
	// Add the xml header to the output
	navDocFileContent = append([]byte(xml.Header), navDocFileContent...)
	// It's generally nice to have files end with a newline
	navDocFileContent = append(navDocFileContent, "\n"...)

	if err := ioutil.WriteFile(navDocFilePath, []byte(navDocFileContent), filePermissions); err != nil {
		return err
	}

	return nil
}

func (t *toc) writeNcxDoc(tempDir string) error {
	ncxDocFilePath := filepath.Join(tempDir, contentFolderName, ncxDocFilename)

	ncxDocFileContent, err := xml.MarshalIndent(t.ncxDoc, "", "  ")
	if err != nil {
		return err
	}
	// Add the xml header to the output
	ncxDocFileContent = append([]byte(xml.Header), ncxDocFileContent...)
	// It's generally nice to have files end with a newline
	ncxDocFileContent = append(ncxDocFileContent, "\n"...)

	if err := ioutil.WriteFile(ncxDocFilePath, []byte(ncxDocFileContent), filePermissions); err != nil {
		return err
	}

	return nil
}
