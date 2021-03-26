package tree

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

type Hr struct {
	VisibleTagData
}

func NewHr(attr *tokens.StringDict, ctx context.Context) (Tag, error) {
	visTag, err := NewVisibleTag("hr", true, attr, ctx)
	return &Hr{visTag}, err
}

func (t *Hr) Validate() error {
	if t.NumChildren() != 0 {
		panic("should've been caught during construction")
	}

	return nil
}
