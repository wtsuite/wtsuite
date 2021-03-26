package functions

import (
  "regexp"

	"github.com/computeportal/wtsuite/pkg/tokens/context"
	tokens "github.com/computeportal/wtsuite/pkg/tokens/html"
)

// check if regexp matches part of the string
func Matches(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  if len(args) != 2 {
    return nil, ctx.NewError("Error: expected 2 arguments")
  }

  arg0, err := tokens.AssertString(args[0])
  if err != nil {
    return nil, err
  }

  arg1, err := tokens.AssertString(args[1])
  if err != nil {
    return nil, err
  }

  re, err := regexp.Compile(arg1.Value())
  if err != nil {
    errCtx := arg1.Context()
    return nil, errCtx.NewError("Error: not a valid regexp (" + err.Error() + ")")
  }

  ok := re.MatchString(arg0.Value())

  return tokens.NewValueBool(ok, ctx), nil
}
