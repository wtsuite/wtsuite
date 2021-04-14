package parsers

import (
  "strings"

	"github.com/wtsuite/wtsuite/pkg/tokens/context"
	"github.com/wtsuite/wtsuite/pkg/tokens/html"
	"github.com/wtsuite/wtsuite/pkg/tokens/patterns"
	"github.com/wtsuite/wtsuite/pkg/tokens/raw"
)

var templateFunctionMap = map[string]string{
  "pre$":   "get",
	"pre-":   "neg",
	"pre!":   "not",
	"bin/":   "div",
	"bin*":   "mul",
	"bin-":   "sub",
	"bin+":   "add",
	"bin<":   "lt",
	"bin<=":  "le",
	"bin>":   "gt",
	"bin>=":  "ge",
	"bin!=":  "ne",
	"bin==":  "eq",
	"bin===": "issame",
	"bin||":  "or",
	"bin&&":  "and",
	// ":", "?", ":=" and "=" are treated explicitely
}

func (p *TemplateParser) buildEndOfLineExpression(ts []raw.Token) (html.Token, []raw.Token, error) {
  if len(ts) == 0 {
    panic("no expression tokens")
  }

  isExpectsMoreOp := func(t raw.Token) bool {
    if raw.IsAnySymbol(t) {
      s, err := raw.AssertAnySymbol(t)
      if err != nil {
        panic(err)
      }

      v := s.Value()

      // only operators, so can't use symbols pattern

      if v == ":" || v == "$" || v == "?" || v == "<" || v == ">" || v == "+" || v == "-" || v == "/" || v == "*" || v == "!=" || v == "==" || v == ":=" || v == "<=" || v == ">=" || v == "!" || v == "===" || v == "&&" || v == "||" || v == "!!" || v == "??" {
        return true
      } else {
        return false
      }
    } else {
      return false
    }
  }

  // find the end
  iStop := 0
  expectsMore := true
  groupCount := 0
  for i, t := range ts {
    if raw.IsSymbol(t, patterns.BRACES_START) || raw.IsSymbol(t, patterns.PARENS_START) || raw.IsSymbol(t, patterns.BRACKETS_START) {
      groupCount += 1
      expectsMore = true
    } else if raw.IsSymbol(t, patterns.BRACES_STOP) || raw.IsSymbol(t, patterns.PARENS_STOP) || raw.IsSymbol(t, patterns.BRACKETS_STOP) {
      if groupCount == 0 {
        errCtx := t.Context()
        return nil, nil, errCtx.NewError("Error: unmatched closing tag")
      }

      groupCount -= 1

      if groupCount == 0 {
        expectsMore = false
      }
    } else if isExpectsMoreOp(t) {
      expectsMore = true
    } else if raw.IsNL(t) && !expectsMore {
      iStop = i
      break
    } else if groupCount == 0 && !raw.IsWhitespace(t) {
      expectsMore = false
    }

    iStop = i+1
  }

  resExpr, err := p.nestAndBuildExpression(ts[0:iStop])

  return resExpr, ts[iStop:], err
}

func (p *TemplateParser) buildTextTagExpression(ts []raw.Token) (html.Token, []raw.Token, error) {
  if len(ts) == 0 {
    panic("no expression tokens")
  }

  isExpectsMoreOp := func(t raw.Token) bool {
    if raw.IsAnySymbol(t) {
      s, err := raw.AssertAnySymbol(t)
      if err != nil {
        panic(err)
      }

      v := s.Value()

      // only operators, so can't use symbols pattern
      if v == ":" || v == "$" || v == "?" || v == "+" || v == "!=" || v == "==" || v == "!" || v == "===" || v == "&&" || v == "||" || v == "!!" || v == "??" {
        return true
      } else {
        return false
      }
    } else {
      return false
    }
  }

  // find the end
  iStop := 0
  expectsMore := true
  groupCount := 0
  for i, t := range ts {
    if raw.IsSymbol(t, patterns.BRACES_START) || raw.IsSymbol(t, patterns.PARENS_START) || raw.IsSymbol(t, patterns.BRACKETS_START) {
      groupCount += 1
      expectsMore = true
    } else if raw.IsSymbol(t, patterns.BRACES_STOP) || raw.IsSymbol(t, patterns.PARENS_STOP) || raw.IsSymbol(t, patterns.BRACKETS_STOP) {
      if groupCount == 0 {
        errCtx := t.Context()
        return nil, nil, errCtx.NewError("Error: unmatched closing tag")
      }

      groupCount -= 1

      if groupCount == 0 {
        expectsMore = false
      }
    } else if isExpectsMoreOp(t) {
      expectsMore = true
    } else if (raw.IsNL(t) || raw.IsAnyWord(t) || raw.IsLiteralString(t) || raw.IsSymbol(t, "<")) && !expectsMore  {
      iStop = i
      break
    } else if (raw.IsAnyWord(t) || raw.IsLiteralString(t) || raw.IsSymbol(t, "<")) && groupCount == 0 {
      expectsMore = false
    }

    iStop = i+1
  }

  resExpr, err := p.nestAndBuildExpression(ts[0:iStop])

  return resExpr, ts[iStop:], err
}

