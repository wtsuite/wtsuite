package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type URL struct {
  BuiltinPrototype
}

func NewURLPrototype() values.Prototype {
  return &URL{newBuiltinPrototype("URL")}
}

func NewURL(ctx context.Context) values.Value {
  return values.NewInstance(NewURLPrototype(), ctx)
}

func (p *URL) Check(other_ values.Interface, ctx context.Context) error {
  if _, ok := other_.(*URL); ok {
    return nil
  } else {
    return checkParent(p, other_, ctx)
  }
}

func (p *URL) GetInstanceMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  s := NewString(ctx)
  switch key {
  case "hash", "host", "hostname", "href", "origin", "password", "pathname", "port", "protocol", "search", "username":
    return s, nil
  case "searchParams":
    return NewURLSearchParams(ctx), nil
  default:
    return nil, nil
  }
}

func (p *URL) SetInstanceMember(key string, includePrivate bool, arg values.Value, ctx context.Context) error {
  s := NewString(ctx)

  switch key {
  case "hash", "host", "hostname", "href", "password", "pathname", "port", "protocol", "search", "username":
    return s.Check(arg, ctx)
  default:
    return ctx.NewError("Error: URL." + key + " not setable")
  }
}

func (p *URL) GetClassMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  s := NewString(ctx)

  switch key {
  case "createObjectURL":
    return values.NewFunction([]values.Value{NewBlob(ctx), s}, ctx), nil
  case "revokeObjectURL":
    return values.NewFunction([]values.Value{s, nil}, ctx), nil
  default:
    return nil, nil
  }
}

func (p *URL) GetClassValue() (*values.Class, error) {
  ctx := p.Context()
  s := NewString(ctx)

  return values.NewClass([][]values.Value{
    []values.Value{s},
    []values.Value{s, s},
  }, NewURLPrototype(), ctx), nil
}
