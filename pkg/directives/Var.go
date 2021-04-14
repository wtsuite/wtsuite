package directives

import (
  "fmt"

	"github.com/wtsuite/wtsuite/pkg/functions"
	//"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
	"github.com/wtsuite/wtsuite/pkg/tokens/patterns"
)

func AssertValidVar(nameToken *tokens.String) error {
	errCtx := nameToken.Context()

	if !patterns.IsValidVar(nameToken.Value()) {
		return errCtx.NewError("Error: invalid var name")
	} else {
		return nil
	}
}

// doesnt change the node
func AddVar(scope Scope, node Node, tag *tokens.Tag) error {
	if !tag.IsEmpty() {
		errCtx := tag.Context()
		return errCtx.NewError("Error: unexpected child tags of var directive")
	}

	subScope := NewSubScope(scope)

	attr, err := tag.Attributes([]string{})
	if err != nil {
		return err
	}

	n := attr.Len()
	if n != 2 && n != 3 {
		errCtx := attr.Context()
		return errCtx.NewError("Error: expected 2 or 3 var attributes")
	}

	isExported := false
	if n == 3 {
		isExported, err = tokens.DictHasFlag(attr, "export")
		if err != nil {
			return err
		}

		if !isExported {
			errCtx := attr.Context()
			return errCtx.NewError("Error: if 3 attributes, expected export flag")
		}
	}

  namesToken, err := tokens.DictList(attr, "names")
  if err != nil {
    return err
  }

  valueToken, ok := attr.Get("value")
  if !ok {
    errCtx := attr.Context()
    return errCtx.NewError("Error: \"value\" attribute not found")
  }
  
  valueToken, err = valueToken.Eval(subScope)
  if err != nil {
    return err
  }

  nameTokens := namesToken.GetTokens()
  var valueParens *tokens.Parens = nil
  if len(nameTokens) != 1 && !tokens.IsParens(valueToken) {
    errCtx := valueToken.Context()
    return errCtx.NewError(fmt.Sprintf("Error: expected %d values, got 1 value", len(nameTokens)))
  } else if tokens.IsParens(valueToken) {
    valueParens, err = tokens.AssertParens(valueToken)
    if err != nil {
      panic(err)
    }

    if valueParens.Len() != len(nameTokens) {
      errCtx := valueToken.Context()
      return errCtx.NewError(fmt.Sprintf("Error: expected %d values, got %d value(s)", len(nameTokens), valueParens.Len()))
    }
  }

  for i, nameToken_ := range nameTokens {
    nameToken, err := tokens.AssertString(nameToken_)
    if err != nil {
      return err
    }

    if err := AssertValidVar(nameToken); err != nil {
      return err
    }

    key := nameToken.Value()
    switch {
    case HasGlobal(key):
      errCtx := nameToken.InnerContext()
      err := errCtx.NewError("Error: cant redefine global")
      return err
    case scope.HasVar(key):
      v := scope.GetVar(key)
      errCtx := nameToken.InnerContext() // unless context is identical, in which case we are probably inside a for loop
      if !(!errCtx.Less(&v.Ctx) && !v.Ctx.Less(&errCtx)) {
        err := errCtx.NewError("Error: cant redefine var")
        err.AppendContextString("Info: defined here", v.Ctx)
        return err
      }

      fallthrough
    default:
      var val tokens.Token
      if len(nameTokens) == 1 {
        val = valueToken
      } else {
        val = valueParens.Values()[i]
      }

      v := functions.Var{val, false, false, isExported,
        nameToken.InnerContext()}
      if err := scope.SetVar(key, v); err != nil {
        return err
      }
    }
  }

	return nil
}

var _addVarOk = registerDirective(patterns.TEMPLATE_VAR_KEYWORD, AddVar)