func (p *TemplateParser) nestAndBuildExpression(ts []raw.Token) (html.Token, error) {
  tsInner := raw.RemoveWhitespace(ts)

  var err error
  tsInner, err = p.nestGroups(tsInner)
	if err != nil {
		return nil, err
	}

	tsInner, err = p.nestOperators(tsInner)
	if err != nil {
		return nil, err
	}

  return p.buildExpression(tsInner)
}

func (p *TemplateParser) buildExpression(vs []raw.Token) (html.Token, error) {
  vs = p.expandTmpGroups(vs)
	/*if len(vs) == 1 {
		if tmp, ok := vs[0].(*raw.Group); ok {
			if tmp.IsTmp() {
				vs = tmp.Fields[0]
			}
		}
	}*/

	switch len(vs) {
	case 0:
		panic("expected at least one token")
	case 1:
		switch v := vs[0].(type) {
		case *raw.LiteralBool:
			return html.NewValueBool(v.Value(), v.Context()), nil
		case *raw.LiteralColor:
			r, g, b, a := v.Values()
			return html.NewValueColor(r, g, b, a, v.Context()), nil
		case *raw.LiteralFloat:
			return html.NewValueUnitFloat(v.Value(), v.Unit(), v.Context()), nil
		case *raw.LiteralInt:
			return html.NewValueInt(v.Value(), v.Context()), nil
		case *raw.LiteralNull:
			return html.NewNull(v.Context()), nil
		case *raw.LiteralString:
      return html.NewValueString(v.Value(), v.Context()), nil
		case *raw.Word:
			//return html.NewFunction("get", []html.Token{html.NewValueString(v.Value(), v.Context())},
				//v.Context()), nil
      return html.NewValueString(v.Value(), v.Context()), nil
		case *raw.Operator:
			// NOTE: raw.Operator tokens are generated by FormulaParser
			// bin= is a special case, and the bin= function call is just used a placeholder
			return p.buildOperatorExpression(v)
		case *raw.Group:
			if v.IsTmp() {
				return p.buildExpression(v.Fields[0])
			} else {
				return p.buildGroupExpression(v)
			}
		default:
			errCtx := v.Context()
      if raw.IsWhitespace(v) {
        return nil, errCtx.NewError("Internal Error: whitespace not filtered out")
      }

      err := errCtx.NewError("Error: invalid syntax")
      panic(err)
			return nil, err
		}
	default:
    /*if len(vs) >= 3 && raw.ContainsSymbol(vs[1:len(vs)-1], patterns.ARROW) {
      fn, remaining, err := p.buildArrowFunctionExpression(vs)
      if err != nil {
        return nil, err
      }

      if len(remaining) == 0 {
        return fn, nil
      } else {
        return p.buildEvalsAndIndexing(fn, remaining)
      }
    } else */
    if len(vs) >= 3 && raw.IsAnyWord(vs[0]) && raw.IsParensGroup(vs[1]) && raw.IsBracesGroup(vs[2]) {
			fn, remaining, err := p.buildDefineFunctionExpression(vs)
			if err != nil {
				return nil, err
			}

			if len(remaining) == 0 {
				return fn, nil
			} else {
				return p.buildEvalsAndIndexing(fn, remaining)
			}
		} else if len(vs) >= 2 && raw.IsAnyWord(vs[0]) && raw.IsParensGroup(vs[1]) {
			fn, remaining, err := p.buildFunctionExpression(vs)
			if err != nil {
				return nil, err
			}

			if len(remaining) == 0 {
				return fn, nil
			} else {
				return p.buildEvalsAndIndexing(fn, remaining)
			}
		} else if len(vs) >= 2 && raw.IsAnyWord(vs[0]) && raw.IsBracketsGroup(vs[1]) {
			obj, remaining, err := p.buildIndexedExpression(vs)
			if err != nil {
				return nil, err
			}

			if len(remaining) == 0 {
				return obj, nil
			} else {
				return p.buildEvalsAndIndexing(obj, remaining)
			}
		} else {
			obj, err := p.buildExpression(vs[0:1])
			if err != nil {
				return nil, err
			}

			remaining := vs[1:]
			if len(remaining) == 0 {
				return obj, nil
			} else {
				return p.buildEvalsAndIndexing(obj, remaining)
			}
		}
	}
}

