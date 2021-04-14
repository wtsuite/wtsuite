package js

import (
  "strings"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	"github.com/wtsuite/wtsuite/pkg/tokens/js/values"
)

type TypeAlias struct {
  lhs *VarExpression
  rhs *TypeExpression
  TokenData
}

func NewTypeAlias(lhs *VarExpression, rhs *TypeExpression, ctx context.Context) (*TypeAlias, error) {
  lhs.GetVariable().SetConstant()

  return &TypeAlias{lhs, rhs, newTokenData(ctx)}, nil
}

func (t *TypeAlias) Name() string {
  return t.lhs.Name()
}

func (t *TypeAlias) GetVariable() Variable {
  return t.lhs.GetVariable()
}

func (t *TypeAlias) Dump(indent string) string {
  var b strings.Builder

  b.WriteString(indent)
  b.WriteString("TypeAlias(")
  b.WriteString(t.Name())
  b.WriteString(")\n")
  b.WriteString(t.rhs.Dump(indent + "t "))

  return b.String()
}

func (t *TypeAlias) AddStatement(st Statement) {
	panic("not a block")
}

func (t *TypeAlias) WriteStatement(usage Usage, indent string, nl string, tab string) string {
  return ""
}

func (t *TypeAlias) HoistNames(scope Scope) error {
  return nil
}

func (t *TypeAlias) ResolveStatementNames(scope Scope) error {
  if err := t.rhs.ResolveExpressionNames(scope); err != nil {
    return err
  }

  if rhsVal, err := t.rhs.EvalExpression(); err != nil {
    return err
  } else {
    v := t.GetVariable()
    v.SetObject(values.NewTypeAlias(rhsVal))

    return scope.SetVariable(t.Name(), v)
  }

}

func (t *TypeAlias) EvalStatement() error {
  return nil
}

func (t *TypeAlias) ResolveStatementActivity(usage Usage) error {
  return t.rhs.ResolveExpressionActivity(usage)
}

func (t *TypeAlias) UniversalStatementNames(ns Namespace) error {
  return t.rhs.UniversalExpressionNames(ns)
  // dont need to check lhs because it will never appear in the final code
}

func (t *TypeAlias) UniqueStatementNames(ns Namespace) error {
  return t.rhs.UniqueExpressionNames(ns)

  // dont need to check lhs because it will never appear in the final code
}

func (t *TypeAlias) Walk(fn WalkFunc) error {
  if err := t.rhs.Walk(fn); err != nil {
    return err
  }

  if err := t.lhs.Walk(fn); err != nil {
    return err
  }

  return fn(t)
}
