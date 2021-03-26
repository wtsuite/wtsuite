package tree

import (
	"strings"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type LoaderScript struct {
	content string
	LeafTag
}

func NewLoaderScript(content string, ctx context.Context) (*LoaderScript, error) {
	return &LoaderScript{content, NewLeafTag(ctx)}, nil
}

func (t *LoaderScript) Write(indent string, nl, tab string) string {
	var b strings.Builder

	b.WriteString(indent)
	b.WriteString("<script>")
	b.WriteString(nl)

  b.WriteString(indent + tab)
	b.WriteString("function onload(){")
	b.WriteString(nl)

  lines := strings.Split(t.content, "\n")
  for _, line := range lines {
    b.WriteString(indent + tab + tab)
    b.WriteString(line)
    b.WriteString(nl)
  }
  b.WriteString(indent + tab)
	b.WriteString("}")
	b.WriteString(nl)

  b.WriteString(indent + tab)
	b.WriteString("window.addEventListener(\"load\",onload,false);")
	b.WriteString(nl)

	b.WriteString(indent)
	b.WriteString("</script>")

	return b.String()
}
