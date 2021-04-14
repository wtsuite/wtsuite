package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type History struct {
  BuiltinPrototype
}

func NewHistoryPrototype() values.Prototype {
  return &History{newBuiltinPrototype("History")}
}

func NewHistory(ctx context.Context) values.Value {
  return values.NewInstance(NewHistoryPrototype(), ctx)
}

func (p *History) GetParent() (values.Prototype, error) {
  return NewEventTargetPrototype(), nil
}

func (p *History) Check(other_ values.Interface, ctx context.Context) error {
  if _, ok := other_.(*History); ok {
    return nil
  } else {
    return checkParent(p, other_, ctx)
  }
}

func (p *History) GetInstanceMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  i := NewInt(ctx)
  s := NewString(ctx)
  a := values.NewAny(ctx)

  switch key {
  case "back", "forward":
    return values.NewFunction([]values.Value{nil}, ctx), nil
  case "go":
    return values.NewFunction([]values.Value{i, nil}, ctx), nil
  case "length":
    return i, nil
  case "pushState", "replaceState":
    return values.NewOverloadedFunction([][]values.Value{
      []values.Value{a, s, nil},
      []values.Value{a, s, s, nil},
    }, ctx), nil
  case "state":
    return a, nil
  default:
    return nil, nil
  }
}

func (p *History) GetClassValue() (*values.Class, error) {
  ctx := p.Context()
  return values.NewUnconstructableClass(NewHistoryPrototype(), ctx), nil
}