func (p *TemplateParser) buildOperatorExpression(v *raw.Operator) (html.Token, error) {
	switch {
  case v.Name() == "bin=>":
    // pipe operator is just syntactic suger for calls
		ab, err := raw.AssertBinaryOperator(v, "bin=>")
		if err != nil {
			return nil, err
		}

    argParens, err := p.wrapFunctionParens(ab.Args()[0:1])
    if err != nil {
      return nil, err
    }

    bodyBraces, remaining, err := p.wrapFunctionBraces(ab.Args()[1:2])
    if err != nil {
      return nil, err
    }

    if len(remaining) != 0 {
      errCtx := raw.MergeContexts(remaining...)
      return nil, errCtx.NewError("Error: unexpected")
    }

    return p.buildDefineFunctionExpressionInternal(argParens, bodyBraces)
  case v.Name() == "bin|":
    // pipe operator is just syntactic suger for calls
		ab, err := raw.AssertBinaryOperator(v, "bin|")
		if err != nil {
			return nil, err
		}

		a, err := p.buildExpression(ab.Args()[0:1])
		if err != nil {
			return nil, err
		}

		b, err := p.buildExpression(ab.Args()[1:2])
		if err != nil {
			return nil, err
		}

		return html.NewFunction("eval", []html.Token{b, html.NewValuesList([]html.Token{a}, a.Context())}, v.Context()), nil
	case v.Name() == "bin:" && raw.IsBinaryOperator(v.Args()[0], "bin?"): // actually a ternary operator

		ab, err := raw.AssertBinaryOperator(v.Args()[0], "bin?")
		if err != nil {
			return nil, err
		}

		a, err := p.buildExpression(ab.Args()[0:1])
		if err != nil {
			return nil, err
		}

		b, err := p.buildExpression(ab.Args()[1:2])
		if err != nil {
			return nil, err
		}

		c, err := p.buildExpression(v.Args()[1:2])
		if err != nil {
			return nil, err
		}

		return html.NewFunction("ifelse", []html.Token{a, b, c}, v.Context()), nil
  case v.Name() == "bin!!":
    // if-not-null-then operator
		a, err := p.buildExpression(v.Args()[0:1])
		if err != nil {
			return nil, err
		}

		b, err := p.buildExpression(v.Args()[1:2])
		if err != nil {
			return nil, err
		}

    cond := html.NewFunction("ne", []html.Token{a, html.NewNull(v.Context())}, v.Context())

    return html.NewFunction("ifelse", []html.Token{cond, b, a}, v.Context()), nil
  case v.Name() == "bin??":
    // nullish coalescing operator
		a, err := p.buildExpression(v.Args()[0:1])
		if err != nil {
			return nil, err
		}

		b, err := p.buildExpression(v.Args()[1:2])
		if err != nil {
			return nil, err
		}

    cond := html.NewFunction("eq", []html.Token{a, html.NewNull(v.Context())}, v.Context())

    return html.NewFunction("ifelse", []html.Token{cond, b, a}, v.Context()), nil
	case v.Name() == "bin=": // lhs must be word
		// accept word or string
    lhsTokens := p.expandTmpGroups(v.Args()[0:1])

    lhsFields := raw.SplitBySymbol(lhsTokens, patterns.COMMA)
    if len(lhsFields) == 0 {
      errCtx := v.Context()
      return nil, errCtx.NewError("Error: bad lhs")
    }

    nameTokens := make([]html.Token, 0)
    for _, lhsField := range lhsFields {
      if len(lhsField) == 0 {
        errCtx := v.Context()
        return nil, errCtx.NewError("Error: bad lhs")
      } else if len(lhsField) != 1 {
        errCtx := raw.MergeContexts(lhsField...)
        return nil, errCtx.NewError("Error: bad lhs")
      }

      lhsWord, err := raw.AssertWord(lhsField[0])
      if err != nil {
        return nil, err
      }

      nameTokens = append(nameTokens, html.NewValueString(lhsWord.Value(), lhsWord.Context()))
    }

    b, err := p.buildExpression(v.Args()[1:2])
    if err != nil {
      return nil, err
    }

    argTokens := append(nameTokens, b)
    return html.NewFunction("new", argTokens, v.Context()), nil
	case strings.HasPrefix(v.Name(), "bin"):
		a, err := p.buildExpression(v.Args()[0:1])
		if err != nil {
			return nil, err
		}
		b, err := p.buildExpression(v.Args()[1:2])
		if err != nil {
			return nil, err
		}
		if fnName, ok := templateFunctionMap[v.Name()]; ok {
			return html.NewFunction(fnName, []html.Token{a, b}, v.Context()), nil
		} else {
			errCtx := v.Context()
			err := errCtx.NewError("Error: binary operator '" + strings.TrimLeft(v.Name(), "bin") + "' not recognized")
			return nil, err
		}
  case v.Name() == "pre$":
    args := p.expandTmpGroups(v.Args())

    if len(args) == 1 {
      arg0, err := raw.AssertWord(args[0])
      if err != nil {
        return nil, err
      }

      return html.NewFunction("get", []html.Token{html.NewValueString(arg0.Value(), arg0.Context())}, v.Context()), nil
    } else {
      return p.buildExpression(args)
    }
	case strings.HasPrefix(v.Name(), "pre"):
		a, err := p.buildExpression(v.Args())
		if err != nil {
			return nil, err
		}
		if fnName, ok := templateFunctionMap[v.Name()]; ok {
			return html.NewFunction(fnName, []html.Token{a}, v.Context()), nil
		} else {
			errCtx := v.Context()
			return nil, errCtx.NewError("Error: pre unary operator '" + strings.TrimLeft(v.Name(), "pre") + "' not recognized")
		}
	case strings.HasPrefix(v.Name(), "post"):
		a, err := p.buildExpression(v.Args())
		if err != nil {
			return nil, err
		}
		if fnName, ok := templateFunctionMap[v.Name()]; ok {
			return html.NewFunction(fnName, []html.Token{a}, v.Context()), nil
		} else {
			errCtx := v.Context()
			return nil, errCtx.NewError("Error: post unary operator '" + strings.TrimLeft(v.Name(), "post") + "' not recognized")
		}
	case strings.HasPrefix(v.Name(), "sing"):
		if fnName, ok := templateFunctionMap[v.Name()]; ok {
			return html.NewFunction(fnName, []html.Token{}, v.Context()), nil
		} else {
			errCtx := v.Context()
			return nil, errCtx.NewError("Error: singular operator '" + strings.TrimLeft(v.Name(), "sing") + "' not recognized")
		}
	default:
		errCtx := v.Context()
		return nil, errCtx.NewError("Error: unrecognized operator '" + v.Name() + "'")
	}
}

