package files

import (
  "bytes"
  "encoding/base64"
  "encoding/gob"
  "fmt"
  "io/ioutil"
  "os"
  "path/filepath"
  "time"
)

const (
  CACHE_DIR_ENV_KEY = "WTCACHE"
  CACHE_REL_DIR = ".cache/wtsuite"
)

// dependency tree to determine if resource needs to be rebuilt or not

// node rebuild is triggered by:
// * node not in tree
// * parameter change
// * any dependency not in tree
// * any dependency has later modification time
type DepNode struct {
  IsDst        bool
  Parameters   string // if parameters change, node must be rebuilt
  Dependencies []string // if any dependencies c
}

// global rebuild is triggered by:
// * forceRebuild==true
// * tree not found
// * globals change
type DepTree struct {
  src     string // location of the cache file
  Globals string
  Nodes map[string]DepNode // key is dst
}

var _depTree *DepTree = nil

func depTreeFile(target string) string {
  var base string
  if cacheDir := os.Getenv(CACHE_DIR_ENV_KEY); cacheDir != "" {
    base = cacheDir
  } else {
    base = filepath.Join(os.Getenv("HOME"), CACHE_REL_DIR)
  }

  if err := os.MkdirAll(base, 0755); err != nil {
    panic(err)
  }

  key := base64.StdEncoding.EncodeToString([]byte(target))

  return filepath.Join(base, key)
}

func lastModified(path string) (time.Time, error) {
	status, err := os.Stat(path)
	if err != nil {
		return time.Now(), err
	}

	return status.ModTime(), nil
}

// target can be a directory or a file
func LoadDepTree(target string, globals string, forceRebuild bool) {
  src := depTreeFile(target)

  tree := &DepTree{
    src: src,
    Globals: globals,
    Nodes: make(map[string]DepNode),
  }

  if !forceRebuild {
    if IsFile(src) {
      b, err := ioutil.ReadFile(src)
      if err == nil {
        buf := bytes.NewBuffer(b)
        decoder := gob.NewDecoder(buf)

        decodeErr := decoder.Decode(&tree)
        if decodeErr != nil || tree.Globals != globals {
          tree = &DepTree{
            src: src,
            Globals: globals,
            Nodes: make(map[string]DepNode),
          }
        } 
      }
    } else if IsDir(src) {
      fmt.Fprintf(os.Stderr, "Error: dep tree file is directory, this shouldn't be possible")
      os.Exit(1)
    }
  }

  _depTree = tree
}

func StartDstUpdate(dst string, parameters string) {
  if _depTree == nil {
    return
  }

  _depTree.Nodes[dst] = DepNode{
    IsDst:        true,
    Parameters:   parameters,
    Dependencies: []string{},
  }
}

func StartDepUpdate(name string, parameters string) {
  if _depTree == nil {
    return
  }

  _depTree.Nodes[name] = DepNode{
    IsDst:        false,
    Parameters:   parameters,
    Dependencies: []string{},
  }
}

func AddDep(this string, dep string) {
  if _depTree == nil {
    return
  }

  node, ok := _depTree.Nodes[this]
  if !ok {
    panic(this + " not found in _depTree (hint: call files.StartDepUpdate()")
  }

  for _, d := range node.Dependencies {
    if d == dep {
      return
    }
  }

  node.Dependencies = append(node.Dependencies, dep)

  _depTree.Nodes[this] = node
}

func HasUpstreamDep(thisPath string, upstreamPath string) bool {
  if _depTree == nil {
    return false
  }

  node, ok := _depTree.Nodes[thisPath]
  if !ok {
    return false
  }

  for _, d := range node.Dependencies {
    if d == upstreamPath || HasUpstreamDep(d, upstreamPath) {
      return true
    }
  }
  
  return false
}

// dst is the actual file path!
func RequiresDepUpdate(this string, parameters string) bool {
  if _depTree == nil {
    return true
  }

  if !IsFile(this) {
    return true
  }

  node, ok := _depTree.Nodes[this]
  if !ok {
    return true
  }

  if !node.IsDst {
    return true
  }

  if node.Parameters != parameters {
    return true
  }

  thisTime, thisTimeErr := lastModified(this)
  if thisTimeErr != nil {
    return true
  }

  for _, dep := range node.Dependencies {
    depTime, depTimeErr := lastModified(dep)
    if depTimeErr != nil {
      return true
    }

    if depTime.After(thisTime) {
      return true
    }
  }

  return false
}

func (t *DepTree) clean() {
  touched := make(map[string]bool)

  var touchUpwards func(node DepNode) = nil
  touchUpwards = func(node DepNode) {
    for _, dep := range node.Dependencies {
      if _, ok := touched[dep]; !ok {
        subNode, ok := _depTree.Nodes[dep]
        if ok {
          touched[dep] = true

          touchUpwards(subNode)
        }
      }
    }
  }

  for k, node := range _depTree.Nodes {
    if node.IsDst {
      if !IsFile(k) {
        delete(_depTree.Nodes, k)
      } else {
        touched[k] = true

        touchUpwards(node)
      }
    }
  }

  for k, _ := range _depTree.Nodes {
    if _, ok := touched[k]; !ok {
      delete(_depTree.Nodes, k)
    }
  }
}

func SaveDepTree() {
  if _depTree == nil {
    return
  }

  _depTree.clean()

  buf := bytes.Buffer{}

  encoder := gob.NewEncoder(&buf)

  if err := encoder.Encode(_depTree); err != nil {
    fmt.Fprintf(os.Stderr, "Warning: " + err.Error())
  } else {
    if err := ioutil.WriteFile(_depTree.src, buf.Bytes(), 0644); err != nil {
      fmt.Fprintf(os.Stderr, "Warning: " + err.Error())
    }
  }
}
