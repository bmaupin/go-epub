package epub

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// ErrUnableToCreateEpub is thrown by Write if it cannot create the destination
// EPUB file
var ErrUnableToCreateEpub = errors.New("Unable to create EPUB file")

// ErrRetrievingImage is thrown by Write if it cannot get the source image that
// was added using AddImage
var ErrRetrievingImage = errors.New("Error retrieving image from source")

const (
	cssFolderName         = "css"
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
	mediaTypeCss      = "text/css"
	mediaTypeEpub     = "application/epub+zip"
	mediaTypeGif      = "image/gif"
	mediaTypeJpeg     = "image/jpeg"
	mediaTypeNcx      = "application/x-dtbncx+xml"
	mediaTypePng      = "image/png"
	mediaTypeSvg      = "image/svg+xml"
	mediaTypeXhtml    = "application/xhtml+xml"
	metaInfFolderName = "META-INF"
	mimetypeFilename  = "mimetype"
	pkgFilename       = "package.opf"
	tempDirPrefix     = "go-epub"
	xhtmlFolderName   = "xhtml"
)

// Write writes the EPUB file. The destination path must be the full path to
// the resulting file, including filename and extension.
func (e *Epub) Write(destFilePath string) error {
	tempDir, err := ioutil.TempDir("", tempDirPrefix)
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			panic(fmt.Sprintf("Error removing temp directory: %s", err))
		}
	}()
	if err != nil {
		panic(fmt.Sprintf("Error creating temp directory: %s", err))
	}

	writeMimetype(tempDir)
	createEpubFolders(tempDir)

	// Must be called after:
	// createEpubFolders()
	writeContainerFile(tempDir)

	// Must be called after:
	// createEpubFolders()
	e.writeCSSFiles(tempDir)

	// Must be called after:
	// createEpubFolders()
	err = e.writeImages(tempDir)
	if err != nil {
		return err
	}

	// Must be called after:
	// createEpubFolders()
	e.writeSections(tempDir)

	// Must be called after:
	// createEpubFolders()
	// writeSections()
	e.writeToc(tempDir)

	// Must be called after:
	// createEpubFolders()
	// writeCSSFiles()
	// writeImages()
	// writeSections()
	// writeToc()
	e.writePackageFile(tempDir)

	// Must be called last
	err = e.writeEpub(tempDir, destFilePath)
	if err != nil {
		return err
	}

	return nil
}

// Create the EPUB folder structure in a temp directory
func createEpubFolders(tempDir string) {
	if err := os.Mkdir(
		filepath.Join(
			tempDir,
			contentFolderName,
		),
		dirPermissions); err != nil {
		// No reason this should happen if tempDir creation was successful
		panic(fmt.Sprintf("Error creating EPUB subdirectory: %s", err))
	}

	if err := os.Mkdir(
		filepath.Join(
			tempDir,
			contentFolderName,
			xhtmlFolderName,
		),
		dirPermissions); err != nil {
		panic(fmt.Sprintf("Error creating xhtml subdirectory: %s", err))
	}

	if err := os.Mkdir(
		filepath.Join(
			tempDir,
			metaInfFolderName,
		),
		dirPermissions); err != nil {
		panic(fmt.Sprintf("Error creating META-INF subdirectory: %s", err))
	}
}

// Write the contatiner file (container.xml), which mostly just points to the
// package file (package.opf)
//
// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v32/META-INF/container.xml
// Spec: http://www.idpf.org/epub/301/spec/epub-ocf.html#sec-container-metainf-container.xml
func writeContainerFile(tempDir string) {
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
		panic(fmt.Sprintf("Error writing container file: %s", err))
	}
}

// Write the CSS files to the temporary directory and add them to the package
// file
func (e *Epub) writeCSSFiles(tempDir string) {
	if len(e.css) > 0 {
		cssFolderPath := filepath.Join(tempDir, contentFolderName, cssFolderName)
		if err := os.Mkdir(cssFolderPath, dirPermissions); err != nil {
			panic(fmt.Sprintf("Unable to create css directory: %s", err))
		}

		for cssFilename, cssFileContent := range e.css {
			cssFilePath := filepath.Join(cssFolderPath, cssFilename)

			// Add the CSS file to the EPUB temp directory
			if err := ioutil.WriteFile(cssFilePath, []byte(cssFileContent), filePermissions); err != nil {
				panic(fmt.Sprintf("Error writing CSS file: %s", err))
			}

			// Add the CSS filename to the package file manifest
			relativePath := filepath.Join(cssFolderName, cssFilename)
			e.pkg.addToManifest(cssFilename, relativePath, mediaTypeCss, "")
		}
	}
}