func (p *TemplateParser) buildParensGroupExpression(v *raw.Group) (*html.Parens, error) {
	if v.IsParens() && (v.IsEmpty() || v.IsSingle() || v.IsComma()) {
		values := make([]html.Token, 0)
		alts := make([]html.Token, 0) // if first token is string, and second is '=', then remainder is alt, otherwise nil
		for _, field := range v.Fields {
      // TODO: 
			if raw.IsBinaryOperator(field[0], "bin=") {
				eq, err := raw.AssertBinaryOperator(field[0], "bin=")
				if err != nil {
					panic(err)
				}

				a, err := p.buildExpression(eq.Args()[0:1])
				if err != nil {
					return nil, err
				}

				b, err := p.buildExpression(eq.Args()[1:2])
				if err != nil {
					return nil, err
				}

				values = append(values, a)
				alts = append(alts, b)
			} else {
				val, err := p.buildExpression(field)
				if err != nil {
					return nil, err
				}

				values = append(values, val)
				alts = append(alts, nil)
			}
		}

    ctx := v.Context()
    ctx = context.SimpleFill(ctx, ctx)
		return html.NewParens(values, alts, ctx), nil
	} else {
		errCtx := v.Context()
		return nil, errCtx.NewError("Error: bad parens")
	}
}

