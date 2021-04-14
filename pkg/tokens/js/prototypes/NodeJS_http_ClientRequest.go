package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type NodeJS_http_ClientRequest struct {
  BuiltinPrototype
}

func NewNodeJS_http_ClientRequestPrototype() values.Prototype {
  return &NodeJS_http_ClientRequest{newBuiltinPrototype("ClientRequest")}
}

func NewNodeJS_http_ClientRequest(ctx context.Context) values.Value {
  return values.NewInstance(NewNodeJS_http_ClientRequestPrototype(), ctx)
}

func (p *NodeJS_http_ClientRequest) GetParent() (values.Prototype, error) {
  return NewNodeJS_EventEmitterPrototype(), nil
}

func (p *NodeJS_http_ClientRequest) Check(other_ values.Interface, ctx context.Context) error {
  if _, ok := other_.(*NodeJS_http_ClientRequest); ok {
    return nil
  } else {
    return checkParent(p, other_, ctx)
  }
}

func (p *NodeJS_http_ClientRequest) GetInstanceMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  b := NewBoolean(ctx)
  s := NewString(ctx)
  buf := NewNodeJS_Buffer(ctx)
  this := NewNodeJS_http_ClientRequest(ctx)

  switch key {
  case "end":
    return values.NewOverloadedMethodLikeFunction([][]values.Value{
      []values.Value{this},
      []values.Value{s, this},
      []values.Value{buf, this},
      []values.Value{s, s, this},
      []values.Value{buf, s, this},
    }, ctx), nil
  case "write":
    callback := values.NewFunction([]values.Value{nil}, ctx)

    return values.NewOverloadedMethodLikeFunction([][]values.Value{
      []values.Value{s, b},
      []values.Value{s, s, b},
      []values.Value{s, callback, b},
      []values.Value{s, s, callback, b},
      []values.Value{buf, b},
      []values.Value{buf, s, b},
      []values.Value{buf, callback, b},
      []values.Value{buf, s, callback, b},
    }, ctx), nil
  default:
    return nil, nil
  }
}

func (p *NodeJS_http_ClientRequest) SetInstanceMember(key string, includePrivate bool, arg values.Value, ctx context.Context) error {
  switch key {
  default:
    return ctx.NewError("Error: http.ClientRequest." + key + " not setable")
  }
}

func (p *NodeJS_http_ClientRequest) GetClassValue() (*values.Class, error) {
  ctx := p.Context()
  return values.NewUnconstructableClass(NewNodeJS_http_ClientRequestPrototype(), ctx), nil
}
