package html

import (
  "errors"
  "image/color"
  "math"
  "reflect"

  "github.com/computeportal/wtsuite/pkg/tokens/context"
)

// convert basic golang types into tokens:
// * bool -> html.Bool
// * int -> html.Int
// * string -> html.String
// * float64 -> html.Float (or html.Int if whole number)
// * []interface{} -> html.List
// * map[string]interface{} -> html.StringDict
// * nil -> html.Null
// * color.RGBA -> html.Color

// there is no conversion for a united float though

func GolangToToken(x_ interface{}, ctx context.Context) (Token, error) {
  if x_ == nil {
    return NewNull(ctx), nil
  }

  switch x := x_.(type) {
  case bool:
    return NewValueBool(x, ctx), nil
  case int:
    return NewValueInt(x, ctx), nil
  case float64:
    // round numbers are integers
    if math.Mod(x, 1.0) == 0.0 {
      return NewValueInt(int(x), ctx), nil
    } else {
      return NewValueFloat(x, ctx), nil
    }
  case string:
    return NewValueString(x, ctx), nil
  case color.RGBA:
    return NewValueColor(int(x.R), int(x.G), int(x.B), int(x.A), ctx), nil
  case *color.RGBA:
    return NewValueColor(int(x.R), int(x.G), int(x.B), int(x.A), ctx), nil
  case []interface{}:
    return GolangSliceToList(x, ctx)
  case map[string]interface{}:
    return GolangStringMapToStringDict(x, ctx)
  default:
    return nil, errors.New("unsupported type: " + reflect.TypeOf(x_).String())
  }
}
