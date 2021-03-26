package main

import (
  "log"
  "net/http"
  "os"
  "path/filepath"
  "strings"

  "github.com/computeportal/wtsuite/pkg/files"
  "github.com/computeportal/wtsuite/pkg/wwwserver"
)

type Router struct {
  logger *log.Logger
  content *wwwserver.Tree
}

func NewRouter(root string) (*Router, error) {
  notFoundPath := filepath.Join(root, "404.html")
  if !files.IsFile(notFoundPath) {
    notFoundPath = ""
  }

  content, err := wwwserver.NewTree(
    root, 
    wwwserver.DefaultIndexNames,
    wwwserver.DefaultMimeTypes,
    notFoundPath,
  )

  if err != nil {
    return nil, err
  }

  return &Router{log.New(os.Stdout, "", log.Ltime), content}, nil
}

func (r *Router) ServeHTTP(resp_ http.ResponseWriter, req *http.Request) {
	// wrap resp_ so we have accecss to the returned status, size, etc
	resp := wwwserver.NewResponseWriter(resp_)

	if req.Method != "GET" {
		resp.WriteError("Error: not a GET request")
  } else {
    if err := r.content.Serve(resp, req); err != nil {
      r.LogError(err)
    }
  }

	r.LogAccess(resp, req)
}

func (r *Router) LogError(err error) {
  r.logger.Printf("Error: %s\n", err.Error())
}

func (r *Router) LogAccess(resp *wwwserver.ResponseWriter, req *http.Request) {
  r.logger.Printf("%s: %s\t(from:%s,\t%d)\n", req.Method, req.URL.Path, strings.Split(req.Referer(), "?")[0], resp.Status())
}