func (p *TemplateParser) buildBracesGroupExpression(v *raw.Group) (*html.RawDict, error) {
	if v.IsBraces() && (v.IsEmpty() || v.IsSingle() || v.IsComma()) {
		keys := make([]html.Token, 0)
		values := make([]html.Token, 0)

		for _, field := range v.Fields {
			if len(field) != 1 {
        if len(field) > 1 {
          errCtx := raw.MergeContexts(field...)
          return nil, errCtx.NewError("Error: bad dict content")
        } else {
          errCtx := v.Context()
          return nil, errCtx.NewError("Error: bad dict content")
        }
			}
			colon, err := raw.AssertBinaryOperator(field[0], "bin:")
			if err != nil {
				return nil, err
			}

			a, err := p.buildExpression(colon.Args()[0:1])
			if err != nil {
				return nil, err
			}

			b, err := p.buildExpression(colon.Args()[1:2])
			if err != nil {
				return nil, err
			}

			keys = append(keys, a)
			values = append(values, b)
		}

		return html.NewValuesRawDict(keys, values, v.Context()), nil
	} else {
		errCtx := v.Context()
		return nil, errCtx.NewError("Error: bad braces")
	}
}

// return value can be List or seq(...) function call
func (p *TemplateParser) buildBracketsGroupExpression(v *raw.Group) (html.Token, error) {
	if v.IsBrackets() && (v.IsEmpty() || v.IsSingle() || v.IsComma()) {
		if v.IsSingle() && len(v.Fields[0]) == 1 && raw.IsOperator(v.Fields[0][0], "bin:") {
			op, err := raw.AssertAnyOperator(v.Fields[0][0])
			if err != nil {
				panic(err)
			}

			start, err := p.buildExpression(op.Args()[0:1])
			if err != nil {
				return nil, err
			}

			if raw.IsOperator(op.Args()[1], "bin:") {
				op2, err := raw.AssertAnyOperator(op.Args()[1])
				if err != nil {
					panic(err)
				}

				incr, err := p.buildExpression(op2.Args()[0:1])
				if err != nil {
					return nil, err
				}

				stop, err := p.buildExpression(op2.Args()[1:])
				if err != nil {
					return nil, err
				}

				return html.NewFunction("seq", []html.Token{start, incr, stop},
					context.MergeContexts(v.Context(), op.Context(), op2.Context())), nil
			} else {
				errCtx := v.Context()
				return nil, errCtx.NewError("Error: forming sequence like this is not allowed, because it is too easily confused with ','")
				/*stop, err := p.buildExpression(op.Args()[1:])
				if err != nil {
					return nil, err
				}

				return html.NewFunction("seq", []html.Token{start, stop},
					context.MergeContexts(v.Context(), op.Context())), nil*/
			}
		} else {
			values := make([]html.Token, 0)

			for _, field := range v.Fields {
				a, err := p.buildExpression(field)
				if err != nil {
					return nil, err
				}

				values = append(values, a)
			}

			return html.NewValuesList(values, v.Context()), nil
		}
	} else {
		errCtx := v.Context()
		return nil, errCtx.NewError("Error: bad brackets")
	}
}

func (p *TemplateParser) buildGroupExpression(v *raw.Group) (html.Token, error) {
	switch {
	case v.IsParens():
		return p.buildParensGroupExpression(v)
	case v.IsBraces():
		return p.buildBracesGroupExpression(v)
	case v.IsBrackets():
		return p.buildBracketsGroupExpression(v)
	default:
		errCtx := v.Context()
		return nil, errCtx.NewError("Error: unhandled group type")
	}
}

func (p *TemplateParser) wrapFunctionParens(vs []raw.Token) (*raw.Group, error) {
  vs = p.expandTmpGroups(vs)

  var err error
  var argParens *raw.Group 

  if len(vs) != 1 || !raw.IsParensGroup(vs[0]) {
    ctx := raw.MergeContexts(vs...)
    if raw.ContainsSymbol(vs, patterns.COMMA) {
      errCtx := ctx
      return nil, errCtx.NewError("Error: wrap with parenthesis")
    }

    first := raw.NewSymbol(patterns.PARENS_START, false, ctx)
    last := raw.NewSymbol(patterns.PARENS_STOP, false, ctx)
    all := append([]raw.Token{first}, vs...)
    all = append(all, last)

    argParens, err = raw.NewGroupFromTokens(all)
    if err != nil {
      return nil, err
    }
  } else {
    argParens, err = raw.AssertParensGroup(vs[0])
    if err != nil {
      panic(err)
    }
  }

  return argParens, nil
}

