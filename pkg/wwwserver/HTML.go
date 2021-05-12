package wwwserver

import (
  "bytes"
  "errors"
  "io"
	"io/ioutil"
	"net/http"
  "strconv"
  "strings"
  "time"
  
  "github.com/wtsuite/wtsuite/pkg/files"
)

type HTML struct {
  tryPath        string // for DefaultNotFound, in case an actual NotFound html is found
  cookiesAsAttrs bool
  FileData
}

func NewHTML(path string, cookiesAsAttrs bool) (*HTML, error) {
  if !files.IsFile(path) {
    return nil, errors.New("\""+path+"\" not found")
  }

  return &HTML{"", cookiesAsAttrs, newFileData(path, "text/html")}, nil
}

func GenErrorPage(title string, status int) []byte {
  var b strings.Builder

  b.WriteString(`<!DOCTYPE html><html lang="en" status="`)
  b.WriteString(strconv.Itoa(status))
  b.WriteString(`><head><meta charset="utf-8"><title>`)
  b.WriteString(title)
  b.WriteString(`</title></head><body><h1>`)
  b.WriteString(title)
  b.WriteString(`</h1></body></html>`)

  return []byte(b.String())
}

// page not found
func Default404(tryPath string) *HTML {
  h := &HTML{tryPath, false, newFileData("", "text/html")}

  h.FileData.cache(GenErrorPage("Not found", 404))

  return h
}

func Default503() *HTML {
  h := &HTML{"", false, newFileData("", "text/html")}

  h.FileData.cache(GenErrorPage("Service unavailable", 503))

  return h
}

func (h *HTML) cache() error {
  if h.path != "" && (h.buf == nil || (!h.frozen && !h.FileData.isUpToDate())) {
		b, err := ioutil.ReadFile(h.path)
		if err != nil {
			return errors.New("unable to read file \""+h.path+"\" at serve time")
		}

    h.FileData.cache(b)
    h.FileData.grabLatestModTime()
  } else if !h.frozen && h.path == "" && h.tryPath != "" && files.IsFile(h.tryPath) {
    h.path = h.tryPath
    return h.cache()
  }

  return nil
}

func (h *HTML) Serve(resp *ResponseWriter, req *http.Request) error {
  return h.ServeStatus(resp, req, http.StatusOK)
}

func (h *HTML) insertCookiesAsAttrs(req *http.Request) ([]byte, error) {
  bBuf := &bytes.Buffer{}

  h.buf.Seek(0, io.SeekStart)
  io.Copy(bBuf, h.buf)

  b := bBuf.Bytes()

  iCut := -1
  for i := 4; i < len(b); i++ {
    c0 := b[i-4]
    c1 := b[i-3]
    c2 := b[i-2]
    c3 := b[i-1]
    c4 := b[i]

    if c0 == '<' && c1 == 'h' && c2 == 't' && c3 == 'm' && c4 == 'l' {
      iCut = i+1
    }
  }

  if iCut == -1 {
    return nil, errors.New("unable to insert cookies")
  }

  bBef := b[0:iCut]
  bAft := b[iCut:]

  resBuf := &bytes.Buffer{}
  resBuf.Write(bBef)

  for _, cookie := range req.Cookies() {
    resBuf.WriteString(" ")
    resBuf.WriteString(cookie.Name)
    resBuf.WriteString("=\"")
    resBuf.WriteString(cookie.Value)
    resBuf.WriteString("\"")
  }

  resBuf.Write(bAft)

  return resBuf.Bytes(), nil
}

func (h *HTML) ServeStatus(resp *ResponseWriter, req *http.Request, status int) error {
	if err := h.cache(); err != nil {
		return err
	}

  if h.cookiesAsAttrs {
    b, err := h.insertCookiesAsAttrs(req)
    if err != nil {
      return err
    }

    fTmp_ := newFileData("", "text/html")
    fTmp := &fTmp_
    fTmp.cache(b)

    h.accessTime = time.Now().In(time.UTC)

    if err := fTmp.ServeStatus(resp, req, status); err != nil {
      return err
    }

    return nil
  } else {
    return h.FileData.ServeStatus(resp, req, status)
  }
}

func (h *HTML) Freeze() error {
  err := h.cache()
  h.frozen = true
  return err
}
