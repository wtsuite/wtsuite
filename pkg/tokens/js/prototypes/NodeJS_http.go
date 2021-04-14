package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"

  "github.com/wtsuite/wtsuite/pkg/tokens/context"
)

func FillNodeJS_httpPackage(pkg values.Package) {
  pkg.AddPrototype(NewNodeJS_http_ClientRequestPrototype())
  pkg.AddPrototype(NewNodeJS_http_IncomingMessagePrototype())
  pkg.AddPrototype(NewNodeJS_http_ServerPrototype())
  pkg.AddPrototype(NewNodeJS_http_ServerResponsePrototype())

  ctx := context.NewDummyContext()
  s := NewString(ctx)
  obj := NewObject(nil, ctx)
  callback := values.NewFunction([]values.Value{
    NewNodeJS_http_IncomingMessage(ctx), nil}, ctx)

  pkg.AddValue("createServer", values.NewFunction([]values.Value{
    NewNodeJS_http_Server(ctx),
  }, ctx))

  pkg.AddValue("request", values.NewOverloadedFunction([][]values.Value{
    []values.Value{obj, callback, NewNodeJS_http_ClientRequest(ctx)},
    []values.Value{s, obj, callback, NewNodeJS_http_ClientRequest(ctx)},
  }, ctx))
}
