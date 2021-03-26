package functions

import (
	"path/filepath"
  "strings"

	"github.com/computeportal/wtsuite/pkg/tokens/context"
	tokens "github.com/computeportal/wtsuite/pkg/tokens/html"
)

func NoExt(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

	if len(args) != 1 {
		return nil, ctx.NewError("Error: expected 1 argument")
	}

	path, err := tokens.AssertString(args[0])
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(path.Value())

  noExt := strings.TrimSuffix(path.Value(), ext)

	return tokens.NewValueString(noExt, ctx), nil
}
