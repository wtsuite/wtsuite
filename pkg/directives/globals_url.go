package directives

import (
	"github.com/computeportal/wtsuite/pkg/functions"
	"github.com/computeportal/wtsuite/pkg/tokens/context"
	tokens "github.com/computeportal/wtsuite/pkg/tokens/html"
)

const URL = "__url__"
var IGNORE_UNSET_URLS = false

var _fileURLs map[string]string = nil
var _activeURL *tokens.String = nil

// path is src path
func RegisterURL(path string, url string) {
	if _fileURLs == nil {
		_fileURLs = make(map[string]string)
	}

	_fileURLs[path] = url
}

func GetActiveURL(ctx context.Context) (*tokens.String, error) {
	if _activeURL == nil {
		return nil, ctx.NewError("Error: __url__ not set here")
	}

	return _activeURL, nil
}

func SetActiveURL(url string) {
	_activeURL = tokens.NewValueString(url, context.NewDummyContext())
}

func UnsetActiveURL() {
	_activeURL = nil
}

func evalFileURL(scope Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  args, err := functions.CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

  if len(args) == 0 {
    return GetActiveURL(ctx)
  }

	if len(args) != 1 {
		return nil, ctx.NewError("Error: expected 0 or 1 arguments")
  }

  arg0, err := args[0].Eval(scope)
  if err != nil {
    return nil, err
  }

  filePathToken, err := functions.AbsPath(arg0, ctx)
  if err != nil {
    return nil, err
  }

  filePath := filePathToken.Value()

	if url, ok := _fileURLs[filePath]; ok {
		return tokens.NewValueString(url, ctx), nil
	} else {
    if !IGNORE_UNSET_URLS {
      return nil, ctx.NewError("Error: url for '" + filePath + "' not set")
    } else {
      // used when doing refactorings, where the url doesnt matter
      return tokens.NewValueString("", ctx), nil
    }
	}
}
