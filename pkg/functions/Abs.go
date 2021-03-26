package functions

import (
  "path/filepath"
  "math"

	"github.com/computeportal/wtsuite/pkg/files"
	"github.com/computeportal/wtsuite/pkg/tokens/context"
	tokens "github.com/computeportal/wtsuite/pkg/tokens/html"
)

// file or dir must exist!
func AbsPath(arg_ tokens.Token, ctx context.Context) (*tokens.String, error) {
  arg, err := tokens.AssertString(arg_)
  if err != nil {
    return nil, err
  }

  var result string
  if filepath.IsAbs(arg.Value()) {
    result = arg.Value()
  } else {
    result = filepath.Join(filepath.Dir(ctx.Path()), arg.Value())
  }

  if !files.Exists(result) {
    errCtx := arg.Context()
		return nil, errCtx.NewError("Error: couldn't find directory or file \"" + arg.Value() + "\"")
  }

	return tokens.NewValueString(result, ctx), nil
}

func absInt(arg_ tokens.Token, ctx context.Context) (tokens.Token, error) {
  arg, err := tokens.AssertInt(arg_)
  if err != nil {
    return nil, err
  }

  argValue := int(math.Abs(float64(arg.Value())))

  return tokens.NewValueInt(argValue, ctx), nil
}

func absFloat(arg_ tokens.Token, ctx context.Context) (tokens.Token, error) {
  arg, err := tokens.AssertAnyIntOrFloat(arg_)
  if err != nil {
    return nil, err
  }

  argValue := math.Abs(arg.Value())

  return tokens.NewValueUnitFloat(argValue, arg.Unit(), ctx), nil
}

// abs(".") or abs("./") returns the current directory, not the current file!
// abs() returns the current file

func Abs(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  if len(args) == 0 {
    return tokens.NewValueString(ctx.Path(), ctx), nil
  } else if len(args) != 1 {
		return nil, ctx.NewError("Error: expected 1 argument")
	}

  switch {
  case tokens.IsString(args[0]):
    return AbsPath(args[0], ctx)
  case tokens.IsInt(args[0]):
    return absInt(args[0], ctx)
  case tokens.IsFloat(args[0]):
    return absFloat(args[0], ctx)
  default:
    return nil, ctx.NewError("Error: expected string or number argument")
  }
}
