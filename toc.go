package epub

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
)

const (
	tocNavBodyTemplate = `
    <nav epub:type="toc">
      <h1>Table of Contents</h1>
      <ol>
      </ol>
    </nav>
`
	tocNavFilename       = "nav.xhtml"
	tocNavItemID         = "nav"
	tocNavItemProperties = "nav"
	tocNavEpubType       = "toc"

	tocNcxFilename = "toc.ncx"
	tocNcxItemID   = "ncx"
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

// toc implements the EPUB table of contents
type toc struct {
	// This holds the body XML for the EPUB v3 TOC file (nav.xhtml). Since this is
	// an XHTML file, the rest of the structure is handled by the xhtml type
	//
	// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v32/EPUB/nav.xhtml
	// Spec: http://www.idpf.org/epub/301/spec/epub-contentdocs.html#sec-xhtml-nav
	navXML *tocNavBody

	// This holds the XML for the EPUB v2 TOC file (toc.ncx). This is added so the
	// resulting EPUB v3 file will still work with devices that only support EPUB v2
	//
	// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v32/EPUB/toc.ncx
	// Spec: http://www.idpf.org/epub/20/spec/OPF_2.0.1_draft.htm#Section2.4.1
	ncxXML *tocNcxRoot

	title string // EPUB title
}

type tocNavBody struct {
	XMLName  xml.Name     `xml:"nav"`
	EpubType string       `xml:"epub:type,attr"`
	H1       string       `xml:"h1"`
	Links    []tocNavItem `xml:"ol>li"`
}

type tocNavItem struct {
	A tocNavLink `xml:a`
}

type tocNavLink struct {
	XMLName xml.Name `xml:"a"`
	Href    string   `xml:"href,attr"`
	Data    string   `xml:",chardata"`
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
	ID      string        `xml:"id,attr"`
	Text    string        `xml:"navLabel>text"`
	Content tocNcxContent `xml:"content"`
}

// Constructor for toc
func newToc() *toc {
	t := &toc{}

	t.navXML = newTocNavXML()

	t.ncxXML = newTocNcxXML()

	return t
}

// Constructor for tocNavBody
func newTocNavXML() *tocNavBody {
	b := &tocNavBody{
		EpubType: tocNavEpubType,
	}
	err := xml.Unmarshal([]byte(tocNavBodyTemplate), &b)
	if err != nil {
		panic(fmt.Sprintf(
			"Error unmarshalling tocNavBody: %s\n"+
				"\ttocNavBody=%#v\n"+
				"\ttocNavBodyTemplate=%s",
			err,
			*b,
			tocNavBodyTemplate))
	}

	return b
}

// Constructor for tocNcxRoot
func newTocNcxXML() *tocNcxRoot {
	n := &tocNcxRoot{}

	err := xml.Unmarshal([]byte(tocNcxTemplate), &n)
	if err != nil {
		panic(fmt.Sprintf(
			"Error unmarshalling tocNcxRoot: %s\n"+
				"\ttocNcxRoot=%#v\n"+
				"\ttocNcxTemplate=%s",
			err,
			*n,
			tocNcxTemplate))
	}

	return n
}

// Add a section to the TOC (navXML as well as ncxXML)
func (t *toc) addSection(index int, title string, relativePath string) {
	l := &tocNavItem{
		A: tocNavLink{
			Href: relativePath,
			Data: title,
		},
	}
	t.navXML.Links = append(t.navXML.Links, *l)

	np := &tocNcxNavPoint{
		ID:   "navPoint-" + strconv.Itoa(index),
		Text: title,
		Content: tocNcxContent{
			Src: relativePath,
		},
	}
	t.ncxXML.NavMap = append(t.ncxXML.NavMap, *np)
}

func (t *toc) setTitle(title string) {
	t.title = title
}

func (t *toc) setUUID(uuid string) {
	t.ncxXML.Meta.Content = uuid
}

// Write the TOC files
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

// Write the the EPUB v3 TOC file (nav.xhtml) to the temporary directory
func (t *toc) writeNavDoc(tempDir string) error {
	navBodyContent, err := xml.MarshalIndent(t.navXML, "    ", "  ")
	if err != nil {
		return err
	}

	n := newXhtml(string(navBodyContent))
	n.setXmlnsEpub(xmlnsEpub)
	n.setTitle(t.title)

	navFilePath := filepath.Join(tempDir, contentFolderName, tocNavFilename)
	n.write(navFilePath)

	return nil
}

// Write the EPUB v2 TOC file (toc.ncx) to the temporary directory
func (t *toc) writeNcxDoc(tempDir string) error {
	t.ncxXML.Title = t.title

	ncxFileContent, err := xml.MarshalIndent(t.ncxXML, "", "  ")
	if err != nil {
		return err
	}

	// Add the xml header to the output
	ncxFileContent = append([]byte(xml.Header), ncxFileContent...)
	// It's generally nice to have files end with a newline
	ncxFileContent = append(ncxFileContent, "\n"...)

	ncxFilePath := filepath.Join(tempDir, contentFolderName, tocNcxFilename)
	if err := ioutil.WriteFile(ncxFilePath, []byte(ncxFileContent), filePermissions); err != nil {
		return err
	}

	return nil
}
