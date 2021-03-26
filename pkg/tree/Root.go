package tree

import (
	"fmt"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	//"github.com/wtsuite/wtsuite/pkg/tree/scripts"
)

// doesn't need to implement Tag interface
// TODO: maybe it is more convenient if it DOES implement the Tag interface
type Root struct {
	tagData
}

func NewRoot(ctx context.Context) *Root {
	return &Root{tagData{"", "", false, nil, nil, make([]Tag, 0), ctx}}
}

func (t *Root) GetDocTypeAndHTML() (*DocType, *HTML, error) {
	var docType *DocType = nil
	var html *HTML = nil

	for _, child := range t.children {
		switch tt := child.(type) {
		case *DocType:
			if docType != nil {
				errCtx := context.MergeContexts(child.Context(), docType.Context())
				return nil, nil, errCtx.NewError("HTML Error: DOCTYPE defined twice")
			} else if html != nil {
				errCtx := context.MergeContexts(child.Context(), html.Context())
				return nil, nil, errCtx.NewError("HTML Error: html defined before DOCTYPE")
			}

			docType = tt
		case *HTML:
			if html != nil {
				errCtx := context.MergeContexts(child.Context(), html.Context())
				return nil, nil, errCtx.NewError("HTML Error: html defined twice")
			}

			html = tt
		default:
			errCtx := child.Context()
			return nil, nil, errCtx.NewError("HTML Error: expected DOCTYPE or html")
		}
	}

	if docType == nil {
		if AUTO_DOC_TYPE {
			docType = NewAutoDocType()
			t.children = []Tag{docType, html}
		} else {
			err := t.ctx.NewError(fmt.Sprintf("HTML Error: no !DOCTYPE defined (nChildren: %d)",
				len(t.children)))
			return nil, nil, err
		}
	}

	if html == nil {
		return nil, nil, t.ctx.NewError("HTML Error: no html defined")
	}

	return docType, html, nil
}

func (t *Root) Validate() error {
	docType, html, err := t.GetDocTypeAndHTML()
	if err != nil {
		return err
	}

	if err := docType.Validate(); err != nil {
		return err
	}

	if err := html.Validate(); err != nil {
		return err
	}

	return err
}

func (t *Root) VerifyElementCount(i int, ecKey string) error {
	for i, child := range t.children {
		if err := child.VerifyElementCount(i, ecKey); err != nil {
			return err
		}
	}

	return nil
}

// dummy args are just for interface
func (t *Root) Write(indent string, nl string, tab string) string {
	return t.writeChildren(indent, nl, tab)
}

func (t *Root) CollectIDs(idMap IDMap) error {
	_, html, err := t.GetDocTypeAndHTML()
	if err != nil {
		return err
	}

	return html.CollectIDs(idMap)
}

func (t *Root) LinkStyle(cssUrl string) error {
	_, html, err := t.GetDocTypeAndHTML()
	if err != nil {
		return err
	}

	return html.LinkStyle(cssUrl)
}

func (t *Root) LinkScriptBundle(bundleURL string, fNames []string) error {
	_, html, err := t.GetDocTypeAndHTML()
	if err != nil {
		return err
	}

	return html.LinkScriptBundle(bundleURL, fNames)
}

func (t *Root) IncludeStyle(styles string) error {
	_, html, err := t.GetDocTypeAndHTML()
	if err != nil {
		return err
	}

	return html.IncludeStyle(styles)
}

func (t *Root) IncludeScript(code string) error {
	_, html, err := t.GetDocTypeAndHTML()
	if err != nil {
		return err
	}

	return html.IncludeScript(code)
}
