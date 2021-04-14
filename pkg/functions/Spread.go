package functions

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

func Spread(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  if len(args) != 1 {
    errCtx := ctx
    return nil, errCtx.NewError("Error: expected 1 argument")
  }

  arg0, err := tokens.AssertList(args[0])
  if err != nil {
    return nil, err
  }

  // put into parens and return
  return tokens.NewParens(arg0.GetTokens(), nil, ctx), nil
}