func (p *TemplateParser) wrapFunctionBraces(vs []raw.Token) (*raw.Group, []raw.Token, error) {
  vs = p.expandTmpGroups(vs)

  remaining := []raw.Token{}

  var err error
  var bodyBraces *raw.Group

  if raw.IsBracesGroup(vs[0]) {
    bodyBraces, err = raw.AssertBracesGroup(vs[0])
    if err != nil {
      panic(err)
    }

    remaining = vs[1:]
  } else {
    ctx := raw.MergeContexts(vs...)
    if raw.ContainsSymbol(vs, patterns.SEMICOLON) || raw.ContainsSymbol(vs, patterns.COMMA) {
      errCtx := ctx
      return nil, nil, errCtx.NewError("Error: must be wrapped with braces")
    }

    first := raw.NewSymbol(patterns.BRACES_START, false, ctx)
    last := raw.NewSymbol(patterns.BRACES_STOP, false, ctx)
    all := append([]raw.Token{first}, vs...)
    all = append(all, last)

    bodyBraces, err = raw.NewGroupFromTokens(all)
    if err != nil {
      return nil, nil, err
    }
  }

  return bodyBraces, remaining, nil
}

func (p *TemplateParser) buildArrowFunctionExpression(vs []raw.Token) (html.Token, []raw.Token, error) {
  // XXX: or use SplitByFirstSymbol
  fields := raw.SplitBySymbol(vs, patterns.ARROW)

  if len(fields) != 2 {
    errCtx := raw.MergeContexts(vs...)
    return nil, nil, errCtx.NewError("Error: bad arrow function")
  }

  argParens, err := p.wrapFunctionParens(fields[0])
  if err != nil {
    return nil, nil, err
  }

  bodyBraces, remaining, err := p.wrapFunctionBraces(fields[1])
  if err != nil {
    return nil, nil, err
  }

  fn, err := p.buildDefineFunctionExpressionInternal(argParens, bodyBraces)
  if err != nil {
    return nil, nil, err
  }

  return fn, remaining, nil
}

func (p *TemplateParser) buildDefineFunctionExpression(vs []raw.Token) (html.Token, []raw.Token, error) {
	// new function
	a, err := raw.AssertWord(vs[0])
	if err != nil {
		panic("unexpected")
	}

	if a.Value() != "function" {
		errCtx := a.Context()
		return nil, nil, errCtx.NewError("Error: expected function keyword")
	}

	argsGroup, err := raw.AssertParensGroup(vs[1])
	if err != nil {
		panic("unexpected")
	}

	statementsGroup, err := raw.AssertBracesGroup(vs[2])
	if err != nil {
		return nil, nil, err
	}

  fn, err := p.buildDefineFunctionExpressionInternal(argsGroup, statementsGroup)
  if err != nil {
    return nil, nil, err
  }

  return fn, vs[3:], nil
}

