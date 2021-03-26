package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type Navigator struct {
  BuiltinPrototype
}

func NewNavigatorPrototype() values.Prototype {
  return &Navigator{newBuiltinPrototype("Navigator")}
}

func NewNavigator(ctx context.Context) values.Value {
  return values.NewInstance(NewNavigatorPrototype(), ctx)
}

func (p *Navigator) GetParent() (values.Prototype, error) {
  return NewEventTargetPrototype(), nil
}

func (p *Navigator) Check(other_ values.Interface, ctx context.Context) error {
  if _, ok := other_.(*Navigator); ok {
    return nil
  } else {
    return checkParent(p, other_, ctx)
  }
}

func (p *Navigator) GetInstanceMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  b := NewBoolean(ctx)
  i := NewInt(ctx)
  s := NewString(ctx)

  switch key {
  case "cookieEnabled":
    return b, nil
  case "language":
    return s, nil
  case "maxTouchPoints":
    return i, nil
  case "onLine":
    return b, nil
  case "sendBeacon":
    return values.NewOverloadedMethodLikeFunction([][]values.Value{
      []values.Value{s, b},
      []values.Value{s, NewArrayBuffer(ctx), b},
      []values.Value{s, NewBlob(ctx), b},
      []values.Value{s, NewString(ctx), b},
      []values.Value{s, NewURLSearchParams(ctx), b},
    }, ctx), nil
  case "userAgent":
    return s, nil
  default:
    return nil, nil
  }
}

func (p *Navigator) GetClassValue() (*values.Class, error) {
  ctx := p.Context()
  return values.NewUnconstructableClass(NewNavigatorPrototype(), ctx), nil
}
