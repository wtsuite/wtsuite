package values

import (
  "strings"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
)

// hidden name
const TUPLE = ".tuple"

type Tuple struct {
  items []Value
  arrProto Prototype

  ValueData
}

func NewTuple(items []Value, ctx context.Context) *Tuple {
  return &Tuple{items, nil, newValueData(ctx)}
}

func NewLiteralTuple(items []Value, arrProto Prototype, ctx context.Context) *Tuple {
  return &Tuple{items, arrProto, newValueData(ctx)}
}

func (v *Tuple) TypeName() string {
  var b strings.Builder
  b.WriteString("[")

  for i, item := range v.items {
    b.WriteString(item.TypeName())

    if i < len(v.items) - 1 {
      b.WriteString(",")
    }
  }

  b.WriteString("]")

  return b.String()
}

func (v *Tuple) GetInterface() Interface {
  return v.arrProto
}

// if arrProto is defined (i.e. literal) then check as array
// if not isLiteral then other must be Tuple as well (from which can't be inherited)
func (v *Tuple) Check(other_ Value, ctx context.Context) error {
  if IsAny(other_) {
    return nil
  } else if v.arrProto != nil {
    instance := NewInstance(v.arrProto, ctx)
    return instance.Check(other_, ctx)
  } else {
    other_ = UnpackContextValue(other_)

    if other, ok := other_.(*Tuple); ok {
      if len(v.items) == len(other.items) {
        someErr := false
        for i, v := range v.items {
          otherV := other.items[i]
          if err := v.Check(otherV, ctx); err != nil {
            someErr = true
            break
          }
        }

        if !someErr {
          return nil
        }
      }
    } 

    return ctx.NewError("Error: expected " + v.TypeName() + ", got " + other_.TypeName())
  }
}

func (v *Tuple) EvalConstructor(args []Value, ctx context.Context) (Value, error) {
  return nil, ctx.NewError("Error: not a constructor")
}

func (v *Tuple) EvalFunction(args []Value, preferMethod bool, ctx context.Context) (Value, error) {
  if v.arrProto != nil {
    return nil, ctx.NewError("Error: can't call an instance")
  } else {
    return nil, ctx.NewError("Error: can't call a tuple")
  }
}

func (v *Tuple) GetMember(key string, includePrivate bool,
	ctx context.Context) (Value, error) {
  a := NewAny(ctx)

  switch key {
  case ".getindex":
    return NewCustomFunction([]Value{a}, func(args []Value, preferMethod bool, ctx_ context.Context) (Value, error) {
      if lit, ok := args[0].LiteralIntValue(); ok && v.items != nil {
        if lit < 0 || lit >= len(v.items) {
          errCtx := args[0].Context()
          return nil, errCtx.NewError("Error: out of tuple range")
        }

        return NewContextValue(v.items[lit], ctx_), nil
      } else {
        if v.arrProto != nil {
          interf, err := FindInstanceMemberInterface(v.arrProto, key, includePrivate, ctx)
          if err != nil {
            return nil, err
          }

          fn, err := interf.GetInstanceMember(key, includePrivate, ctx)
          if err != nil {
            return nil, err
          }

          return fn.EvalFunction(args, preferMethod, ctx_)
        } else {
          return a, nil
        }
      }
    }, ctx), nil
  case ".setindex":
    return NewCustomFunction([]Value{a, a}, func(args []Value, preferMethod bool, ctx_ context.Context) (Value, error) {
      if lit, ok := args[0].LiteralIntValue(); ok && v.items != nil {
        if lit < 0 || lit >= len(v.items) {
          errCtx := args[0].Context()
          return nil, errCtx.NewError("Error: out of tuple range")
        }

        checkVal := v.items[lit]
        return nil, checkVal.Check(args[1], args[1].Context())
      } else if v.arrProto != nil {
        interf, err := FindInstanceMemberInterface(v.arrProto, key, includePrivate, ctx)
        if err != nil {
          return nil, err
        }

        fn, err := interf.GetInstanceMember(key, includePrivate, ctx)
        if err != nil {
          return nil, err
        }

        return fn.EvalFunction(args, preferMethod, ctx_)
      } else {
        return nil, ctx.NewError("Error: can't set index of tuple with non-literal index")
      }
    }, ctx), nil
  default:
    if v.arrProto != nil {
      interf, err := FindInstanceMemberInterface(v.arrProto, key, includePrivate, ctx)
      if err != nil {
        return nil, err
      }

      return interf.GetInstanceMember(key, includePrivate, ctx)
    } else {
      return nil, ctx.NewError("Error: can't get member of tuple")
    }
  }
}

func (v *Tuple) SetMember(key string, includePrivate bool, arg Value,
	ctx context.Context) error {
  if v.arrProto != nil {
    interf, err := FindInstanceMemberInterface(v.arrProto, key, includePrivate, ctx)
    if err != nil {
      return err
    }

    return interf.SetInstanceMember(key, includePrivate, arg, ctx)
  } else {
    return ctx.NewError("Error: can't set member of tuple")
  }
}
