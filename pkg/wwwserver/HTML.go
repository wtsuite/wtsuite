package wwwserver

import (
  "errors"
	"io/ioutil"
	"net/http"
  "strconv"
  "strings"
  
  "github.com/wtsuite/wtsuite/pkg/files"
)

type HTML struct {
  tryPath string // for DefaultNotFound, in case an actual NotFound html is found
  FileData
}

func NewHTML(path string) (*HTML, error) {
  if !files.IsFile(path) {
    return nil, errors.New("\""+path+"\" not found")
  }

  return &HTML{"", newFileData(path, "text/html")}, nil
}

func genErrorPage(title string, status int) []byte {
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
  h := &HTML{tryPath, newFileData("", "text/html")}

  h.FileData.cache(genErrorPage("Not found", 404))

  return h
}

func Default503() *HTML {
  h := &HTML{"", newFileData("", "text/html")}

  h.FileData.cache(genErrorPage("Service unavailable", 503))

  return h
}

func (h *HTML) cache() error {
  if h.path != "" && (h.buf == nil || !h.FileData.isUpToDate()) {
		b, err := ioutil.ReadFile(h.path)
		if err != nil {
			return errors.New("unable to read file \""+h.path+"\" at serve time")
		}

    h.FileData.cache(b)
    h.FileData.grabLatestModTime()
  } else if h.path == "" && h.tryPath != "" && files.IsFile(h.tryPath) {
    h.path = h.tryPath
    return h.cache()
  }

  return nil
}

func (h *HTML) Serve(resp *ResponseWriter, req *http.Request) error {
  return h.ServeStatus(resp, req, http.StatusOK)
}

func (h *HTML) ServeStatus(resp *ResponseWriter, req *http.Request, status int) error {
	if err := h.cache(); err != nil {
		return err
	}

  return h.FileData.ServeStatus(resp, req, status)
}

func (h *HTML) Freeze() error {
  return h.cache()
}
