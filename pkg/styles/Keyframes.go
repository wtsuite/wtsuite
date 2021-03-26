package styles

import (
  "strings"

	tokens "github.com/wtsuite/wtsuite/pkg/tokens/html"
)

// TODO: uniqueness of keyframes name?
type Keyframe struct {
  pos string // XXX: should we do parsing of from/to/100% etc.?
  attr *tokens.StringDict
}

type KeyframesRule struct {
  id string
  kfs []*Keyframe
}

func NewKeyframesRule(sel Selector, key *tokens.String, attr *tokens.StringDict) ([]Rule, error) {
  ctx := key.Context()
  if sel != nil {
    return nil, ctx.NewError("Error: @keyframes must be top-level (can't bubble)")
  }

  args := strings.Fields(strings.TrimLeft(key.Value(), "@"))

  if len(args) != 2 {
    return nil, ctx.NewError("Error: expected @keyframes <id> {...}")
  }

  id := args[1]

  kfs := make([]*Keyframe, 0)

  if err := attr.Loop(func(k *tokens.String, v_ tokens.Token, last bool) error {
    v, err := tokens.AssertStringDict(v_)
    if err != nil {
      return err
    }

    kfs = append(kfs, &Keyframe{k.Value(), v})

    return nil
  }); err != nil {
    return nil, err
  }

  rule := &KeyframesRule{id, kfs}

  return []Rule{rule}, nil
}

func (r *KeyframesRule) ExpandNested() ([]Rule, error) {
  return []Rule{r}, nil
}

func (kf *Keyframe) write(indent string, nl string, tab string) (string, error) {
  var b strings.Builder

  b.WriteString(indent)
  b.WriteString(kf.pos)
  b.WriteString("{")
  b.WriteString(nl)
  inner, err := kf.attr.ToString(indent + tab, nl)
  if err != nil {
    return "", err
  }
  b.WriteString(inner)
  b.WriteString(indent)
  b.WriteString("}")
  b.WriteString(nl)

  return b.String(), nil
}

func (r *KeyframesRule) Write(indent string, nl string, tab string) (string, error) {
  var b strings.Builder

  b.WriteString(indent)
  b.WriteString("@keyframes ")
  b.WriteString(r.id)
  b.WriteString("{")
  b.WriteString(nl)
  for _, kf := range r.kfs {
    inner, err := kf.write(indent + tab, nl, tab)
    if err != nil {
      return "", err
    }
    b.WriteString(inner)
    b.WriteString(nl)
  }

  b.WriteString(indent)
  b.WriteString("}")
  b.WriteString(nl)

  return b.String(), nil
}
