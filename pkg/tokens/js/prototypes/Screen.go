package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type Screen struct {
  BuiltinPrototype
}

func NewScreenPrototype() values.Prototype {
  return &Screen{newBuiltinPrototype("Screen")}
}

func NewScreen(ctx context.Context) values.Value {
  return values.NewInstance(NewScreenPrototype(), ctx)
}

func (p *Screen) GetParent() (values.Prototype, error) {
  return NewEventTargetPrototype(), nil
}

func (p *Screen) Check(other_ values.Interface, ctx context.Context) error {
  if _, ok := other_.(*Screen); ok {
    return nil
  } else {
    return checkParent(p, other_, ctx)
  }
}

func (p *Screen) GetInstanceMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  i := NewInt(ctx)

  switch key {
  case "availHeight", "availWidth", "colorDepth", "height", "pixelDepth", "width":
    return i, nil
  default:
    return nil, nil
  }
}

func (p *Screen) GetClassValue() (*values.Class, error) {
  ctx := p.Context()
  return values.NewUnconstructableClass(NewScreenPrototype(), ctx), nil
}
