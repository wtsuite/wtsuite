package directives

import (
	"github.com/wtsuite/wtsuite/pkg/functions"
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

const FILE = "__file__"

func SetFile(scope Scope, path string, ctx context.Context) {
	// set the __file__ internal variable immediately
  if err := scope.SetVar(FILE, functions.Var{
		tokens.NewValueString(path, ctx),
		true,
		false,
		false,
		ctx,
	}); err != nil {
    panic(err)
  }
}