// Write the EPUB file itself by zipping up everything from a temp directory
func (e *Epub) writeEpub(tempDir string, destFilePath string) error {
	f, err := os.Create(destFilePath)
	if err != nil {
		return ErrUnableToCreateEpub
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	z := zip.NewWriter(f)
	defer func() {
		if err := z.Close(); err != nil {
			panic(err)
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
			// tempDir and path are both internal, so we shouldn't get here
			panic(fmt.Sprintf("Error closing EPUB file: %s", err))
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
			panic(fmt.Sprintf("Error creating zip writer: %s", err))
		}

		r, err := os.Open(path)
		if err != nil {
			panic(fmt.Sprintf("Error opening file being added to EPUB: %s", err))
		}
		defer func() {
			if err := r.Close(); err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(w, r)
		if err != nil {
			panic(fmt.Sprintf("Error copying contents of file being added EPUB: %s", err))
		}

		return nil
	}

	// Add the mimetype file first
	mimetypeFilePath := filepath.Join(tempDir, mimetypeFilename)
	mimetypeInfo, err := os.Lstat(mimetypeFilePath)
	if err != nil {
		panic(fmt.Sprintf("Unable to get FileInfo for mimetype file: %s", err))
	}
	err = addFileToZip(mimetypeFilePath, mimetypeInfo, nil)
	if err != nil {
		panic(fmt.Sprintf("Unable to add mimetype file to EPUB: %s", err))
	}

	skipMimetypeFile = true

	err = filepath.Walk(tempDir, addFileToZip)
	if err != nil {
		panic(fmt.Sprintf("Unable to add file to EPUB: %s", err))
	}

	return nil
}

// Get images from their source and save them in the temporary directory
func (e *Epub) writeImages(tempDir string) error {
	if len(e.images) > 0 {
		imageFolderPath := filepath.Join(tempDir, contentFolderName, imageFolderName)
		if err := os.Mkdir(imageFolderPath, dirPermissions); err != nil {
			panic(fmt.Sprintf("Unable to create img directory: %s", err))
		}

		for imageFilename, imageSource := range e.images {
			// Get the image from the source
			u, err := url.Parse(imageSource)
			if err != nil {
				return ErrRetrievingImage
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
				return ErrRetrievingImage
			}
			defer func() {
				if err := r.Close(); err != nil {
					panic(err)
				}
			}()

			imageFilePath := filepath.Join(
				imageFolderPath,
				imageFilename,
			)

			// Add the image to the EPUB temp directory
			w, err := os.Create(imageFilePath)
			if err != nil {
				panic(fmt.Sprintf("Unable to create image file: %s", err))
			}
			defer func() {
				if err := w.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(w, r)
			if err != nil {
				// There shouldn't be any problem with the writer, but the reader
				// might have an issue
				return ErrRetrievingImage
			}

			// Determine the media type
			imageMediaType := ""
			if filepath.Ext(imageFilename) == ".gif" {
				imageMediaType = mediaTypeGif
			} else if filepath.Ext(imageFilename) == ".jpg" || filepath.Ext(imageFilename) == ".jpeg" {
				imageMediaType = mediaTypeJpeg
			} else if filepath.Ext(imageFilename) == ".png" {
				imageMediaType = mediaTypePng
			} else if filepath.Ext(imageFilename) == ".svg" {
				imageMediaType = mediaTypeSvg
			}

			// Add the image to the package file manifest
			e.pkg.addToManifest(imageFilename, filepath.Join(imageFolderName, imageFilename), imageMediaType, "")
		}
	}

	return nil
}

// Write the mimetype file
//
// Sample: https://github.com/bmaupin/epub-samples/blob/master/minimal-v32/mimetype
// Spec: http://www.idpf.org/epub/301/spec/epub-ocf.html#sec-zip-container-mime
func writeMimetype(tempDir string) {
	mimetypeFilePath := filepath.Join(tempDir, mimetypeFilename)

	if err := ioutil.WriteFile(mimetypeFilePath, []byte(mediaTypeEpub), filePermissions); err != nil {
		panic(fmt.Sprintf("Error writing mimetype file: %s", err))
	}
}

func (e *Epub) writePackageFile(tempDir string) {
	e.pkg.write(tempDir)
}

// Write the section files to the temporary directory and add the sections to
// the TOC and package files
func (e *Epub) writeSections(tempDir string) {
	if len(e.sections) > 0 {
		sectionIndex := 0
		for sectionFilename, section := range e.sections {
			sectionIndex++
			sectionFilePath := filepath.Join(tempDir, contentFolderName, xhtmlFolderName, sectionFilename)
			section.write(sectionFilePath)

			relativePath := filepath.Join(xhtmlFolderName, sectionFilename)
			e.toc.addSection(sectionIndex, section.Title(), relativePath)
			e.pkg.addToManifest(sectionFilename, relativePath, mediaTypeXhtml, "")
			e.pkg.addToSpine(sectionFilename)
		}
	}
}

// Write the TOC file to the temporary directory and add the TOC entries to the
// package file
func (e *Epub) writeToc(tempDir string) {
	e.pkg.addToManifest(tocNavItemID, tocNavFilename, mediaTypeXhtml, tocNavItemProperties)
	e.pkg.addToManifest(tocNcxItemID, tocNcxFilename, mediaTypeNcx, "")

	e.toc.write(tempDir)
}
