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
  <body></body>
</html>
`
)

type body struct {
	Data string `xml:",innerxml"`
}

type xhtml struct {
	XMLName xml.Name `xml:"http://www.w3.org/1999/xhtml html"`
	Title   string   `xml:"head>title"`
	Body    body     `xml:"body"`
}

func (x *xhtml) setBody(body string) {
	x.Body.Data = body
}
