package main

import (
	"fmt"
	"reflect"
	"strings"

	"xorm.io/builder"
)

type xbGroup struct {
	cond builder.Cond
	or   bool
}

func DeepCond(bean interface{}) (builder.Cond, error) {
	return DeepCondAlias(bean, "")
}

func DeepCondAlias(bean interface{}, tableAlias string) (builder.Cond, error) {
	// validation
	beanValue := reflect.Indirect(reflect.ValueOf(bean))
	if !beanValue.IsValid() {
		return nil, ErrNilValue
	}
	beanType := beanValue.Type()
	if beanType.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}
	// collect fields and sub-groups
	var groups []xbGroup
	fields := make([]field, 0, beanType.NumField())
	for i := 0; i < beanType.NumField(); i++ {
		//init
		f := beanType.Field(i)
		value := reflect.Indirect(beanValue.Field(i))
		if !value.IsValid() || isUnexportedField(&f) {
			continue
		}
		tg := getTag(f.Tag)
		// if value is zero and field no allow zero, skip
		if value.IsZero() && !tg.zero {
			continue
		}
		// if value is a nested struct and type not in excluding type's set
		if value.Kind() == reflect.Struct && !isExcludeStruct(value.Type()) {
			alias := IfElse(f.Anonymous, tableAlias, xormNames.Obj2Table(f.Name))
			// already skipped invalid fields so don't care errors
			cond, _ := DeepCondAlias(value.Interface(), alias)
			groups = append(groups, xbGroup{cond, tg.or})
		} else {
			// collect common fields
			fields = append(fields, field{&tg, &f, &value})
		}
	}
	cond := buildCond(fields, tableAlias)
	for _, v := range groups {
		if cond == nil {
			cond = v.cond
			continue
		}
		if v.or {
			cond = cond.Or(v.cond)
		} else {
			cond = cond.And(v.cond)
		}
	}
	return cond, nil
}

func buildCond(fs []field, alias string) builder.Cond {
	var cond builder.Cond
	for _, f := range fs {
		var actualName, cmp string
		if f.tag.opt == "" {
			//regex
			res := cmpSuffix.FindStringSubmatch(f.fie.Name)
			if len(res) == 3 {
				actualName = res[1]
				cmp = res[2]
			} else {
				actualName = f.fie.Name
				cmp = "EQ"
			}
		} else {
			actualName = f.fie.Name
			cmp = f.tag.opt
		}
		actualName = xormNames.Obj2Table(actualName)
		actualName = IfElse(alias != "", fmt.Sprintf("`%s`.`%s`", alias, actualName), actualName)
		c := getCond(cmp, actualName, f.val)
		if !f.tag.null {
			c = c.And(builder.NotNull{actualName})
		}
		if cond == nil {
			cond = c
		} else if f.tag.or {
			cond = cond.Or(c)
		} else {
			cond = cond.And(c)
		}
	}
	return cond
}

func getCond(cmp string, key string, refVal *reflect.Value) builder.Cond {
	value := refVal.Interface()
	switch strings.ToUpper(cmp) {
	case "IN":
		fallthrough
	case "EQ":
		return builder.Eq{key: value}
	case "NIN":
		fallthrough
	case "NOT-IN":
		fallthrough
	case "NEQ":
		return builder.Neq{key: value}
	case "GT":
		return builder.Gt{key: value}
	case "LT":
		return builder.Lt{key: value}
	case "LE":
		return builder.Lte{key: value}
	case "GE":
		return builder.Gte{key: value}
	case "LIKE-L":
		return builder.Like{key, fmt.Sprint("%", value)}
	case "LIKE-R":
		return builder.Like{key, fmt.Sprint(value, "%")}
	case "LIKE":
		return builder.Like{key, fmt.Sprint("%", value, "%")}
	case "BTW":
		return builder.Between{
			Col: key, 
			LessVal: refVal.Index(0).Interface(), 
			MoreVal: refVal.Index(1).Interface(),
		}	
	}
	panic("unknown " + cmp)
}
