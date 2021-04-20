package directives

import (
  "sort"
)

type FileScope struct {
	permissive bool
  pagesWithRelURLs []string
  cache *FileCache
  ScopeData
}

func NewFileScope(permissive bool, cache *FileCache) *FileScope {
  return &FileScope{permissive, make([]string, 0), cache, newScopeData(nil)}
}

func (s *FileScope) Permissive() bool {
  return s.permissive
}

func (s *FileScope) NotifyRelativeURL(path string) {
  for _, p := range s.pagesWithRelURLs {
    if p == path {
      return
    }
  }

  s.pagesWithRelURLs = append(s.pagesWithRelURLs, path)
}

func (s *FileScope) GetCache() *FileCache {
  return s.cache
}

func (s *FileScope) PagesWithRelURLs() []string {
  sort.Strings(s.pagesWithRelURLs)

  return s.pagesWithRelURLs
}
