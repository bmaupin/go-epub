package epub

import (
  "encoding/xml"
  "log"
)

const (
	navBodyTemplate = `    <nav epub:type="toc">
      <h1>Table of Contents</h1>
      <ol>
        <li><a href="xhtml/section0001.xhtml">Section 1</a></li>
      </ol>
    </nav>
`
	navDocFilename = "nav.xhtml"
	xmlnsEpub      = `xmlns:epub="http://www.idpf.org/2007/ops"`
)

type tocXmlNav struct {
    XMLName xml.Name `xml:"http://www.w3.org/1999/xhtml html"`
    H1 string `xml:"h1"`
//    A tocXmlNav
}

type toc struct {
	navDoc *xhtml
}

func newToc() (*toc, error) {
	t := &toc{}
	t.navDoc = &xhtml{}

  output, err := xml.MarshalIndent(t.navDoc, "", `   `)
  log.Println(string(output))

  err = xml.Unmarshal([]byte(xhtmlTemplate), &t.navDoc)
  if err != nil {
    log.Println("xml.Unmarshal error: %s", err)
  }

  output, err = xml.MarshalIndent(t.navDoc, "", `   `)
  log.Println(string(output))

  t.navDoc.setBody(navBodyTemplate)

  output, err = xml.MarshalIndent(t.navDoc, "", `   `)
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