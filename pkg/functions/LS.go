package functions

import (
  "io/ioutil"
  "os"
  "path/filepath"

	"github.com/computeportal/wtsuite/pkg/files"
	"github.com/computeportal/wtsuite/pkg/tokens/context"
	tokens "github.com/computeportal/wtsuite/pkg/tokens/html"
)

// doesnt list the directories!
// list files in the order they are encountered
func listFiles(root *tokens.String, recursive bool, ctx context.Context) (tokens.Token, error) {
  lst := tokens.NewEmptyList(ctx)

  if recursive {
    if err := filepath.Walk(root.Value(), func(path string, info os.FileInfo, err error) error {
      if err != nil {
        return err
      }

      if !info.IsDir() {
        lst.Append(tokens.NewValueString(path, ctx))
      }

      return nil
    }); err != nil {
      errCtx := ctx
      return nil, errCtx.NewError("Error: " + err.Error())
    }
  } else {
    infos, err := ioutil.ReadDir(root.Value())
    if err != nil {
      errCtx := ctx
      return nil, errCtx.NewError("Error: " + err.Error())
    }

    for _, info := range infos {
      lst.Append(tokens.NewValueString(info.Name(), ctx))
    }
    // just use ls, and ignore the directories
  }

  return lst, nil
}

func LS(scope tokens.Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  recursive := true
  root := tokens.NewValueString(filepath.Dir(ctx.Path()), ctx)

  if len(args) == 1 || len(args) == 2 {
    root, err = AbsPath(args[0], args[0].Context())
    if err != nil {
      return nil, err
    }

    if !files.IsDir(root.Value()) {
      errCtx := root.Context()
      return nil, errCtx.NewError("Error: not a directory")
    }
  }

  if len(args) == 2 {
    arg1, err := tokens.AssertBool(args[1])
    if err != nil {
      return nil, err
    }

    recursive = arg1.Value()
  } else if len(args) > 2 {
    errCtx := ctx
    return nil, errCtx.NewError("Error: expected 2 arguments at most")
  }



  return listFiles(root, recursive, ctx)
}
