package xbuild

import (
	"errors"
	"reflect"
	"strings"
	"time"
	"xorm.io/xorm/names"
)

type sqlTag struct {
	ignore bool   //ignore ignoring field
	zero   bool   //zero allowed zero value if true otherwise skip it (default false)
	or     bool   //or use OR to concat condition (default false)
	null   bool   //null allowed null value if true otherwise concat 'AND xxx IS NOT NULL' (default true)
	opt    string //opt eq/lt/ge...
	col    string //column name
}

func (s *sqlTag) Oper() string {
	return ifElse(s.opt == "", "EQ", s.opt)
}

func (s *sqlTag) Column(def string) string {
	return ifElse(s.col == "", def, s.col)
}

type field struct {
	tag *sqlTag
	fie *reflect.StructField
	val reflect.Value
}

func (f *field) IsValid() bool {
	return f.val.IsValid() && (!f.val.IsZero() || f.tag.zero)
}

var (
	ErrNilValue  = errors.New("nil value")
	ErrNotStruct = errors.New("not struct")
)

var xormNames names.Mapper = names.GonicMapper{}
var excludeTypes = map[reflect.Type]struct{}{
	reflect.TypeOf(time.Time{}): {},
}

func SetXormNames(mapper names.Mapper) {
	xormNames = mapper
}

func isExcludeStruct(refType reflect.Type) bool {
	var ok bool
	_, ok = excludeTypes[refType]
	return ok
}

func getTag(tg reflect.StructTag) (gtg sqlTag) {
	gtg.null = true
	if str, ok := tg.Lookup("sql"); ok {
		lst := strings.Split(str, ",")
		for _, opt := range lst {
			switch opt {
			case "-":
				gtg.ignore = true
				return
			case "zero":
				gtg.zero = true
			case "or":
				gtg.or = true
			case "no-null":
				gtg.null = false
			default:
				if strings.ContainsRune(opt, '=') {
					if kv := strings.Split(opt, "="); len(kv) > 1 {
						switch kv[0] {
						case "opt":
							gtg.opt = kv[1]
						case "col":
							gtg.col = kv[1]
						}
					}
				}
			}
		}
	}
	return
}

func ifElse(cond bool, trueStr, falseStr string) string {
	if cond {
		return trueStr
	}
	return falseStr
}

func orElse(cond bool, trueVal, falseVal interface{}) interface{} {
	if cond {
		return trueVal
	}
	return falseVal
}

func realKind(val reflect.Value) reflect.Kind {
	return reflect.Indirect(val).Kind()
}

func realType(val reflect.Value) reflect.Type {
	return reflect.Indirect(val).Type()
}

func validPtr2Struct(val reflect.Value) error {
	if val.Kind() != reflect.Ptr {
		return errors.New("needs a pointer to a value")
	} else if val.Elem().Kind() == reflect.Ptr {
		return errors.New("a pointer to a pointer is not allowed")
	} else if val.IsNil() {
		return ErrNilValue
	}
	return nil
}

// isNilPtr only a reference type may be true; zero value always false
func isNilPtr(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice,
		reflect.UnsafePointer:
		return val.IsNil()
	default:
		return false
	}
}

func isExported(val reflect.Value) bool {
	return val.CanInterface()
}

// toInterface avoid big mem-copy
func toInterface(val reflect.Value) interface{} {
	switch val.Kind() {
	case reflect.Chan, reflect.Func,
		reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice,
		reflect.UnsafePointer:
		return val.Interface()
	case reflect.Struct, reflect.Array:
		return val.Addr().Interface()
	}
	return val.Interface()
}
