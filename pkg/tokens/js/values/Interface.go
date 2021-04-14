package values

import (
	"github.com/wtsuite/wtsuite/pkg/tokens/context"
)

type Interface interface {
  Name() string

  Context() context.Context

  Check(other Interface, ctx context.Context) error

  // if true: can be exported to databases etc.
  IsUniversal() bool // for actual interfaces: all implementations need to be universal
  IsRPC() bool // 

  // get extended interfaces in case of js.Interface, get implements interfaces in case of js.Class
  GetInterfaces() ([]Interface, error)

  // get all prototypes that implement this interface (actual prototypes dont need include themselves) (used by InstanceOf.Write())
  GetPrototypes() ([]Prototype, error)

  // returns nil if it doesnt exist
  GetInstanceMember(key string, includePrivate bool, ctx context.Context) (Value, error)

  SetInstanceMember(key string, includePrivate bool, arg Value, ctx context.Context) error
}

// returns nil if not an Instance with an Interface
func GetInterface(v_ Value) Interface {
  v_ = UnpackContextValue(v_)

  switch v := v_.(type) {
  case *Instance:
    return v.GetInterface()
  case *Tuple:
    return v.GetInterface()
  case *LiteralIntInstance:
    return v.GetInterface()
  case *LiteralBooleanInstance:
    return v.GetInterface()
  case *LiteralStringInstance:
    return v.GetInterface()
  default:
    return nil
  }
}

// returns nil if some error
func GetArrayContent(interf Interface) Value {
  var err error

  key := ".getindex"
  ctx := context.NewDummyContext()

  interf, err = FindInstanceMemberInterface(interf, key, false, ctx)
  if err != nil {
    return nil
  }

  fn, err := interf.GetInstanceMember(key, false, ctx)
  if err != nil {
    return nil
  }

  retVal, err := fn.EvalFunction([]Value{NewAny(ctx)}, false, ctx)
  if err != nil {
    return nil
  }

  return retVal
}
