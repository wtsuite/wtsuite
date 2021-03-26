package values

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type Callable interface {
  EvalCall(args []Value, ctx context.Context) (Value, error)
}
