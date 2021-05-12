package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type FileList struct {
  BuiltinPrototype
}

func NewFileListPrototype() values.Prototype {
  return &FileList{newBuiltinPrototype("FileList")}
}

func NewFileList(ctx context.Context) values.Value {
  return values.NewInstance(NewFileListPrototype(), ctx)
}

func (p *FileList) Check(other_ values.Interface, ctx context.Context) error {
  if _, ok := other_.(*FileList); ok {
    return nil
  } else {
    return checkParent(p, other_, ctx)
  }
}

func (p *FileList) GetInstanceMember(key string, includePrivate bool, ctx context.Context) (values.Value, error) {
  i := NewInt(ctx)
  file := NewFile(ctx)

  switch key {
  case ".getindex", "item":
    return values.NewFunction([]values.Value{i, file}, ctx), nil
  case "length":
    return i, nil
  default:
    return nil, nil
  }
}

func (p *FileList) GetClassValue() (*values.Class, error) {
  ctx := p.Context()
  return values.NewUnconstructableClass(NewFileListPrototype(), ctx), nil
}
