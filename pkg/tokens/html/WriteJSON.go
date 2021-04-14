package html

import (
	"fmt"
	"strings"
)

func isMultiLineList(lst *List) bool {
  multiline := false

  if err := lst.Loop(func(i int, t Token, last bool) error {
    switch t.(type) {
    case *Float:
    case *String:
    case *Int:
    case *Bool:
    case *Null:
      return nil
    default:
      multiline = true
    }

    return nil
  }); err != nil {
    panic(err)
  }

  return multiline
}

func WriteJSON(argToken Token, indent string, tab string, nl string) (string, error) {
	var b strings.Builder

	switch tt := argToken.(type) {
	case *StringDict:
		b.WriteString("{")
		if err := tt.Loop(func(key *String, val Token, last bool) error {
      b.WriteString(nl)
      b.WriteString(indent)
      b.WriteString(tab)
			b.WriteString("\"")
			b.WriteString(key.Value())
			b.WriteString("\":")
			sub, err := WriteJSON(val, indent + tab, tab, nl)
			if err != nil {
				return err
			}
			b.WriteString(sub)
			if !last {
				b.WriteString(",")
			}
			return nil
		}); err != nil {
			return b.String(), err
		}

    b.WriteString(nl)
    b.WriteString(indent)
		b.WriteString("}")
	case *IntDict:
		b.WriteString("{")
		if err := tt.Loop(func(key *Int, val Token, last bool) error {
      b.WriteString(nl)
      b.WriteString(indent)
      b.WriteString(tab)
			b.WriteString("\"")
			b.WriteString(fmt.Sprintf("%d", key.Value()))
			b.WriteString("\":")
			sub, err := WriteJSON(val, indent + tab, tab, nl)
			if err != nil {
				return err
			}
			b.WriteString(sub)
			if !last {
				b.WriteString(",")
			}
			return nil
		}); err != nil {
			return b.String(), err
		}

    b.WriteString(nl)
    b.WriteString(indent)
		b.WriteString("}")
	case *List:
		b.WriteString("[")
    multiline := isMultiLineList(tt)

		if err := tt.Loop(func(i int, item Token, last bool) error {
			sub, err := WriteJSON(item, indent + tab, tab, nl)
			if err != nil {
				return err
			}

      if multiline {
        b.WriteString(nl)
        b.WriteString(indent)
        b.WriteString(tab)
      }

			b.WriteString(sub)
			if !last {
				b.WriteString(",")
			}
			return nil
		}); err != nil {
			return b.String(), err
		}

    if multiline {
      b.WriteString(nl)
      b.WriteString(indent)
    }

		b.WriteString("]")
	case *Int:
		b.WriteString(tt.Write())
	case *Float:
		if tt.Unit() != "" {
			b.WriteString("\"")
		}
		b.WriteString(tt.Write())
		if tt.Unit() != "" {
			b.WriteString("\"")
		}
	case *Bool:
		if tt.Value() {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case *Null:
		b.WriteString("null")
	case Primitive:
		b.WriteString("\"")
		b.WriteString(tt.Write())
		b.WriteString("\"")
	default:
		errCtx := tt.Context()
		return b.String(), errCtx.NewError("Error: expected string, dict or list")
	}

	return b.String(), nil
}
