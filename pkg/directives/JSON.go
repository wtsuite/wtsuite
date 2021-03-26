package directives

import (
  "errors"
  "io/ioutil"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

func BuildJSON(path string, ctx context.Context) (tokens.Token, error) {
  cache := NewFileCache()
  
  scope, _, err := BuildFile(cache, path, false, nil)
  if err != nil {
    return nil, err
  }

  if !scope.HasVar("main") {
    errCtx := ctx
    return nil, errCtx.NewError("Error: var \"main\" not found in \"" + path + "\"")
  }

  v := scope.GetVar("main")
  if !v.Exported {
    errCtx := ctx
    return nil, errCtx.NewError("Error: var \"main\" not exported from \"" + path + "\"")
  }

  t := v.Value
  if !(tokens.IsStringDict(t) || tokens.IsList(t)) {
    errCtx := t.Context()
    return nil, errCtx.NewError("Error: expected list or dict")
  }

  return t, nil
}

func BuildJSONFile(input string, outputPath string) error {
  token, err := BuildJSON(input, context.NewDummyContext())
  if err != nil {
    return err
  }

  return WriteJSONToFile(token, outputPath)
}

func WriteJSONToFile(token tokens.Token, path string) error {
  content, err := tokens.WriteJSON(token)
  if err != nil {
    return err
  }

  if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
    return errors.New("Error: " + err.Error())
  }

  return nil
}
