package main

import (
  "os"
  "path/filepath"
  "regexp"
  "sort"
  "strconv"
  "strings"

	"github.com/wtsuite/wtsuite/pkg/directives"
	"github.com/wtsuite/wtsuite/pkg/files"
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
	"github.com/wtsuite/wtsuite/pkg/tokens/raw"
	"github.com/wtsuite/wtsuite/pkg/styles"
)

var BAD_DST *regexp.Regexp = regexp.MustCompile(`^[\.]+\/`)

type PageConfig struct {
  url    string
  dst    string
  src    string
  params []string
}

type ScriptConfig struct {
  src   string
  hash  string
  pages []string
}

type StyleConfig struct {
  src   string
  dst   string
  url   string
  pages []string
  sheet styles.Sheet
}

type FileConfig struct {
  url string
  dst string
  src string
}


type SiteConfig struct {
  outputDir string
  Pages   []PageConfig
  Scripts []ScriptConfig 
  Styles  []StyleConfig 
  Files   []FileConfig
  //Search  []search.SearchIndexConfig
}

func parseDstURL(outputDir string, url *tokens.String) (string, string, error) {
  if BAD_DST.MatchString(url.Value()) {
    errCtx := url.Context()
    return "", "", errCtx.NewError("Error: bad dst url")
  }

  res := url.Value()
  if strings.HasPrefix(res, "/") {
    res = strings.TrimPrefix(res, "/")
  }

  return filepath.Join(outputDir, res), "/" + res, nil
}

func stringList(t_ tokens.Token) ([]*tokens.String, error) {
  switch {
  case tokens.IsString(t_):
    t, err := tokens.AssertString(t_)
    if err != nil {
      return nil, err
    }

    return []*tokens.String{t}, nil
  case tokens.IsList(t_):
    t, err := tokens.AssertList(t_)
    if err != nil {
      return nil, err
    }

    lst := make([]*tokens.String, 0)
    if err := t.Loop(func(i int, val_ tokens.Token, _ bool) error {
      val, err := tokens.AssertString(val_)
      if err != nil {
        return err
      }

      lst = append(lst, val)

      return nil
    }); err != nil {
      return nil, err
    }

    return lst, nil
  default:
    errCtx := t_.Context()
    return nil, errCtx.NewError("Error: expected string or list of strings")
  }
}

func ReadConfigFile(fname string, outputDir string) (*SiteConfig, error) {
  if err := files.ResolvePackages(fname); err != nil {
    return nil, err
  }

  t_, err := directives.BuildJSON(fname, context.NewDummyContext())
  if err != nil {
    return nil, err
  }

  t, err := tokens.AssertStringDict(t_)
  if err != nil {
    return nil, err
  }

  cfg := &SiteConfig {
    outputDir: outputDir,
    Pages: nil,
    Scripts: make([]ScriptConfig, 0),
    Styles: make([]StyleConfig, 0),
    Files: make([]FileConfig, 0),
    //Search: make([]search.SearchIndexConfig),
  }

  pages_, ok := t.Get("pages")
  if !ok {
    errCtx := t.Context()
    return nil, errCtx.NewError("pages not found in dict")
  }

  pages, err := tokens.AssertStringDict(pages_)
  if err != nil {
    return nil, err
  }

  cfg.Pages, err = readPages(fname, outputDir, pages)
  if err != nil {
    return nil, err
  }

  if files_, ok := t.Get("files"); ok {
    files__, err := tokens.AssertStringDict(files_)
    if err != nil {
      return nil, err
    }

    cfg.Files, err = readFiles(fname, outputDir, files__, cfg.Pages)
    if err != nil {
      return nil, err
    }
  }

  if scripts_, ok := t.Get("scripts"); ok {
    scripts, err := tokens.AssertStringDict(scripts_)
    if err != nil {
      return nil, err
    }

    cfg.Scripts, err = readScripts(fname, outputDir, scripts, cfg.Pages)
    if err != nil {
      return nil, err
    }
  }

  if styles_, ok := t.Get("styles"); ok {
    styles, err := tokens.AssertStringDict(styles_)
    if err != nil {
      return nil, err
    }

    cfg.Styles, err = readStyles(fname, outputDir, styles, cfg.Pages, cfg.Files)
    if err != nil {
      return nil, err
    }
  }

  if err := t.Loop(func(key *tokens.String, _ tokens.Token, last bool) error {
    switch key.Value(){
    case "pages", "files", "search", "styles", "scripts":
      return nil
    default:
      errCtx := key.Context()
      return errCtx.NewError("Error: unrecognized entry")
    }
  }); err != nil {
    return nil, err
  }

  // TODO: parse the search

  return cfg, nil
}

