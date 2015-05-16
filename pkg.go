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
	pkgAuthorId       = "role"
	pkgAuthorData     = "aut"
	pkgAuthorProperty = "role"
	pkgAuthorRefines  = "#creator"
	pkgAuthorScheme   = "marc:relators"
	pkgCreatorId      = "creator"
	pkgFileTemplate   = `<?xml version="1.0" encoding="UTF-8"?>
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
	pkgModifiedProperty = "dcterms:modified"
	pkgUniqueIdentifier = "pub-id"

	xmlnsDc = "http://purl.org/dc/elements/1.1/"
)

type pkg struct {
	xml          *pkgRoot
	authorMeta   *pkgMeta
	modifiedMeta *pkgMeta
}

type pkgRoot struct {
	XMLName          xml.Name    `xml:"http://www.idpf.org/2007/opf package"`
	UniqueIdentifier string      `xml:"unique-identifier,attr"`
	Version          string      `xml:"version,attr"`
	Metadata         pkgMetadata `xml:"metadata"`
	Item             []pkgItem   `xml:"manifest>item"`
	Spine            pkgSpine    `xml:"spine"`
}

type pkgCreator struct {
	XMLName xml.Name `xml:"dc:creator"`
	Id      string   `xml:"id,attr"`
	Data    string   `xml:",chardata"`
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
	Refines  string `xml:"refines,attr,omitempty"`
	Property string `xml:"property,attr"`
	Scheme   string `xml:"scheme,attr,omitempty"`
	Id       string `xml:"id,attr,omitempty"`
	Data     string `xml:",chardata"`
}

type pkgMetadata struct {
	XmlnsDc    string        `xml:"xmlns:dc,attr"`
	Identifier pkgIdentifier `xml:"dc:identifier"`
	Title      string        `xml:"dc:title"`
	Language   string        `xml:"dc:language"`
	Creator    *pkgCreator
	Meta       []pkgMeta `xml:"meta"`
}

type pkgSpine struct {
	Itemref []pkgItemref `xml:"itemref"`
}

func newPackage() *pkg {
	p := &pkg{
		xml: &pkgRoot{
			Metadata: pkgMetadata{
				XmlnsDc: xmlnsDc,
				Identifier: pkgIdentifier{
					Id: pkgUniqueIdentifier,
				},
			},
		},
	}

	err := xml.Unmarshal([]byte(pkgFileTemplate), &p.xml)
	if err != nil {
		log.Fatalf("xml.Unmarshal error: %s", err)
	}

	return p
}

func (p *pkg) setAuthor(author string) {
	p.xml.Metadata.Creator = &pkgCreator{
		Data: author,
		Id:   pkgCreatorId,
	}
	p.authorMeta = &pkgMeta{
		Data:     pkgAuthorData,
		Id:       pkgAuthorId,
		Property: pkgAuthorProperty,
		Refines:  pkgAuthorRefines,
		Scheme:   pkgAuthorScheme,
	}

	p.xml.Metadata.Meta = updateMeta(p.xml.Metadata.Meta, p.authorMeta)
}

func (p *pkg) setLang(lang string) {
	p.xml.Metadata.Language = lang
}

func (p *pkg) setModified(timestamp string) {
	//	var indexToReplace int

	p.modifiedMeta = &pkgMeta{
		Data:     timestamp,
		Property: pkgModifiedProperty,
	}

	p.xml.Metadata.Meta = updateMeta(p.xml.Metadata.Meta, p.modifiedMeta)
}

func (p *pkg) setTitle(title string) {
	p.xml.Metadata.Title = title
}

func (p *pkg) setUUID(uuid string) {
	p.xml.Metadata.Identifier.Data = uuid
}

func updateMeta(a []pkgMeta, m *pkgMeta) []pkgMeta {
	indexToReplace := -1

	if len(a) > 0 {
		// If we've already added the modified meta element to the meta array
		for i, meta := range a {
			if meta == *m {
				indexToReplace = i
				break
			}
		}
	}

	// If the array is empty or the meta element isn't in it
	if indexToReplace == -1 {
		// Add the meta element to the array of meta elements
		a = append(a, *m)

		// If the meta element is found
	} else {
		// Replace it
		a[indexToReplace] = *m
	}

	return a
}

func (p *pkg) write(tempDir string) error {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	p.setModified(now)

	contentFolderPath := filepath.Join(tempDir, contentFolderName)
	if err := os.Mkdir(contentFolderPath, dirPermissions); err != nil {
		return err
	}

	pkgFilePath := filepath.Join(contentFolderPath, pkgFilename)

	output, err := xml.MarshalIndent(p.xml, "", "  ")
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
