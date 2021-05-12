package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type File struct {
  BuiltinPrototype
}

func NewFilePrototype() values.Prototype {
  return &File{newBuiltinPrototype("File")}
}

func NewFile(ctx context.Context) values.Value {
  return values.NewInstance(NewFilePrototype(), ctx)
}

func (p *File) GetParent() (values.Prototype, error) {
  return NewBlobPrototype(), nil
}

func IsFile(v values.Value) bool {
  ctx := context.NewDummyContext()

  fileCheck := NewFile(ctx)

  return fileCheck.Check(v, ctx) == nil
}

func (p *File) Check(other_ values.Interface, ctx context.Context) error {
  if _, ok := other_.(*File); ok {
    return nil
  } else {
    return checkParent(p, other_, ctx)
  }
}

func (p *File) GetInstanceMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  i := NewInt(ctx)
  f := NewNumber(ctx)
  s := NewString(ctx)

  switch key {
  case "name":
    return s, nil
  case "lastModified":
    return f, nil
  case "size":
    return i, nil
  case "type":
    return s, nil
  default:
    return nil, nil
  }
}

func (p *File) GetClassValue() (*values.Class, error) {
  ctx := p.Context()
  return values.NewUnconstructableClass(NewFilePrototype(), ctx), nil
}
