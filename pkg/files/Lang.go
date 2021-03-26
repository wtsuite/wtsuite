package files

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type Lang string

const (
  SCRIPT Lang = "script"
  TEMPLATE = "template"
  SHADER = "shader"
)

type PathLang struct {
  Path string
  Lang Lang
  Context context.Context
}
