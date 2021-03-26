package tree

import (
	"strings"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

type HTML struct {
	classes []string
	tagData
}

func NewHTML(attr *tokens.StringDict, ctx context.Context) (Tag, error) {
	td, err := newTag("html", false, attr, ctx)
	if err != nil {
		return nil, err
	}
	return &HTML{[]string{}, td}, nil
}

func (t *HTML) getHeadBody() (*Head, *Body, error) {
	var head *Head = nil
	var body *Body = nil

	for _, child := range t.children {
		switch tt := child.(type) {
		case *Head:
			if head != nil {
				errCtx := context.MergeContexts(child.Context(), head.Context())
				return nil, nil, errCtx.NewError("HTML Error: head defined twice")
			} else if body != nil {
				errCtx := context.MergeContexts(child.Context(), body.Context())
				return nil, nil, errCtx.NewError("HTML Error: body defined before head")
			}

			head = tt
		case *Body:
			if body != nil {
				errCtx := context.MergeContexts(child.Context(), body.Context())
				return nil, nil, errCtx.NewError("HTML Error: body defined twice")
			}

			body = tt
		default:
			errCtx := child.Context()
      err := errCtx.NewError("HTML Error: expected head or body (" + child.Name() + ")")
      context.AppendContextString(err, "Info: children of this tag", t.Context())
			return nil, nil, err
		}
	}

	if head == nil {
		return nil, nil, t.ctx.NewError("HTML Error: no head defined")
	}

	if body == nil {
		return nil, nil, t.ctx.NewError("HTML Error: no body defined")
	}

	return head, body, nil
}

func (t *HTML) Validate() error {
	head, body, err := t.getHeadBody()

	if err != nil {
		return err
	}

	if err := head.Validate(); err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	return err
}

func (t *HTML) CollectIDs(idMap IDMap) error {
	_, body, err := t.getHeadBody()
	if err != nil {
		return err
	}

	idMap.Set("html", t)

	return body.CollectIDs(idMap)
}

func (t *HTML) LinkStyle(cssUrl string) error {
	head, _, err := t.getHeadBody()
	if err != nil {
		return err
	}

  // if cssUrl == "" then css must be added by caller using the t.IncludeStyle() method
  if cssUrl != "" {
    ctx := t.Context()
    linkTag, err := NewStyleSheetLink(cssUrl, ctx)
    if err != nil {
      return err
    }

    // must be inserted before any other style tag, but not necessarily before any other link tag
    iInsert := -1
    for i, childTag := range head.Children() {
      if childTag.Name() == "style" {
        iInsert = i
      }
    }

    if iInsert < 0 {
      head.AppendChild(linkTag)
    } else {
      head.InsertChild(iInsert, linkTag)
    }
  }

  return nil
}

func (t *HTML) LinkScriptBundle(bundleURL string, fNames []string) error {
	head, body, err := t.getHeadBody()
	if err != nil {
		return err
	}

	ctx := head.Context()
	srcScript, _ := NewSrcScript(bundleURL, ctx)
	head.AppendChild(srcScript)

	// write the loader content
	var b strings.Builder
  for _, fName := range fNames {
    b.WriteString(fName)
    b.WriteString("();")
  }

	loaderContent := b.String()

	loaderScript, err := NewLoaderScript(loaderContent, ctx)
	if err != nil {
		return err
	}

	body.AppendChild(loaderScript)

	return nil
}

func (t *HTML) IncludeStyle(style string) error {
	head, _, err := t.getHeadBody()
  if err != nil {
    return err
  }

	ctx := head.Context()
  styleTag, err := NewStyle(tokens.NewEmptyStringDict(ctx), style, ctx)
  if err != nil {
    return err
  }

  head.AppendChild(styleTag)
  return nil
}

func (t *HTML) IncludeScript(code string) error {
	head, _, err := t.getHeadBody()
  if err != nil {
    return err
  }

	ctx := head.Context()
	script, _ := NewLoaderScript(code, ctx)
	head.AppendChild(script)

  return nil
}

func (t *HTML) GetClasses() []string {
	return t.classes
}

func (t *HTML) SetClasses(cs []string) {
	t.classes = cs
}

func (t *HTML) Write(indent string, nl, tab string) string {
	hasClasses := len(t.classes) > 0

	if hasClasses {
		keyToken := tokens.NewValueString("class", t.Context())
		valueToken := tokens.NewValueString(strings.Join(t.classes, " "), t.Context())
		t.attributes.Set(keyToken, valueToken)
	}

	result := t.tagData.Write(indent, nl, tab)

	if hasClasses {
		t.attributes.Delete("class")
	}

	return result
}

func (t *HTML) InnerHTML() string {
	return ""
}

func (t *HTML) ToJSType() string {
	return "HTMLElement"
}
