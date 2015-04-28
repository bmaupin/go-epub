package epub

import (
	"encoding/xml"
)

const (
    xhtmlTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title></title>
  </head>
  <body>
  </body>
</html>
`
)

type Body struct {
	Data string `xml:",chardata"`
}

type Head struct {
	Title string
}

type Xhtml struct {
    XMLName xml.Name `xml:"http://www.w3.org/1999/xhtml html"`
    Head Head `xml:"head"`
    Body Body `xml:"body"`
}
