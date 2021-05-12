package wwwserver

import (
  "errors"
  "net/http"
  "path/filepath"
  "os"
  "strings"
  "time"

  "github.com/wtsuite/wtsuite/pkg/files"
)

var DefaultIndexNames []string = []string{"index.html"}

var DefaultMimeTypes map[string]string = map[string]string {
  ".bin":   "application/octet-stream",
  ".css":   "text/css",
  ".gif":   "image/gif",
  ".html":  "text/html",
  ".js":    "text/javascript",
  ".json":  "application/json",
  ".png":   "image/png",
  ".svg":   "image/svg+xml",
  ".txt":   "text/plain",
  ".wasm":  "application/wasm",
  ".woff2": "font/woff2",
}

type Tree struct {
  root     string
  accessTime time.Time
  indexNames []string
  mimeTypes map[string]string
	files    map[string]Resource // url to resource map
	notFound *HTML
  cookiesAsAttrs bool
}

func NewTree(root string, indexNames []string, mimeTypes map[string]string,
	notFoundPath string, cookiesAsAttrs bool) (*Tree, error) {
	files := make(map[string]Resource)

	reuseMap := make(map[string]Resource) // path -> resource, so we can reuse resource for different urls

  if indexNames == nil {
    indexNames = DefaultIndexNames
  }

  if mimeTypes == nil {
    mimeTypes = DefaultMimeTypes
  }

  if notFoundPath == "" {
    notFoundPath = filepath.Join(root, "404.html")
  }

	rootLen := len(root)

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New("unable to walk file tree at root \""+root+"\"")
		}

		var url string
		if path == root {
			url = "/"
		} else {
			// abbreviate the path by cutting off the root
			url = path[rootLen:]
		}

		resource, err := pathToResource(path, info.IsDir(), indexNames, mimeTypes, reuseMap, cookiesAsAttrs)
		if err != nil {
			return err
		}

		if resource != nil {
			files[url] = resource

			if info.IsDir() {
				if strings.HasSuffix(url, "/") {
					files[strings.TrimSuffix(url, "/")] = resource
				} else {
					files[url+"/"] = resource
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	notFoundResource, err := pathToResource(notFoundPath, false, indexNames, mimeTypes, reuseMap, cookiesAsAttrs)
	if err != nil {
		notFoundResource = nil
	}

	if notFoundResource == nil {
    // create a default
    notFoundResource = Default404(notFoundPath)
	}

	notFoundFile, ok := notFoundResource.(*HTML)
	if !ok {
		return nil, errors.New("404 file \""+notFoundPath+"\" isn't an html resource")
	}

  // remember root/indexNames/mimeTypes if files are added later
	return &Tree{
    root, 
    time.Time{},
    indexNames,
    mimeTypes,
    files, 
    notFoundFile,
    cookiesAsAttrs,
  }, nil
}

// returns nil if no (valid) resource found at path
func pathToResource(path string, isDir bool, indexNames []string, mimeTypes map[string]string, reuseMap map[string]Resource, cookiesAsAttrs bool) (Resource, error) {
	resourcePath := path
	if isDir {
		indexFound := false
		// look for index, first one found wins
		for _, testIndex := range indexNames {
			testPath := filepath.Join(path, testIndex)
			if files.IsFile(testPath) {
				resourcePath = testPath
				indexFound = true
				break
			} 
		}

		if !indexFound {
			return nil, nil
		}
	}

	if reuseMap != nil {
    if prev, ok := reuseMap[resourcePath]; ok {
      // assume it already passed the mimetype test
      return prev, nil
    }
	} 

  // test mime types
  ext := filepath.Ext(resourcePath)

  if mimeType, ok := mimeTypes[ext]; ok {
    var resource Resource

    if mimeType == "text/html" {
      var err error
      resource, err = NewHTML(resourcePath, cookiesAsAttrs)
      if err != nil {
        return nil, err
      }
    } else {
      var err error
      resource, err = NewFile(resourcePath, mimeType)
      if err != nil {
        return nil, err
      }
    }

    if reuseMap != nil {
      reuseMap[resourcePath] = resource
    }

    return resource, nil
  }

  return nil, nil
}

// all errors result in false
func (t *Tree) updateIfFile(url string) (Resource, bool) {
  if t.root == "" {
    // lives only in memory
    return nil, false
  }

  path := filepath.Join(t.root, url)

  info, err := os.Stat(path)
  if err != nil {
    return nil, false
  }

  resource, err := pathToResource(path, info.IsDir(), t.indexNames, t.mimeTypes, nil, t.cookiesAsAttrs)
  if resource == nil || err != nil {
    return nil, false
  }

  t.files[url] = resource

  if info.IsDir() {
    if strings.HasSuffix(url, "/") {
      t.files[strings.TrimSuffix(url, "/")] = resource
    } else {
      t.files[url+"/"] = resource
    }
  }

  return resource, true
}

func (t *Tree) indexFile(path string) (Resource, bool) {
  for _, testIndex := range t.indexNames {
    testPath := filepath.Join(path, testIndex)
    if resource, ok := t.files[testPath]; ok {
      return resource, true
    }
  }

  return nil, false
}

func (t *Tree) Serve(resp *ResponseWriter, req *http.Request) error {
  t.accessTime = time.Now().In(time.UTC)

	if resource, ok := t.files[req.URL.Path]; ok {
		return resource.Serve(resp, req)
  } else if resource, ok := t.indexFile(req.URL.Path); ok {
    return resource.Serve(resp, req)
  } else {
    if resource, ok := t.updateIfFile(req.URL.Path); ok {
      return resource.Serve(resp, req)
    } else {
      return t.notFound.ServeStatus(resp, req, http.StatusNotFound)
    }
	}
}

func (t *Tree) ServeFrozen(resp *ResponseWriter, req *http.Request) error {
  t.accessTime = time.Now().In(time.UTC)

	if resource, ok := t.files[req.URL.Path]; ok {
		return resource.ServeFrozen(resp, req)
  } else if resource, ok := t.indexFile(req.URL.Path); ok {
		return resource.ServeFrozen(resp, req)
  } else {
    return t.notFound.ServeStatus(resp, req, http.StatusNotFound)
	}
}

func (t *Tree) Freeze() error {
  for _, resource := range t.files {
    if err := resource.Freeze(); err != nil {
      return err
    }
  }

  return nil
}

func (t *Tree) ListFiles() []string {
  paths := make([]string, 0)

  for k, _ := range t.files {
    paths = append(paths, k)
  }

  return paths
}

func (t *Tree) AccessTime() time.Time {
  return t.accessTime
}
