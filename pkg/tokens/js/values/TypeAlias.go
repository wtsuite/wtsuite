package values

type TypeAlias struct {
  content Value
}

func NewTypeAlias(content Value) *TypeAlias {
  return &TypeAlias{content}
}

func (t *TypeAlias) Content() Value {
  return t.content
}
