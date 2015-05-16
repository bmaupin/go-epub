package epub

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	// TODO
	temp = `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" unique-identifier="pub-id" xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id">urn:uuid:fe93046f-af57-475a-a0cb-a0d4bc99ba6d</dc:identifier>
    <dc:title>Your title here</dc:title>
    <dc:language>en</dc:language>
    <meta refines="#creator" property="role" scheme="marc:relators" id="role">aut</meta>
    <meta property="dcterms:modified">2011-01-01T12:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav" />
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml" />
    <item id="section0001.xhtml" href="xhtml/section0001.xhtml" media-type="application/xhtml+xml" />
  </manifest>
  <spine toc="ncx">
    <itemref idref="section0001.xhtml" />
  </spine>
</package>
`
	pkgFileTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" unique-identifier="pub-id" xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id"></dc:identifier>
    <dc:title></dc:title>
    <dc:language></dc:language>
  </metadata>
  <manifest>
  </manifest>
  <spine toc="ncx">
  </spine>
</package>
`

	contentUniqueIdentifier = "pub-id"
	contentXmlnsDc          = "http://purl.org/dc/elements/1.1/"

	pkgModifiedProperty = "dcterms:modified"
)

type pkg struct {
	XMLName          xml.Name    `xml:"http://www.idpf.org/2007/opf package"`
	UniqueIdentifier string      `xml:"unique-identifier,attr"`
	Version          string      `xml:"version,attr"`
	Metadata         pkgMetadata `xml:"metadata"`
	Item             []pkgItem   `xml:"manifest>item"`
	Spine            pkgSpine    `xml:"spine"`
}

type pkgIdentifier struct {
	Id   string `xml:"id,attr"`
	Data string `xml:",chardata"`
}

type pkgItem struct {
	Href       string `xml:"href,attr"`
	Id         string `xml:"id,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type pkgItemref struct {
	Idref string `xml:"idref,attr"`
}

type pkgMeta struct {
	Property string `xml:"property,attr"`
	Data     string `xml:",chardata"`
}

type pkgMetadata struct {
	XmlnsDc    string        `xml:"xmlns:dc,attr"`
	Identifier pkgIdentifier `xml:"dc:identifier"`
	Title      string        `xml:"dc:title"`
	Language   string        `xml:"dc:language"`
	Creator    string        `xml:"dc:creator,omitempty"`
	Meta       []pkgMeta     `xml:"meta"`
}

type pkgSpine struct {
	Itemref []pkgItemref `xml:"itemref"`
}

func newPackage() *pkg {
	v := &pkg{
		Metadata: pkgMetadata{
			XmlnsDc: contentXmlnsDc,
			Identifier: pkgIdentifier{
				Id: contentUniqueIdentifier,
			},
		},
	}

	err := xml.Unmarshal([]byte(pkgFileTemplate), &v)
	if err != nil {
		log.Fatalf("xml.Unmarshal error: %s", err)
	}

	return v
}

func (p *pkg) setAuthor(author string) {
	p.Metadata.Creator = author
}

func (p *pkg) setLang(lang string) {
	p.Metadata.Language = lang
}

func (p *pkg) setModified(timestamp string) {
	//	p.Metadata.Meta.Data = timestamp
	m := pkgMeta{
		Data:     timestamp,
		Property: pkgModifiedProperty,
	}
	p.Metadata.Meta = append(p.Metadata.Meta, m)
}

func (p *pkg) setTitle(title string) {
	p.Metadata.Title = title
}

func (p *pkg) setUUID(uuid string) {
	p.Metadata.Identifier.Data = uuid
}

func (p *pkg) write(tempDir string) error {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	p.setModified(now)

	contentFolderPath := filepath.Join(tempDir, contentFolderName)
	if err := os.Mkdir(contentFolderPath, dirPermissions); err != nil {
		return err
	}

	pkgFilePath := filepath.Join(contentFolderPath, pkgFilename)

	output, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	// Add the xml header to the output
	pkgFileContent := append([]byte(xml.Header), output...)
	// It's generally nice to have files end with a newline
	pkgFileContent = append(pkgFileContent, "\n"...)

	if err := ioutil.WriteFile(pkgFilePath, []byte(pkgFileContent), filePermissions); err != nil {
		return err
	}

	return nil
}
