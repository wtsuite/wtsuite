package html

import (
  "strconv"
	"strings"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type Function struct {
	name string
	args *Parens
	TokenData
}

func NewValueFunction(name string, args *Parens, ctx context.Context) *Function {
	return &Function{name, args, TokenData{ctx}}
}

func NewFunction(name string, args []Token, ctx context.Context) *Function {
	return NewValueFunction(name, NewParens(args, nil, ctx), ctx)
}

func (t *Function) Dump(indent string) string {
	var b strings.Builder

	b.WriteString(indent + "Function " + t.name + "\n")

	for i, arg := range t.args.values {
    alt := t.args.alts[i]

    if alt == nil {
      b.WriteString(arg.Dump(indent + "  "))
    } else {
      b.WriteString(arg.Dump(indent + "  "))
      b.WriteString(alt.Dump(indent + "= "))
    }
	}

	return b.String()
}

func IsAnyFunction(t Token) bool {
	_, ok := t.(*Function)
	return ok
}

func IsFunction(t Token, name string) bool {
	if fn, ok := t.(*Function); ok {
		return fn.name == name
	}

	return false
}

func AssertFunction(t Token) (*Function, error) {
	if fn, ok := t.(*Function); ok {
		return fn, nil
	} else {
		errCtx := t.Context()
		return nil, errCtx.NewError("Error: expected a function")
	}
}

func (t *Function) EvalPartial(scope Scope) (Token, error) {
  outerVals := make([]Token, 0)
  outerAlts := make([]Token, 0)

  innerVals := make([]Token, 0)
  innerAlts := make([]Token, 0)

  anyPartial := false

  for i, v := range t.args.Values() {
    a := t.args.Alts()[i]
    j := len(outerVals)
    outerName := "x" + strconv.Itoa(j)
    outerArg := NewValueString(outerName, v.Context())
    innerArg := NewFunction("get", []Token{outerArg}, v.Context())

    if IsNonLiteralValueString(v, "_") {
      if a != nil {
        errCtx := a.Context()
        return nil, errCtx.NewError("Error: unexpected kwarg")
      }

      outerVals = append(outerVals, outerArg)
      outerAlts = append(outerAlts, nil)

      innerVals = append(innerVals, innerArg)
      innerAlts = append(innerAlts, nil)

      anyPartial = true
    } else if IsNonLiteralValueString(a, "_") {
      if _, err := AssertString(v); err != nil {
        return nil, err
      }

      outerVals = append(outerVals, v)
      outerAlts = append(outerAlts, outerArg)

      innerVals = append(innerVals, v)
      innerAlts = append(innerAlts, innerArg)

      anyPartial = true
    } else {
      innerVals = append(innerVals, v)
      innerAlts = append(innerAlts, a)
    }
  }

  if anyPartial {
    outerArgs := NewParens(outerVals, outerAlts, t.Context())
    innerArgs := NewParens(innerVals, innerAlts, t.Context())

    inner := NewValueFunction(t.name, innerArgs, t.Context())

    fn := NewFunction("function", []Token{outerArgs, inner}, t.Context())

    return fn.Eval(scope)
  } else {
    return nil, nil
  }
}

func (t *Function) Eval(scope Scope) (Token, error) {
  // if any of the args are underscores then a partially applied function is returned
  partial, err := t.EvalPartial(scope); 
  if err != nil {
    panic(err)
    return nil, err
  }

  if partial != nil {
    return partial, nil
  }

  res, err := scope.Eval(t.name, t.args, t.Context())
  if err != nil {
    return nil, err
  }

  if _, ok := res.(*Function); ok {
    errCtx := res.Context()
    err := errCtx.NewError("Internal Error: result of an eval can't be another function")
    panic(err)
  }

  return ChangeContext(res, t.Context()), nil
}

func (t *Function) EvalLazy(tag FinalTag) (Token, error) {
  errCtx := t.Context()
  return nil, errCtx.NewError("Error: unable to evaluate lazily")
}

func (a *Function) IsSame(other Token) bool {
	if b, ok := other.(*Function); ok {
		if a.name == b.name {
      if a.args.IsSame(b.args) {
        return true
      }
		}
	}

	return false
}

func (t *Function) Args() *Parens {
	return t.args
}
