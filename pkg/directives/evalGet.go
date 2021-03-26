package directives

import (
  "strings"

	"github.com/wtsuite/wtsuite/pkg/functions"
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
	"github.com/wtsuite/wtsuite/pkg/tokens/patterns"
)

// not the same as indexing in a function
func evalGet(scope Scope, args_ *tokens.Parens, ctx context.Context) (tokens.Token, error) {
  var err error
  args_, err = args_.EvalAsArgs(scope)
	if err != nil {
		return nil, err
	}

  args, err := functions.CompleteArgs(args_, nil)
  if err != nil {
    return nil, err
  }

	var fallback tokens.Token = nil
	if len(args) == 2 {
		fallback = args[1]
	} else if len(args) != 1 {
		return nil, ctx.NewError("Error: expected 1 or 2 arguments")
	}

	nameToken, err := tokens.AssertString(args[0])
	if err != nil {
		return nil, err
	}

	key := nameToken.Value()
	switch {
	case HasDefine(key):
		return GetDefine(key), nil
	case scope.HasVar(key): // prefer variable over builtin function
		return scope.GetVar(key).Value, nil
	case functions.HasFun(key):
		return functions.NewBuiltInFun(key, ctx), nil
	case fallback != nil:
		return fallback, nil
  case strings.Contains(key, "."):
    parts := strings.Split(key, ".") // TODO: also split the context
    partContexts := context.SplitByPeriod(nameToken.InnerContext(), len(parts))
    baseName := parts[0]
    if !patterns.IsValidVar(baseName) {
      errCtx := partContexts[0]
      return nil, errCtx.NewError("Error: not a valid var name")
    }

    var base tokens.Token = nil
    switch {
    case HasDefine(baseName):
      base = GetDefine(baseName)
    case scope.HasVar(baseName):
      base = scope.GetVar(baseName).Value
    case functions.HasFun(baseName):
      base = functions.NewBuiltInFun(baseName, ctx)
    default:
      errCtx := partContexts[0]
      return nil, errCtx.NewError("Error: '" + baseName + "' undefined")
    }

    // now evaluate a get function
    fn := tokens.NewFunction("get", []tokens.Token{base}, partContexts[0])

    for i, part := range parts {
      if i == 0 {
        continue
      }

      if !patterns.IsValidVar(part) {
        errCtx := partContexts[i]
        return nil, errCtx.NewError("Error: not a valid var name")
      }

      fn = tokens.NewFunction("get", []tokens.Token{
        fn,
        tokens.NewValueString(part, partContexts[i]),
      }, partContexts[i])
    }

    return fn.Eval(scope)
	default:
		errCtx := nameToken.InnerContext()
		err := errCtx.NewError("Error: variable '" + key + "' not defined")
		if key == ELEMENT_COUNT {
			context.AppendString(err, "Hint: "+ELEMENT_COUNT+" is only available inside tags")
		}

    vNames := formatValidVarNames(scope.ListValidVarNames())
		context.AppendString(err, "Info: available names\n"+vNames)
		return nil, err
	}
}
