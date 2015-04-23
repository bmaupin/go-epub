package epub

type toc struct {
    navBodyTemplate = `    <nav epub:type="toc">
      <h1>Table of Contents</h1>
      <ol>
        <li><a href="xhtml/section0001.xhtml">Section 1</a></li>
      </ol>
    </nav>
`
    xmlnsEpub = `xmlns:epub="http://www.idpf.org/2007/ops"`
}
