package functions

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

func Any(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  if len(args) != 1 {
    return nil, ctx.NewError("Error: expected 1 argument")
  }

  lst, err := tokens.AssertList(args[0])
  if err != nil {
    return nil, err
  }

  any := false

  if err := lst.LoopValues(func(t tokens.Token) error {
    b, err := tokens.AssertBool(t)
    if err != nil {
      return err
    }

    if b.Value() {
      any = true
    }

    return nil
  }); err != nil {
    return nil, err
  }

  return tokens.NewValueBool(any, ctx), nil
}
