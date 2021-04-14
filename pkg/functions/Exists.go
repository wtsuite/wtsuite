package functions

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

// everything except null and undefined
func Exists(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

	if len(args) != 1 {
		return nil, ctx.NewError("Error: expected 1 argument")
	}

	arg1, err := args[0].Eval(scope)
	if err != nil {
		return nil, err
	}

	name, err := tokens.AssertString(arg1)
	if err != nil {
		return nil, err
	}

	fn := tokens.NewFunction("get", []tokens.Token{name, tokens.NewNull(ctx)}, ctx)

	res, err := fn.Eval(scope)
	if err != nil {
		panic(err)
	}

	resExists := !tokens.IsNull(res)

	return tokens.NewValueBool(resExists, ctx), nil
}
