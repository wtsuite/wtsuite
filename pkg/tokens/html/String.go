package html

import (
	"reflect"
  "strings"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	"github.com/wtsuite/wtsuite/pkg/tokens/patterns"
)

type String struct {
	value   string
	TokenData
}

func NewValueString(value string, ctx context.Context) *String {
	return &String{value, TokenData{ctx}}
}

func NewString(value string, ctx context.Context) (*String, error) {
	return NewValueString(value, ctx), nil
}

func NewDummyContextString(value string) *String {
	return &String{value, TokenData{context.NewDummyContext()}}
}

func (t *String) Value() string {
	return t.value
}

func (t *String) Len() int {
	return len(t.value)
}

func (t *String) Eval(scope Scope) (Token, error) {
	return t, nil
}

func (t *String) EvalLazy(tag FinalTag) (Token, error) {
	return t, nil
}

func (t *String) Write() string {
	// without the quotes
	return t.value
}

func (t *String) IsLiteral() bool {
  ctx := t.Context()
  ctxStr := ctx.Content()

  return strings.HasPrefix(ctxStr, "\"") || strings.HasPrefix(ctxStr, "'")
}

func (t *String) Dump(indent string) string {
  var b strings.Builder

  b.WriteString(indent)
  b.WriteString("String(")
  b.WriteString(t.Write())
  b.WriteString(")\n")

  return b.String()
}

func (t *String) Echo(indent string) string {
  var b strings.Builder

  b.WriteString(indent)
  b.WriteString("\"")
  b.WriteString(t.Write())
  b.WriteString("\"")

	return b.String()
}

func IsString(t Token) bool {
	_, ok := t.(*String)
	return ok
}

func IsValueString(t_ Token, val string) bool {
	t, ok := t_.(*String)
	if ok {
    return t.Value() == val
  } else {
    return false
  }
}

func IsNonLiteralValueString(t_ Token, val string) bool {
	t, ok := t_.(*String)
	if ok {
    return (t.Value() == val) && (!t.IsLiteral())
  } else {
    return false
  }
}

func AssertString(t Token) (*String, error) {
	if s, ok := t.(*String); !ok {
		errCtx := t.Context()
		err := errCtx.NewError("Error: expected string (got " + reflect.TypeOf(t).String() + ")")
		return nil, err
	} else {
		return s, nil
	}
}

func AssertWord(t Token) (*String, error) {
  errCtx := t.Context()
	if s, ok := t.(*String); !ok {
		err := errCtx.NewError("Error: expected string (got " + reflect.TypeOf(t).String() + ")")
    panic(err)
		return nil, err
	} else {
    if !patterns.IsValidVar(s.Value()) {
      return nil, errCtx.NewError("Error: not a valid word (" + s.Value() + ")")
    }

		return s, nil
	}
}

func (t *String) InnerContext() context.Context {
	n := len(t.value)
	if n == t.ctx.Len()-2 {
		return t.ctx.NewContext(1, n+1)
	} else {
		return t.TokenData.Context()
	}
}

func (a *String) IsSame(other Token) bool {
	if b, ok := other.(*String); ok {
		return a.value == b.value
	} else {
		return false
	}
}

func (t *String) TrimLeft(prefix string) *String {
  if strings.HasPrefix(t.value, prefix) {

    newValue := strings.TrimLeft(t.value, prefix)

    ctx := t.Context()
    /*if ctx.Content() == t.value {
      // also trim the context

      ctx = ctx.NewContext(len(prefix), 0)
    }*/

    return NewValueString(newValue, ctx)
  } else {
    return t
  }
}
