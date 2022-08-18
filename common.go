package main

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"xorm.io/xorm/names"
)

type sqlTag struct {
	zero bool   //zero allowed zero value if true otherwise skip it (default false)
	or   bool   //or use OR to concat condition (default false)
	null bool   //null allowed null value if true otherwise concat 'AND xxx IS NOT NULL' (default true)
	opt  string //opt eq/lt/ge...
}

type field struct {
	tag *sqlTag
	fie *reflect.StructField
	val *reflect.Value
}

func (f *field) IsValid() bool {
	return f.val.IsValid() && (!f.val.IsZero() || f.tag.zero)
}

var (
	ErrNilValue  = errors.New("nil value")
	ErrNotStruct = errors.New("not struct")
)

var cmpSuffix = regexp.MustCompile(`^(\w+)(GT|LT|GE|LE|EQ|NEQ|IN|NIN)$`)
var xormNames names.Mapper = names.GonicMapper{}
var excludeTypes = map[reflect.Type]struct{}{
	reflect.TypeOf(time.Time{}): {},
}

func isUnexportedField(f *reflect.StructField) bool {
	name := f.Name
	if f.Anonymous {
		name = f.Type.Name()
	}
	return unicode.IsLower([]rune(name[0:1])[0])
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
						}
					}
				}
			}
		}
	}
	return
}

func IfElse(cond bool, trueStr, falseStr string) string {
	if cond {
		return trueStr
	}
	return falseStr
}
