package macros

import (
  "strings"

	"github.com/computeportal/wtsuite/pkg/tokens/js"
	"github.com/computeportal/wtsuite/pkg/tokens/js/values"
	"github.com/computeportal/wtsuite/pkg/tokens/context"
)

// type is taken from rhs
type ConstIfUndefined struct {
  varNotFound bool // if true, write the statement (undefined defined variables eg. unused function variables, shouldn't be possible)
  Macro
}

func NewConstIfUndefined(args []js.Expression, ctx context.Context) (js.Statement, error) {
  if len(args) != 2 {
    return nil, ctx.NewError("Error: expected 2 arguments")
  }

  // arg1 
  if _, err := js.AssertVarExpression(args[0]); err != nil {
    return nil, err
  }

  return &ConstIfUndefined{false, newMacro(args, ctx)}, nil
}

func (m *ConstIfUndefined) Dump(indent string) string {
  return indent + "ConstIfUndefined(...)"
}

func (m *ConstIfUndefined) AddStatement(st js.Statement) {
  panic("not a block")
}

func (m *ConstIfUndefined) Name() string {
  ve, err := js.AssertVarExpression(m.args[0])
  if err != nil {
    panic(err)
  }

  return ve.Name()
}

func (m *ConstIfUndefined) GetVariable() js.Variable {
  ve, err := js.AssertVarExpression(m.args[0])
  if err != nil {
    panic(err)
  }

  return ve.GetVariable()
}

func (m *ConstIfUndefined) WriteStatement(usage js.Usage, indent string, nl string, tab string) string {
  var b strings.Builder

  if (m.varNotFound) {
    b.WriteString(indent)
    b.WriteString("const ")
    b.WriteString(m.Name())
    b.WriteString("=")
    b.WriteString(m.args[1].WriteExpression())
  } 

  return b.String()
}

func (m *ConstIfUndefined) HoistNames(scope js.Scope) error {
  return nil
}

func (m *ConstIfUndefined) ResolveStatementNames(scope js.Scope) error {
  name := m.Name()

  if !scope.HasVariable(name) {
    m.varNotFound = true

    variable := m.GetVariable()
    variable.SetConstant()

    if err := scope.SetVariable(name, variable); err != nil {
      panic(err)
      return err
    }
  } else {
    m.args[0].ResolveExpressionNames(scope)
  }

  return m.args[1].ResolveExpressionNames(scope)
}

func (m *ConstIfUndefined) EvalStatement() error {
  rhs, err := m.args[1].EvalExpression()
  if err != nil {
    return err
  }

  rhs = values.RemoveLiteralness(rhs)

  if !m.varNotFound {
    // make sure the rhs matches
    lhs, err := m.args[0].EvalExpression()
    if err != nil {
      return err
    }

    lhs = values.RemoveLiteralness(lhs)

    if err := lhs.Check(rhs, m.args[1].Context()); err != nil {
      return err
    }
  } else {
    variable := m.GetVariable()
    variable.SetValue(rhs)
  }

  return nil
}

func (m *ConstIfUndefined) ResolveStatementActivity(usage js.Usage) error {
  if err := m.args[1].ResolveExpressionActivity(usage); err != nil {
    return err
  }

  if m.varNotFound {
    if err := usage.Rereference(m.GetVariable(), m.Context()); err != nil {
      return err
    }
  }

  return nil
}

func (m *ConstIfUndefined) UniversalStatementNames(ns js.Namespace) error {
  return m.args[1].UniversalExpressionNames(ns)
}

func (m *ConstIfUndefined) UniqueStatementNames(ns js.Namespace) error {
  ns.LetName(m.GetVariable())

  return m.args[1].UniqueExpressionNames(ns)
}
