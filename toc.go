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
	xmlnsEpub      = "http://www.idpf.org/2007/ops"
)

type tocXmlNavLink struct {
	A struct {
		XMLName xml.Name `xml:"a"`
		Href    string   `xml:"href,attr"`
		Data    string   `xml:",chardata"`
	} `xml:a`
}

type tocXmlNav struct {
	XMLName  xml.Name        `xml:"nav"`
	EpubType string          `xml:"epub:type,attr"`
	H1       string          `xml:"h1"`
	Links    []tocXmlNavLink `xml:"ol>li"`
}

type toc struct {
	navDoc *xhtml
}

func newToc() (*toc, error) {
	t := &toc{}

	t.navDoc = &xhtml{}
	err := xml.Unmarshal([]byte(xhtmlTemplate), &t.navDoc)
	if err != nil {
		return t, err
	}

	t.navDoc.XmlnsEpub = xmlnsEpub

	n := &tocXmlNav{
		EpubType: navDocEpubType,
	}
	err = xml.Unmarshal([]byte(navDocBodyTemplate), &n)
	if err != nil {
		return t, err
	}

	navDocBodyContent, err := xml.MarshalIndent(n, "", `   `)
	if err != nil {
		return t, err
	}

	t.navDoc.setBody("\n" + string(navDocBodyContent) + "\n")

	return t, nil
}

func (t *toc) setTitle(title string) {
	t.navDoc.setTitle(title)
}

func (t *toc) write(tempDir string) error {
	navDocFilePath := filepath.Join(tempDir, contentFolderName, navDocFilename)

	navDocFileContent, err := xml.MarshalIndent(t.navDoc, "", `   `)
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
