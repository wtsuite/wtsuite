package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

func FillNodeJS_fsPackage(pkg values.Package) {
  ctx := context.NewDummyContext()
  b := NewBoolean(ctx)
  i := NewInt(ctx)
  s := NewString(ctx)
  buf := NewNodeJS_Buffer(ctx)

  pkg.AddValue("existsSync", values.NewFunction([]values.Value{s, b}, ctx))

  pkg.AddValue("readFileSync", values.NewOverloadedFunction([][]values.Value{
      []values.Value{s, buf},
      []values.Value{s, NewLiteralString("buffer", ctx), buf},
      []values.Value{s, s, s},
    }, ctx))

  pkg.AddValue("unlinkSync", values.NewOverloadedFunction([][]values.Value{
      []values.Value{s, nil},
      []values.Value{buf, nil},
    }, ctx))

  callback := values.NewFunction([]values.Value{NewError(ctx), nil}, ctx)
  opt := NewConfigObject(map[string]values.Value{
    "encoding": s,
    "mode": i,
    "flag": s,
  }, ctx)
  pkg.AddValue("writeFile", values.NewOverloadedFunction([][]values.Value{
    []values.Value{s, s, callback, nil},
    []values.Value{buf, s, callback, nil},
    []values.Value{i, s, callback, nil},
    []values.Value{s, buf, callback, nil},
    []values.Value{buf, buf, callback, nil},
    []values.Value{i, buf, callback, nil},
    []values.Value{s, s, s, callback, nil},
    []values.Value{s, s, opt, callback, nil},
    []values.Value{buf, s, opt, callback, nil},
    []values.Value{i, s, opt, callback, nil},
    []values.Value{s, buf, s, callback, nil},
    []values.Value{s, buf, opt, callback, nil},
    []values.Value{buf, buf, opt, callback, nil},
    []values.Value{i, buf, opt, callback, nil},
  }, ctx))
}
