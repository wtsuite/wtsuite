package directives

import (
  "fmt"

	"github.com/wtsuite/wtsuite/pkg/functions"
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

func evalNew(scope Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  var err error
	args_, err = args_.EvalAsArgs(scope)
	if err != nil {
		return nil, err
	}

  args, err := functions.CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

	if len(args) < 2 {
		return nil, ctx.NewError("Error: expected at least 2 arguments")
	}

  n := len(args) - 1

	valueToken := args[n]
  var valueParens *tokens.Parens = nil

  if n == 1 {
    if tokens.IsParens(valueToken) {
      valueParens, err = tokens.AssertParens(valueToken)
      if err != nil {
        panic(err)
      }

      errCtx := ctx
      return nil, errCtx.NewError(fmt.Sprintf("Error: expected %d return values, got %d", n, valueParens.Len()))
    }
  } else if n != 1 {
    if !tokens.IsParens(valueToken) {
      errCtx := ctx
      return nil, errCtx.NewError(fmt.Sprintf("Error: expected %d return values, got 1", n))
    }

    valueParens, err = tokens.AssertParens(valueToken)
    if err != nil {
      panic(err)
    }

    if n != valueParens.Len() {
      errCtx := ctx
      return nil, errCtx.NewError(fmt.Sprintf("Error: expected %d return values, got %d", n, valueParens.Len()))
    }
  }

  for i := 0; i < n; i++ {
    arg := args[i]

    nameToken, err := tokens.AssertString(arg)
    if err != nil {
      return nil, err
    }

    if err := AssertValidVar(nameToken); err != nil {
      return nil, err
    }

    key := nameToken.Value()
    switch {
    case HasGlobal(key):
      errCtx := nameToken.InnerContext()
      return nil, errCtx.NewError("Error: can't redefine global")
    case scope.HasVar(key):
      errCtx := nameToken.InnerContext()
      return nil, errCtx.NewError("Error: can't redefine variable")
    default:
      var val tokens.Token
      if valueParens != nil {
        val = valueParens.Values()[i]
      } else {
        val = valueToken
      }

      v := functions.Var{val, false, false, false, ctx}
      if err := scope.SetVar(key, v); err != nil {
        return nil, err
      }
    }
  }


	return valueToken, nil
}
