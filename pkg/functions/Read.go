package functions

import (
  "encoding/json"
  "io/ioutil"

	"github.com/computeportal/wtsuite/pkg/files"
	"github.com/computeportal/wtsuite/pkg/tokens/context"
	tokens "github.com/computeportal/wtsuite/pkg/tokens/html"
)

func Read(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  if len(args) != 1 && len(args) != 2 {
    errCtx := ctx
    return nil, errCtx.NewError("Error: expected 1 or 2 argument")
  }

  fPath, err := tokens.AssertString(args[0])
  if err != nil {
    return nil, err
  }

  encoding := "utf-8"
  if len(args) == 2 {
    arg1, err := tokens.AssertString(args[1])
    if err != nil {
      return nil, err
    }

    encoding = arg1.Value()
  }

  fPathAbs, err := AbsPath(fPath, ctx)
  if err != nil {
    return nil, err
  }

  b, err := ioutil.ReadFile(fPathAbs.Value())
  if err != nil {
    errCtx := fPath.Context()
    return nil, errCtx.NewError("Error: " + err.Error())
  }

  files.AddDep(ctx.Path(), fPathAbs.Value())

  switch encoding {
  case "utf-8":
    return tokens.NewValueString(string(b), ctx), nil
  case "json":
    var obj interface{}
    if err := json.Unmarshal(b, &obj); err != nil {
      errCtx := fPath.Context()
      return nil, errCtx.NewError("Error: " + err.Error())
    }

    return tokens.GolangToToken(obj, ctx)
  default:
    errCtx := args[1].Context()
    return nil, errCtx.NewError("Error: encoding \"" + encoding + "\" not handled (hint: utf-8 or json)")
  }
}