func (p *TemplateParser) buildDefineFunctionExpressionInternal(argsGroup *raw.Group, statementsGroup *raw.Group) (html.Token, error) {
	argsWithDefaults, err := p.buildParensGroupExpression(argsGroup)
	if err != nil {
		return nil, err
	}

  statements := make([]html.Token, 0)

  addStatement := func(ts []raw.Token, last bool) error {
    ctx := raw.MergeContexts(ts...)

    if last && raw.ContainsSymbol(ts, patterns.COMMA) {
      returnFields := raw.SplitBySymbol(ts, patterns.COMMA)
      returnExprs := make([]html.Token, 0)
      for _, returnField := range returnFields {
        if len(returnField) == 0 {
          errCtx := ctx
          return errCtx.NewError("Error: bad return field")
        }

        expr, err := p.buildExpression(returnField)
        if err != nil {
          return err
        }

        returnExprs = append(returnExprs, expr)
      }

      st := html.NewFunction("spread", []html.Token{html.NewValuesList(returnExprs, ctx)}, ctx)
      statements = append(statements, st)
    } else {
      st, err := p.buildExpression(ts)
      if err != nil {
        return err
      }

      statements = append(statements, st)
    }

    return nil
  }

	ctx := statementsGroup.Context()
  if (statementsGroup.IsComma()) {
    field, err := statementsGroup.FlattenCommas()
    if err != nil {
      return nil, err
    }

    if err := addStatement(p.expandTmpGroups(field), true); err != nil {
      return nil, err
    }
  } else if !(statementsGroup.IsSingle() || statementsGroup.IsSemiColon()) {
		errCtx := ctx
		return nil, errCtx.NewError("Error: bad statements for function def")
	} else {
    for i, field := range statementsGroup.Fields {
      isLast := i == len(statementsGroup.Fields)-1
      if err := addStatement(p.expandTmpGroups(field), isLast); err != nil {
        panic(err)
        return nil, err
      }
    }
  }

	// wrap statements in a get
	list := html.NewValuesList(statements, ctx)
	index := html.NewValueInt(len(statements)-1, ctx)
	wrapper := html.NewFunction("get", []html.Token{list, index}, ctx)

	return html.NewFunction("function", []html.Token{argsWithDefaults, wrapper}, ctx), nil
}

// also return the remaining
func (p *TemplateParser) buildFunctionExpression(vs []raw.Token) (html.Token, []raw.Token, error) {
	// new function
	a, err := raw.AssertWord(vs[0])
	if err != nil {
		panic("unexpected")
	}

	argsGroup, err := raw.AssertParensGroup(vs[1])
	if err != nil {
		panic("unexpected")
	}

	if !(argsGroup.IsEmpty() || argsGroup.IsSingle() || argsGroup.IsComma()) {
		errCtx := vs[1].Context()
		return nil, nil, errCtx.NewError("Error: bad function args")
	}

	args_, err := p.buildGroupExpression(argsGroup)
	if err != nil {
		return nil, nil, err
	}

	args, err := html.AssertParens(args_)
	if err != nil {
		panic(err)
	}

	return html.NewValueFunction(a.Value(), args,
		context.MergeContexts(a.Context(), vs[1].Context())), vs[2:], nil
}

func (p *TemplateParser) buildIndexedExpression(vs []raw.Token) (html.Token, []raw.Token, error) {
	indices := make([]html.Token, 0)

	varName, err := raw.AssertWord(vs[0])
	if err != nil {
		return nil, nil, err
	}

	obj := html.NewValueString(varName.Value(), varName.Context())

	ctx := vs[0].Context()
	for _, v := range vs[1:] {
		if raw.IsBracketsGroup(v) {
			indexGroup, err := raw.AssertBracketsGroup(v)
			if err != nil {
				return nil, nil, err
			}

			if !indexGroup.IsSingle() {
				errCtx := indexGroup.Context()
				return nil, nil, errCtx.NewError("Error: bad index (hint: multi indexing not supported)")
			}

			field := indexGroup.Fields[0]
			if len(field) == 1 && (raw.IsOperator(field[0], "sing:") || raw.IsOperator(field[0], "pre:") || raw.IsOperator(field[0], "post:") || raw.IsOperator(field[0], "bin:")) {
				break // these require a slice instead
			}

			index, err := p.buildExpression(field)
			if err != nil {
				return nil, nil, err
			}

			indices = append(indices, index)
		} else {
			break
		}
	}

	// nest these, so have get(get(get(dictname), "key"), index) etc.
	res := html.NewFunction("get", []html.Token{obj}, ctx)
	for _, index := range indices {
		res = html.NewFunction("get", []html.Token{res, index}, index.Context())
	}

	return res, vs[(len(indices) + 1):], nil
}

