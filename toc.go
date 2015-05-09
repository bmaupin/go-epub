package epub

import (
	"encoding/xml"
	"log"
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
	xmlnsEpub      = `xmlns:epub="http://www.idpf.org/2007/ops"`
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
		log.Println("xml.Unmarshal error: %s", err)
	}

	n := &tocXmlNav{
		EpubType: navDocEpubType,
	}
	err = xml.Unmarshal([]byte(navDocBodyTemplate), &n)
	if err != nil {
		log.Println("xml.Unmarshal error: %s", err)
	}

	navDocBodyContent, err := xml.MarshalIndent(n, "", `   `)
	if err != nil {
		log.Println("xml.Unmarshal error: %s", err)
	}

	t.navDoc.setBody(string(navDocBodyContent))

// TODO
	output, err := xml.MarshalIndent(t.navDoc, "", `   `)
	log.Println(string(output))

	return t, err
}

/*
func (t *toc) write() {
	contentFolderPath := filepath.Join(tempDir, contentFolderName)

	navDocFilePath := filepath.Join(contentFolderPath, navDocFilename)

	output, err := xml.MarshalIndent(e.pkgdoc, "", `   `)
	if err != nil {
		return err
	}
	// Add the xml header to the output
	pkgdocFileContent := append([]byte(xml.Header), output...)
	// It's generally nice to have files end with a newline
	pkgdocFileContent = append(pkgdocFileContent, "\n"...)

	if err := ioutil.WriteFile(pkgdocFilePath, []byte(pkgdocFileContent), filePermissions); err != nil {
		return err
	}

	return nil
}
*/
