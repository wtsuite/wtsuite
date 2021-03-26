package functions

import (
  "path/filepath"

	"github.com/wtsuite/wtsuite/pkg/files"
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

func Rel(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  if len(args) != 1 && len(args) != 2 {
    return nil, ctx.NewError("Error: expected 1 argument")
  }

  s, err := tokens.AssertString(args[0])
  if err != nil {
    return nil, err
  }

  rootStr := filepath.Dir(ctx.Path())

  if len(args) == 2 {
    root, err := AbsPath(args[1], args[1].Context())
    if err != nil {
      return nil, err
    }

    rootStr = root.Value()

    if !files.IsDir(rootStr) {
      errCtx := root.Context()
      return nil, errCtx.NewError("Error: not a directory")
    }
  }

  // rootStr should be valid here

  relPath, err := filepath.Rel(rootStr, s.Value())
  if err != nil {
    errCtx := ctx
    return nil, errCtx.NewError("Error: " + err.Error())
  }

  return tokens.NewValueString(relPath, ctx), nil
}
