package epub

import (
    "archive/zip"
//    "bytes"
    "encoding/xml"
	"fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    
//    "github.com/satori/go.uuid"
)

const (
    containerFilename = "container.xml"
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
    filePermissions = 0644
    metaInfFolderName = "META-INF"
    mimetypeContent = "application/epub+zip"
    mimetypeFilename = "mimetype"
    pkgdocFilename = "package.opf"
    tempDirPrefix = "go-epub"
    tocFilename = "toc.ncx"
)

func (e *Epub) Write() {
    // Files to include in the built epub
//    filesToInclude := []string{}

    tempDir, err := ioutil.TempDir("", tempDirPrefix)
    defer os.Remove(tempDir)
    if err != nil {
        log.Fatalf("os.Remove error: %s", err)
    }

    err = writeMimetype(tempDir)
    if err != nil {
        log.Fatalf("writeMimetype error: %s", err)
    }
    err = writeContainerFile(tempDir)
    if err != nil {
        log.Fatalf("writeContainerFile error: %s", err)
    }
    
    if e.filename == "" {
    	log.Println("filename nil")
//    	efilename = ""
    }
    
    err = e.writeEpub(tempDir, e.filename)
    if err != nil {
        log.Fatalf("writeEpub error: %s", err)
    }
}

func writeContainerFile(tempDir string) error {
    metaInfFolderPath := filepath.Join(tempDir, metaInfFolderName)
    if err := os.Mkdir(metaInfFolderPath, dirPermissions); err != nil {
        return err
    }
    
    containerFilePath := filepath.Join(metaInfFolderPath, containerFilename)
    if err := ioutil.WriteFile(
    		containerFilePath,
    		[]byte(
    			fmt.Sprintf(
					containerFileTemplate, 
					contentFolderName,
					pkgdocFilename,
	    		),
	    	),
    		filePermissions,
			); err != nil {
        return err
    }
    
    return nil
}

/*   
    type Rootfile struct {
        FullPath string `xml:"full-path,attr"`
        MediaType string `xml:"media-type,attr"`
    }

    type Rootfiles struct {
        Rootfile []Rootfile `xml:"rootfile"`
    }

    type Container struct {
        XMLName xml.Name `xml:"urn:oasis:names:tc:opendocument:xmlns:container container"`
        Version string `xml:"version,attr"`
        Rootfiles Rootfiles `xml:"rootfiles"`
    }

    v := &Container{Version: containerVersion}
    
    v.Rootfiles.Rootfile = append(
        v.Rootfiles.Rootfile, 
        Rootfile{
            FullPath: filepath.Join(oebpsFolderName, contentFilename), 
            MediaType: mediaTypeOebpsPackage,
        },
    )

    output, err := xml.MarshalIndent(v, "", `   `)
    if err != nil {
        return err
    }
    // Add the xml header to the output
    containerContent := append([]byte(xml.Header), output...)
    // It's generally nice to have files end with a newline
    containerContent = append(containerContent, "\n"...)
    
    containerFilePath := filepath.Join(metaInfFolderPath, containerFilename)
    if err := ioutil.WriteFile(containerFilePath, containerContent, filePermissions); err != nil {
        return err
    }
    
    return nil
}

/*
func writeContentFile(tempDir string) error {    
    oebpsFolderPath := filepath.Join(tempDir, oebpsFolderName)
    if err := os.Mkdir(oebpsFolderPath, dirPermissions); err != nil {
        return err
    }
    
    type Package struct {
        XMLName xml.Name `xml:"http://www.idpf.org/2007/opf package"`
        UniqueIdentifier string `xml:"unique-identifier,attr"`
        Version string `xml:"version,attr"`
    }
    
    
    
    
    
    type Rootfile struct {
        FullPath string `xml:"full-path,attr"`
        MediaType string `xml:"media-type,attr"`
    }

    type Rootfiles struct {
        Rootfile []Rootfile `xml:"rootfile"`
    }

    type Container struct {
        XMLName xml.Name `xml:"urn:oasis:names:tc:opendocument:xmlns:container container"`
        Version string `xml:"version,attr"`
        Rootfiles Rootfiles `xml:"rootfiles"`
    }

    v := &Container{Version: containerVersion}
    
    v.Rootfiles.Rootfile = append(
        v.Rootfiles.Rootfile, 
        Rootfile{
            FullPath: filepath.Join(oebpsFolderName, contentFilename), 
            MediaType: mediaTypeOebpsPackage,
        },
    )

    output, err := xml.MarshalIndent(v, "", `   `)
    if err != nil {
        return err
    }
    // Add the xml header to the output
    containerContent := append([]byte(xml.Header), output...)
    // It's generally nice to have files end with a newline
    containerContent = append(containerContent, "\n"...)
    
    containerFilePath := filepath.Join(metaInfFolderPath, containerFilename)
    if err := ioutil.WriteFile(containerFilePath, containerContent, filePermissions); err != nil {
        return err
    }
    
    return nil
}
*/
func writeMimetype(tempDir string) error {
    mimetypeFilePath := filepath.Join(tempDir, mimetypeFilename)
    
    if err := ioutil.WriteFile(mimetypeFilePath, []byte(mimetypeContent), filePermissions); err != nil {
        return err
    }
    
    return nil
}

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

		if skipMimetypeFile == true {
			// Skip the mimetype file 
			if path == filepath.Join(tempDir, mimetypeFilename) {
				return nil
			}
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

		r, err := os.Open(path)
		if err != nil {
			log.Fatalf("os.Open error: %s", err)
		}
		defer func() {
			if err := r.Close(); err != nil {
				log.Fatalf("os.File.Close error: %s", err)
			}
		}()

		w, err := z.Create(relativePath)
		if err != nil {
			log.Fatalf("zip.Writer.Create error: %s", err)
		}

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

func (e *Epub) writePkgdocFile(tempDir string) error {
    pkgdocFilePath := filepath.Join(tempDir, pkgdocFilename)
    
    output, _ := xml.MarshalIndent(e.pkgdoc, "", `   `)
    // Add the xml header to the output
    pkgdocFileContent := append([]byte(xml.Header), output...)
    // It's generally nice to have files end with a newline
    pkgdocFileContent = append(pkgdocFileContent, "\n"...)
    
    if err := ioutil.WriteFile(pkgdocFilePath, []byte(pkgdocFileContent), filePermissions); err != nil {
        return err
    }
    
    return nil
}
