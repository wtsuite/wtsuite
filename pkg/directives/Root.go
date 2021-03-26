package directives

import (
	"github.com/wtsuite/wtsuite/pkg/tree"
	//"github.com/wtsuite/wtsuite/pkg/tree/scripts"
)

func NewRoot(cache *FileCache, path string) (*tree.Root, error) {
	_, node, err := BuildFile(cache, path, true, nil)
	if err != nil {
		return nil, err
	}

  return FinalizeRoot(node)
}

func FinalizeRoot(node *RootNode) (*tree.Root, error) {

	root_ := node.tag
	root, ok := root_.(*tree.Root)
	if !ok {
		panic("expected root")
	}

	root.FoldDummy()

	tree.RegisterParents(root)

  if err := root.EvalLazy(); err != nil {
    return nil, err
  }

  // checks for uniqueness of id
	idMap := tree.NewIDMap()
	if err := root.CollectIDs(idMap); err != nil {
		return nil, err
	}

	if err := root.Validate(); err != nil {
		return nil, err
	}

  // apply @wrap of any inlined stylesheets
  // don't forget this for external stylesheets!
  for _, s := range node.sheets {
    var err error
    root, err = s.ApplyExtensions(root)
    if err != nil {
      return nil, err
    }
  }

	return root, nil
}
