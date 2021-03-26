package functions

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

type LambdaScope interface {
	Eval(key string, args *tokens.Parens, ctx context.Context) (tokens.Token, error)
  Permissive() bool
	SetVar(name string, v Var) error
}

// must be registered by directives package
var NewLambdaScope func(fnScope tokens.Scope, callerScope tokens.Scope) LambdaScope