func readPages(configFile string, outputDir string, pages *tokens.StringDict) ([]PageConfig, error) {
  res := make([]PageConfig, 0)

  if err := pages.Loop(func(key *tokens.String, value_ tokens.Token, last bool) error {
    dst, url, err := parseDstURL(outputDir, key)
    if err != nil {
      return err
    }

    var src string
    params := make([]string, 0)

    args, err := stringList(value_)
    if err != nil {
      return err
    }

    if len(args) < 1 {
      errCtx := value_.Context()
      return errCtx.NewError("Error: expected at least 1 entry")
    }

    src, err = files.Search(configFile, args[0].Value())
    if err != nil {
      errCtx := args[0].Context()
      return errCtx.NewError(err.Error())
    }

    for _, param := range args[1:] {
      params = append(params, param.Value())
    }
    
    res = append(res, PageConfig{dst: dst, url: url, src: src, params: params})

    return nil
  }); err != nil {
    return nil, err
  }

  if len(res) == 0 {
    errCtx := pages.Context()
    return nil, errCtx.NewError("Error: should have at least one entry")
  }

  return res, nil
}

func readScripts(configFile string, outputDir string, scripts *tokens.StringDict, pages []PageConfig) ([]ScriptConfig, error) {
  res := make([]ScriptConfig, 0)

  if err := scripts.Loop(func(key *tokens.String, val_ tokens.Token, _ bool) error {
    src, err := files.Search(configFile, key.Value())
    if err != nil {
      errCtx := key.Context()
      return errCtx.NewError(err.Error())
    }

    lst, err := stringList(val_)
    if err != nil {
      return err
    }

    scriptPages := make([]string, 0)
    for _, p_ := range lst {
      _, pURL, err := parseDstURL(outputDir, p_)
      if err != nil {
        return err
      }

      found := false
      for _, pCheck := range pages {
        if pCheck.url == pURL {
          found = true

          scriptPages = append(scriptPages, pURL)

          break
        }
      }

      if !found {
        errCtx := p_.Context()
        return errCtx.NewError("Error: not a valid page")
      }
    }

    res = append(res, ScriptConfig{src: src, hash: raw.ShortHash(src), pages: scriptPages})

    return nil
  }); err != nil {
    return nil, err
  }

  return res, nil
}

func uniqueURL(base string, ext string, urlExists func(string) bool) string {
  if !urlExists(base + ext) {
    return base + ext
  }

  i := 0;
  for urlExists(base + strconv.Itoa(i) + ext) {
    i++
  }

  return base + strconv.Itoa(i) + ext
}

func uniqueStyleURL(src string, pages []PageConfig, files_ []FileConfig) string {
  urlExists := func(url string) bool {
    for _, p := range pages {
      if p.url == url {
        return true
      }
    }

    for _, f := range files_ {
      if f.url == url {
        return true
      }
    }

    return false
  }

  return uniqueURL("style" + raw.ShortHash(src), ".css", urlExists)
}

func uniqueScriptBundleURL(pages []PageConfig, files_ []FileConfig, styles []StyleConfig) string {
  urlExists := func(url string) bool {
    for _, p := range pages {
      if p.url == url {
        return true
      }
    }

    for _, f := range files_ {
      if f.url == url {
        return true
      }
    }

    for _, s := range styles {
      if s.url == url {
        return true
      }
    }

    return false
  }

  return uniqueURL("bundle", ".js", urlExists)
}

func uniqueMathFontURL(pages []PageConfig, files_ []FileConfig, styles []StyleConfig, jsBundleURL string) string {
  urlExists := func(url string) bool {
    for _, p := range pages {
      if p.url == url {
        return true
      }
    }

    for _, f := range files_ {
      if f.url == url {
        return true
      }
    }

    for _, s := range styles {
      if s.url == url {
        return true
      }
    }

    return jsBundleURL == url
  }

  return uniqueURL("math", ".woff2", urlExists)
}

func readStyles(configFile string, outputDir string, styles *tokens.StringDict, pages []PageConfig, files_ []FileConfig) ([]StyleConfig, error) {
  res := make([]StyleConfig, 0)

  if err := styles.Loop(func(key *tokens.String, val_ tokens.Token, _ bool) error {
    src, err := files.Search(configFile, key.Value())
    if err != nil {
      errCtx := key.Context()
      return errCtx.NewError(err.Error())
    }

    lst, err := stringList(val_)
    if err != nil {
      return err
    }

    stylePages := make([]string, 0)
    for _, p_ := range lst {
      _, pURL, err := parseDstURL(outputDir, p_)
      if err != nil {
        return err
      }

      found := false
      for _, pCheck := range pages {
        if pCheck.url == pURL {
          found = true

          stylePages = append(stylePages, pURL)

          break
        }
      }

      if !found {
        errCtx := p_.Context()
        return errCtx.NewError("Error: not a valid page")
      }
    }

    url := uniqueStyleURL(src, pages, files_)
    dst := filepath.Join(outputDir, url)

    res = append(res, StyleConfig{src: src, dst: dst, url: url, pages: stylePages, sheet: nil})

    return nil
  }); err != nil {
    return nil, err
  }

  return res, nil
}

