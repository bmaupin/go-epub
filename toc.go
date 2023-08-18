package epub

import (
	"encoding/xml"
	"fmt"
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
    <meta name="dtb:depth" content="" />
  </head>
  <docTitle>
    <text></text>
  </docTitle>
  <docAuthor>
    <text></text>
  </docAuthor>
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
	// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v3plus2/EPUB/nav.xhtml
	// Spec: http://www.idpf.org/epub/301/spec/epub-contentdocs.html#sec-xhtml-nav
	navXML *tocNavBody

	// This holds the XML for the EPUB v2 TOC file (toc.ncx). This is added so the
	// resulting EPUB v3 file will still work with devices that only support EPUB v2
	//
	// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v3plus2/EPUB/toc.ncx
	// Spec: http://www.idpf.org/epub/20/spec/OPF_2.0.1_draft.htm#Section2.4.1
	ncxXML *tocNcxRoot

	title  string // EPUB title
	author string // EPUB author
}

type tocNavBody struct {
	XMLName  xml.Name     `xml:"nav"`
	EpubType string       `xml:"epub:type,attr"`
	H1       string       `xml:"h1"`
	Links    []tocNavItem `xml:"ol>li"`
}

type tocNavItem struct {
	A        tocNavLink    `xml:"a"`
	Children *[]tocNavItem `xml:"ol>li,omitempty"`
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
	Author  string           `xml:"docAuthor>text"`
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
	XMLName  xml.Name          `xml:"navPoint"`
	ID       string            `xml:"id,attr"`
	Text     string            `xml:"navLabel>text"`
	Content  tocNcxContent     `xml:"content"`
	Children *[]tocNcxNavPoint `xml:"navPoint,omitempty"`
}

// Constructor for toc
func newToc() (*toc, error) {
	t := &toc{}
	var err error

	t.navXML, err = newTocNavXML()
	if err != nil {
		return nil, fmt.Errorf("can't create navXML %w", err)
	}

	t.ncxXML, err = newTocNcxXML()
	if err != nil {
		return nil, fmt.Errorf("can't create navXML becuse of %w", err)
	}

	return t, nil
}

// Constructor for tocNavBody
func newTocNavXML() (*tocNavBody, error) {
	b := &tocNavBody{
		EpubType: tocNavEpubType,
	}
	err := xml.Unmarshal([]byte(tocNavBodyTemplate), &b)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling tocNavBody: %w\n"+"\ttocNavBody=%#v\n"+"\ttocNavBodyTemplate=%s", err, *b, tocNavBodyTemplate)
	}

	return b, nil
}

// Constructor for tocNcxRoot
func newTocNcxXML() (*tocNcxRoot, error) {
	n := &tocNcxRoot{}

	err := xml.Unmarshal([]byte(tocNcxTemplate), &n)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling tocNcxRoot: %w\n"+"\ttocNcxRoot=%#v\n"+"\ttocNcxTemplate=%s", err, *n, tocNcxTemplate)
	}

	return n, nil
}

// Add a section to the TOC (navXML as well as ncxXML)
func (t *toc) addSection(index int, title string, relativePath string) {
	relativePath = filepath.ToSlash(relativePath)
	l := &tocNavItem{
		A: tocNavLink{
			Href: relativePath,
			Data: title,
		},
		Children: nil,
	}
	t.navXML.Links = append(t.navXML.Links, *l)

	np := &tocNcxNavPoint{
		ID:   "navPoint-" + strconv.Itoa(index),
		Text: title,
		Content: tocNcxContent{
			Src: relativePath,
		},
		Children: nil,
	}
	t.ncxXML.NavMap = append(t.ncxXML.NavMap, *np)
}

// Add a sub section to the TOC (navXML as well as ncxXML)
func (t *toc) addSubSection(parent string, index int, title string, relativePath string) {
	var parentNcxIndex int
	var parentNavIndex int

	relativePath = filepath.ToSlash(relativePath)
	parent = filepath.ToSlash(parent)

	for index, nav := range t.navXML.Links {
		if nav.A.Href == parent {
			parentNavIndex = index
		}
	}
	l := tocNavItem{
		A: tocNavLink{
			Href: relativePath,
			Data: title,
		},
	}
	if len(t.navXML.Links) > parentNavIndex {
		// Create a new array if none exists
		if t.navXML.Links[parentNavIndex].Children == nil {
			n := make([]tocNavItem, 0)
			t.navXML.Links[parentNavIndex].Children = &n
		}
		children := append(*t.navXML.Links[parentNavIndex].Children, l)
		t.navXML.Links[parentNavIndex].Children = &children
	} else {
		t.navXML.Links = append(t.navXML.Links, l)
	}

	// Get parent object
	for index, ncx := range t.ncxXML.NavMap {
		if ncx.Content.Src == parent {
			parentNcxIndex = index
		}
	}
	np := tocNcxNavPoint{
		ID:   "navPoint-" + strconv.Itoa(index),
		Text: title,
		Content: tocNcxContent{
			Src: relativePath,
		},
		Children: nil,
	}
	if parentNcxIndex > len(t.ncxXML.NavMap) {
		if t.ncxXML.NavMap[parentNcxIndex].Children == nil {
			n := make([]tocNcxNavPoint, 0)
			t.ncxXML.NavMap[parentNcxIndex].Children = &n
		}
		children := append(*t.ncxXML.NavMap[parentNcxIndex].Children, np)
		t.ncxXML.NavMap[parentNcxIndex].Children = &children
	} else {
		t.ncxXML.NavMap = append(t.ncxXML.NavMap, np)
	}
}

func (t *toc) setIdentifier(identifier string) {
	t.ncxXML.Meta.Content = identifier
}

func (t *toc) setTitle(title string) {
	t.title = title
}

func (t *toc) setAuthor(author string) {
	t.author = author
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
		return fmt.Errorf("Error marshalling XML for EPUB v3 TOC file: %w\n"+"\tXML=%#v", err, t.navXML)
	}

	n, err := newXhtml(string(navBodyContent))
	if err != nil {
		return fmt.Errorf("can't create xhtml for TOC file becuase of: %w", err)
	}
	n.setXmlnsEpub(xmlnsEpub)
	n.setTitle(t.title)

	navFilePath := filepath.Join(tempDir, contentFolderName, tocNavFilename)
	err = n.write(navFilePath)
	if err != nil {
		return fmt.Errorf("can't write TOC file becuase of: %w", err)
	}
	return nil
}

// Write the EPUB v2 TOC file (toc.ncx) to the temporary directory
func (t *toc) writeNcxDoc(tempDir string) error {
	t.ncxXML.Title = t.title
	t.ncxXML.Author = t.author

	ncxFileContent, err := xml.MarshalIndent(t.ncxXML, "", "  ")
	if err != nil {
		return fmt.Errorf("Error marshalling XML for EPUB v2 TOC file: %w\n"+"+\tXML=%#v", err, t.ncxXML)
	}

	// Add the xml header to the output
	ncxFileContent = append([]byte(xml.Header), ncxFileContent...)
	// It's generally nice to have files end with a newline
	ncxFileContent = append(ncxFileContent, "\n"...)

	ncxFilePath := filepath.Join(tempDir, contentFolderName, tocNcxFilename)
	if err := filesystem.WriteFile(ncxFilePath, []byte(ncxFileContent), filePermissions); err != nil {
		return fmt.Errorf("Error writing EPUB v2 TOC file: %w", err)
	}
	return nil
}
