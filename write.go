package epub

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const (
	containerFilename     = "container.xml"
	containerFileTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="%s/%s" media-type="application/oebps-package+xml" />
  </rootfiles>
</container>
`
	contentFolderName = "EPUB"
	// Permissions for any new directories we create
	dirPermissions = 0755
	// Permissions for any new files we create
	filePermissions   = 0644
	imageFolderName   = "img"
	mediaTypeNcx      = "application/x-dtbncx+xml"
	mediaTypeEpub     = "application/epub+zip"
	mediaTypeXhtml    = "application/xhtml+xml"
	metaInfFolderName = "META-INF"
	mimetypeFilename  = "mimetype"
	pkgFilename       = "package.opf"
	sectionFileFormat = "section%04d.xhtml"
	tempDirPrefix     = "go-epub"
	xhtmlFolderName   = "xhtml"
)

// Write writes the EPUB file. The destination path must be the full path to
// the resulting file, including filename and extension.
func (e *Epub) Write(destFilePath string) error {
	tempDir, err := ioutil.TempDir("", tempDirPrefix)
	defer os.Remove(tempDir)
	if err != nil {
		log.Fatalf("os.Remove error: %s", err)
	}

	// Must be called first
	err = createEpubFolders(tempDir)
	if err != nil {
		return err
	}

	// Must be called after:
	// createEpubFolders()
	err = e.writeImages(tempDir)
	if err != nil {
		return err
	}

	// Must be called after:
	// createEpubFolders()
	err = writeMimetype(tempDir)
	if err != nil {
		return err
	}

	// Must be called after:
	// createEpubFolders()
	err = e.writeSections(tempDir)
	if err != nil {
		return err
	}

	// Must be called after:
	// createEpubFolders()
	// writeSections()
	err = e.writeToc(tempDir)
	if err != nil {
		return err
	}

	// Must be called after:
	// createEpubFolders()
	err = writeContainerFile(tempDir)
	if err != nil {
		return err
	}

	// Must be called after:
	// createEpubFolders()
	// writeImages()
	// writeSections()
	// writeToc()
	err = e.writePackageFile(tempDir)
	if err != nil {
		return err
	}

	// Must be called last
	err = e.writeEpub(tempDir, destFilePath)
	if err != nil {
		return err
	}

	// TODO

	//	output, err := xml.MarshalIndent(e.toc.navDoc.xml, "", "  ")
	//	output = append([]byte(xhtmlDoctype), output...)

	//	output, err := xml.MarshalIndent(e.pkg.xml, "", "  ")

	//  output, err := xml.MarshalIndent(e.toc.ncxXML, "", "  ")
	//	output = append([]byte(xml.Header), output...)
	//	fmt.Println(string(output))

	return nil
}

// Create the EPUB folder structure in a temp directory
func createEpubFolders(tempDir string) error {
	if err := os.Mkdir(
		filepath.Join(
			tempDir,
			contentFolderName,
		),
		dirPermissions); err != nil {
		return err
	}

	if err := os.Mkdir(
		filepath.Join(
			tempDir,
			contentFolderName,
			xhtmlFolderName,
		),
		dirPermissions); err != nil {
		return err
	}

	if err := os.Mkdir(
		filepath.Join(
			tempDir,
			metaInfFolderName,
		),
		dirPermissions); err != nil {
		return err
	}

	return nil
}

// Write the contatiner file (container.xml), which mostly just points to the
// package file (package.opf)
//
// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v32/META-INF/container.xml
// Spec: http://www.idpf.org/epub/301/spec/epub-ocf.html#sec-container-metainf-container.xml
func writeContainerFile(tempDir string) error {
	containerFilePath := filepath.Join(tempDir, metaInfFolderName, containerFilename)
	if err := ioutil.WriteFile(
		containerFilePath,
		[]byte(
			fmt.Sprintf(
				containerFileTemplate,
				contentFolderName,
				pkgFilename,
			),
		),
		filePermissions,
	); err != nil {
		return err
	}

	return nil
}

// Write the EPUB file itself by zipping up everything from a temp directory
func (e *Epub) writeEpub(tempDir string, destFilePath string) error {
	f, err := os.Create(destFilePath)
	if err != nil {
		log.Fatalf("os.Create error: %s", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("os.File.Close error: %s", err)
		}
	}()

	z := zip.NewWriter(f)
	defer func() {
		if err := z.Close(); err != nil {
			log.Fatalf("zip.Writer.Close error: %s", err)
		}
	}()

	skipMimetypeFile := false

	var addFileToZip = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the path of the file relative to the folder we're zipping
		relativePath, err := filepath.Rel(tempDir, path)
		if err != nil {
			log.Fatalf("filepath.Rel error: %s", err)
		}

		// Only include regular files, not directories
		if !info.Mode().IsRegular() {
			return nil
		}

		var w io.Writer
		if path == filepath.Join(tempDir, mimetypeFilename) {
			// Skip the mimetype file if it's already been written
			if skipMimetypeFile == true {
				return nil
			}
			// The mimetype file must be uncompressed according to the EPUB spec
			w, err = z.CreateHeader(&zip.FileHeader{
				Name:   relativePath,
				Method: zip.Store,
			})
		} else {
			w, err = z.Create(relativePath)
		}
		if err != nil {
			log.Fatalf("zip.Writer.Create error: %s", err)
		}

		r, err := os.Open(path)
		if err != nil {
			log.Fatalf("os.Open error: %s", err)
		}
		defer func() {
			if err := r.Close(); err != nil {
				log.Fatalf("os.File.Close error: %s", err)
			}
		}()

		_, err = io.Copy(w, r)
		if err != nil {
			log.Fatalf("io.Copy error: %s", err)
		}

		return nil
	}

	// Add the mimetype file first
	mimetypeFilePath := filepath.Join(tempDir, mimetypeFilename)
	mimetypeInfo, err := os.Lstat(mimetypeFilePath)
	if err != nil {
		log.Fatalf("os.Lstat error: %s", err)
	}
	addFileToZip(mimetypeFilePath, mimetypeInfo, nil)

	skipMimetypeFile = true

	err = filepath.Walk(tempDir, addFileToZip)
	if err != nil {
		log.Fatalf("os.Lstat error: %s", err)
	}

	return nil
}

// Get images from their source and save them in the temporary directory
func (e *Epub) writeImages(tempDir string) error {
	imageFolderPath := filepath.Join(tempDir, contentFolderName, imageFolderName)
	if err := os.Mkdir(imageFolderPath, dirPermissions); err != nil {
		return err
	}

	for imageFilename, imageSource := range e.images {
		u, err := url.Parse(imageSource)
		if err != nil {
			return err
		}

		var r io.ReadCloser
		var resp *http.Response
		// If it's a URL
		if u.Scheme == "http" || u.Scheme == "https" {
			resp, err = http.Get(imageSource)
			r = resp.Body

			// Otherwise, assume it's a local file
		} else {
			r, err = os.Open(imageSource)
		}
		if err != nil {
			return err
		}
		defer func() {
			err = r.Close()
		}()

		imageFilePath := filepath.Join(
			imageFolderPath,
			imageFilename,
		)

		w, err := os.Create(imageFilePath)
		if err != nil {
			return err
		}
		defer func() {
			err = w.Close()
		}()

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
	}

	return nil
}

// Write the mimetype file
//
// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v32/mimetype
// Spec: http://www.idpf.org/epub/301/spec/epub-ocf.html#sec-zip-container-mime
func writeMimetype(tempDir string) error {
	mimetypeFilePath := filepath.Join(tempDir, mimetypeFilename)

	if err := ioutil.WriteFile(mimetypeFilePath, []byte(mediaTypeEpub), filePermissions); err != nil {
		return err
	}

	return nil
}

func (e *Epub) writePackageFile(tempDir string) error {
	err := e.pkg.write(tempDir)
	if err != nil {
		return err
	}

	return nil
}

// Write the section files to the temporary directory and add the sections to
// the TOC and package files
func (e *Epub) writeSections(tempDir string) error {
	for i, section := range e.sections {
		sectionIndex := i + 1
		sectionFilename := fmt.Sprintf(sectionFileFormat, sectionIndex)
		sectionFilePath := filepath.Join(tempDir, contentFolderName, xhtmlFolderName, sectionFilename)
		section.write(sectionFilePath)

		relativePath := filepath.Join(xhtmlFolderName, sectionFilename)
		e.toc.addSection(sectionIndex, section.Title(), relativePath)
		e.pkg.addToManifest(sectionFilename, relativePath, mediaTypeXhtml, "")
		e.pkg.addToSpine(sectionFilename)
	}

	return nil
}

// Write the TOC file to the temporary directory and add the TOC entries to the
// package file
func (e *Epub) writeToc(tempDir string) error {
	e.pkg.addToManifest(tocNavItemID, tocNavFilename, mediaTypeXhtml, tocNavItemProperties)
	e.pkg.addToManifest(tocNcxItemID, tocNcxFilename, mediaTypeNcx, "")

	err := e.toc.write(tempDir)
	if err != nil {
		return err
	}

	return nil
}
