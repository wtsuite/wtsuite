package directives

import (
	"github.com/wtsuite/wtsuite/pkg/tree"
)

type StyleSheet interface {
  ApplyExtensions(root *tree.Root) (*tree.Root, error)
}