func readFiles(configFile string, outputDir string, files_  *tokens.StringDict, pages []PageConfig) ([]FileConfig, error) {
  res := make([]FileConfig, 0)

  if err := files_.Loop(func(key *tokens.String, val_ tokens.Token, _ bool) error {
    val, err := tokens.AssertString(val_)
    if err != nil {
      return err
    }

    src, err := files.Search(configFile, val.Value())
    if err != nil {
      errCtx := val.Context()
      return errCtx.NewError(err.Error())
    }

    dst, url, err := parseDstURL(outputDir, key)
    if err != nil {
      return err
    }

    // there can't be any conflict with pages
    for _, pCheck := range pages {
      if pCheck.url == url {
        errCtx := key.Context()
        return errCtx.NewError("Error: already a page url")
      }
    }

    res = append(res, FileConfig{url: url, src: src, dst: dst})

    return nil
  }); err != nil {
    return nil, err
  }

  return res, nil
}

func (cfg *SiteConfig) JSURL() string {
  return uniqueScriptBundleURL(cfg.Pages, cfg.Files, cfg.Styles)
}

func (cfg *SiteConfig) JSDst() string {
  return filepath.Join(cfg.outputDir, cfg.JSURL())
}

func (cfg *SiteConfig) MathFontURL() string {
  return uniqueMathFontURL(cfg.Pages, cfg.Files, cfg.Styles, cfg.JSURL())
}

func (cfg *SiteConfig) MathFontDst() string {
  return filepath.Join(cfg.outputDir, cfg.MathFontURL())
}

func (cfg *SiteConfig) PageStyles(url string) []string {
  res := make([]string, 0)
  for _, s := range cfg.Styles {
    for _, p := range s.pages {
      if p == url {
        res = append(res, s.url)
      }
    }
  }

  sort.Strings(res)

  // only the unique ones
  res2 := make([]string, 0)
  for i, r := range res {
    if i == 0 || r != res[i-1] {
      res2 = append(res2, r)
    }
  }

  return res2
}

// returns hashes!
func (cfg *SiteConfig) PageScripts(url string) []string {
  res := make([]string, 0)
  for _, s := range cfg.Scripts {
    for _, p := range s.pages {
      if p == url {
        res = append(res, s.hash)
      }
    }
  }

  sort.Strings(res)

  // only the unique ones
  res2 := make([]string, 0)
  for i, r := range res {
    if i == 0 || r != res[i-1] {
      res2 = append(res2, r)
    }
  }

  return res2
}

func (cfg *SiteConfig) FindStyle(cssURL string) StyleConfig {
  for _, s := range cfg.Styles {
    if s.url == cssURL {
      return s
    }
  }

  panic("style " + cssURL + " not found")
}

func (cfg *SiteConfig) FindPage(pgURL string) PageConfig {
  for _, p := range cfg.Pages {
    if p.url == pgURL {
      return p
    }
  }

  panic("page " + pgURL + " not found")
}

func (cfg *SiteConfig) PageParameterString(pgURL string) string {
  styleURLs := cfg.PageStyles(pgURL)
  scriptBundleURL := cfg.JSURL()
  scriptHashes := cfg.PageScripts(pgURL)

  page := cfg.FindPage(pgURL)

  var b strings.Builder
  writeList := func(key string, lst []string) {
    b.WriteString(key)
    b.WriteString(":[")
    for i, x := range lst {
      b.WriteString(x)
      if i < len(lst) - 1 {
        b.WriteString(",")
      }
    }

    b.WriteString("]")
  }

  b.WriteString("{")
  writeList("parameters", page.params)
  writeList(",styles", styleURLs)
  writeList(",scripts", scriptHashes)
  b.WriteString(",scriptBundle:")
  b.WriteString(scriptBundleURL)
  b.WriteString("}")

  return b.String()
}

// remove all the files (and unneeded directories)
func (cfg *SiteConfig) CleanOutput() error {
  toKeep := make(map[string]string)

  keep := func(p string) {
    key := strings.TrimSpace(p)
    toKeep[key] = key

    dir := filepath.Dir(p)
    for dir != cfg.outputDir && len(dir) > 1 {
      key := strings.TrimSpace(dir)
      toKeep[key] = key
      dir = filepath.Dir(dir)
    }
  }
  
  keep(cfg.outputDir)

  for _, p := range cfg.Pages {
    keep(p.dst)
  }

  for _, f := range cfg.Files {
    keep(f.dst)
  }

  for _, s := range cfg.Styles {
    keep(s.dst)
  }

  keep(cfg.JSDst())
  keep(cfg.MathFontDst())

  toRemove := make([]string, 0)

  if err := filepath.Walk(cfg.outputDir, func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }

    if _, ok := toKeep[path]; !ok {
      toRemove = append(toRemove, path)
    }

    return nil
  }); err != nil {
    return err
  }

  for _, p := range toRemove {
    if err := os.RemoveAll(p); err != nil {
      return err
    }
  }

  return nil
}
