package epub

import (
    "encoding/xml"
    "log"
)

const (
// TODO
	temp = `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" unique-identifier="pub-id" xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id">urn:uuid:fe93046f-af57-475a-a0cb-a0d4bc99ba6d</dc:identifier>
    <dc:title>Your title here</dc:title>
    <dc:language>en</dc:language>
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
	packageFileTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" unique-identifier="pub-id" xmlns="http://www.idpf.org/2007/opf">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="pub-id"></dc:identifier>
    <dc:title></dc:title>
    <dc:language></dc:language>
    <meta property="dcterms:modified"></meta>
  </metadata>
  <manifest>
  </manifest>
  <spine toc="ncx">
  </spine>
</package>
`

	contentUniqueIdentifier = "pub-id"
	contentXmlnsDc = "http://purl.org/dc/elements/1.1/"
)

type itemref struct {
    Idref string `xml:"idref,attr"`
}

type spine struct {
    Itemref []itemref `xml:"itemref"`
}

type item struct {
    Href string `xml:"href,attr"`
    Id string `xml:"id,attr"`
    MediaType string `xml:"media-type,attr"`
    Properties string `xml:"properties,attr"`
}

type meta struct {
	Property string `xml:"property,attr"`
	Data string `xml:",chardata"`
}

type identifier struct {
    Id string `xml:"id,attr"`
    Data string `xml:",chardata"`
}

type metadata struct {
    XmlnsDc string `xml:"xmlns:dc,attr"`
    Identifier identifier `xml:"dc:identifier"`
    Title string `xml:"title"`
    Language string `xml:"language"`
    Meta meta `xml:"meta"`
}

type pkgdoc struct {
    XMLName xml.Name `xml:"http://www.idpf.org/2007/opf package"`
    UniqueIdentifier string `xml:"unique-identifier,attr"`
    Version string `xml:"version,attr"`
    Metadata metadata `xml:"metadata"`
    Item []item `xml:"manifest>item"`
    Spine spine `xml:"spine"`
}

func newPkgdoc() *pkgdoc {
	v := &pkgdoc{}

	err := xml.Unmarshal([]byte(packageFileTemplate), &v)
	if err != nil {
		log.Fatalf("xml.Unmarshal error: %s", err)
	}
	
	v.Metadata.XmlnsDc = contentXmlnsDc
	v.Metadata.Identifier.Id = contentUniqueIdentifier
    
    return v
}

func (p *pkgdoc) setLang(lang string) {
	p.Metadata.Language = lang
}

func (p *pkgdoc) setModified(timestamp string) {
	p.Metadata.Meta.Data = timestamp
}

func (p *pkgdoc) setTitle(title string) {
  p.Metadata.Title = title
}

func (p *pkgdoc) setUUID(uuid string) {
	p.Metadata.Identifier.Data = uuid
}
