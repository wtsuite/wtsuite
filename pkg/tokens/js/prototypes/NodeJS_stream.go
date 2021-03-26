package prototypes

import (
  "github.com/wtsuite/wtsuite/pkg/tokens/js/values"
)

func FillNodeJS_streamPackage(pkg values.Package) {
  pkg.AddPrototype(NewNodeJS_stream_ReadablePrototype())
}
