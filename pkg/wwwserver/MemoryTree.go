package wwwserver

import (
  "time"
)

func NewMemoryTree() *Tree {
  notFound := Default404("")

  return &Tree{"", time.Time{}, DefaultIndexNames, nil, make(map[string]Resource), notFound, false}
}

func (t *Tree) RegisterRaw(path string, b []byte, mimeType string) bool {
  fd := &FileData{"", time.Time{}, time.Time{}, mimeType, "", nil, nil, true}

  fd.cache(b)

  var res Resource = nil
  if mimeType == "text/html" {
    res = &HTML{"", false, *fd}
  } else {
    res = &File{*fd}
  }

  t.files[path] = res

  return true
}