func (p *TemplateParser) buildEvalsAndIndexing(obj html.Token, vs []raw.Token) (html.Token, error) {
  prevPeriod := false

	for _, v := range vs {
		if !raw.IsGroup(v) {
      if prevPeriod {
        if raw.IsSymbol(v, ".") {
          errCtx := v.Context()
          return nil, errCtx.NewError("Error: can't have two periods in a raw")
        }

        w, err := raw.AssertWord(v)
        if err != nil {
          return nil, err
        }

        wParts := strings.Split(w.Value(), ".")
        wContexts := context.SplitByPeriod(w.Context(), len(wParts))

        for i, wPart := range wParts {
          if !patterns.IsValidVar(wPart) {
            errCtx := v.Context()
            err := errCtx.NewError("Error: invalid syntax")
            panic(err)
            return nil, err
          }

          obj = html.NewFunction("get", []html.Token{
            obj,
            html.NewValueString(wPart, wContexts[i]),
          }, wContexts[i])
        }

        prevPeriod = false
        continue
      } else {
        if raw.IsSymbol(v, ".") {
          prevPeriod = true
          continue
        } 
      }

			errCtx := v.Context()
      err := errCtx.NewError("Error: unexpected (expected group)")
			return nil, err
		}

		gr, err := raw.AssertGroup(v)
		if err != nil {
			panic(err)
		}

		switch {
		case gr.IsBrackets() && gr.IsSingle():
			field := gr.Fields[0]

			if len(field) == 1 && (raw.IsOperator(field[0], "sing:") ||
				raw.IsOperator(field[0], "post:") ||
				raw.IsOperator(field[0], "pre:") ||
				raw.IsOperator(field[0], "bin:")) {
				op, err := raw.AssertAnyOperator(field[0])
				if err != nil {
					panic(err)
				}

				switch {
				case op.Name() == "sing:":
					ctx := op.Context()
					obj = html.NewFunction("slice", []html.Token{obj, html.NewNull(ctx),
						html.NewValueInt(1, ctx), html.NewNull(ctx)}, ctx)
				case op.Name() == "post:":
					ctx := op.Context()
					a, err := p.buildExpression(op.Args())
					if err != nil {
						return nil, err
					}

					obj = html.NewFunction("slice", []html.Token{obj, a,
						html.NewValueInt(1, ctx), html.NewNull(ctx)}, ctx)
				case op.Name() == "pre:" || op.Name() == "bin:":
					var start html.Token
					if op.Name() == "pre:" {
						start = html.NewNull(op.Context())
					} else {
						start, err = p.buildExpression(op.Args()[0:1])
						if err != nil {
							return nil, err
						}
					}

					op2_ := op.Args()[0]
					if raw.IsOperator(op2_, "post:") || raw.IsOperator(op2_, "bin:") {
						op2, err := raw.AssertAnyOperator(op2_)
						if err != nil {
							panic(err)
						}

						switch {
						case op2.Name() == "post:":
							incr, err := p.buildExpression(op2.Args())
							if err != nil {
								return nil, err
							}

							stop := html.NewNull(op2.Context())

							obj = html.NewFunction("slice", []html.Token{obj, start,
								incr, stop}, context.MergeContexts(op.Context(), op2.Context()))
						case op2.Name() == "bin:":
							incr, err := p.buildExpression(op2.Args()[0:1])
							if err != nil {
								return nil, err
							}

							stop, err := p.buildExpression(op2.Args()[1:])
							if err != nil {
								return nil, err
							}

							obj = html.NewFunction("slice", []html.Token{obj, start,
								incr, stop}, context.MergeContexts(op.Context(), op2.Context()))
						}
					} else {
						var stop html.Token
						if op.Name() == "pre:" {
							stop, err = p.buildExpression(op.Args()[0:1])
							if err != nil {
								return nil, err
							}
						} else {
							stop, err = p.buildExpression(op.Args()[1:2])
							if err != nil {
								return nil, err
							}
						}

						obj = html.NewFunction("slice", []html.Token{obj, start,
							html.NewValueInt(1, op.Context()), stop}, op.Context())
					}
				default:
					panic("unhandled")
				}
			} else {
				index, err := p.buildExpression(gr.Fields[0])
				if err != nil {
					return nil, err
				}

				obj = html.NewFunction("get", []html.Token{obj, index}, gr.Context())
			}
		case gr.IsParens() && (gr.IsEmpty() || gr.IsSingle() || gr.IsComma()):
			args := make([]html.Token, 0)

			for _, field := range gr.Fields {
				arg, err := p.buildExpression(field)
				if err != nil {
					return nil, err
				}

				args = append(args, arg)
			}

			obj = html.NewFunction("eval", []html.Token{obj, html.NewValuesList(args, gr.Context())}, gr.Context())
		default:
			errCtx := gr.Context()
			err := errCtx.NewError("Error: bad indexing/evaluating")
			return nil, err
		}
	}

	return obj, nil
}
