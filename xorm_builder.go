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
	beanValue := reflect.ValueOf(bean)
	// bean must be a ptr to struct
	if err := validPtr2Struct(beanValue); err != nil {
		return nil, err
	}
	beanType := realType(beanValue)
	// collect fields and sub-groups
	var groups []*xbGroup
	fields := make([]*field, 0, beanType.NumField())
	for i := 0; i < beanType.NumField(); i++ {
		f := beanType.Field(i)
		value := beanValue.Elem().Field(i)
		if isNilPtr(value) || !isExported(value) {
			// skip if value is nil or un-exported
			continue
		}
		tg := getTag(f.Tag)
		if value.IsZero() && !tg.zero {
			// if value is zero and field no allow zero, skip
			continue
		}
		// if value is a nested struct and type not in excluding type's set
		if realKind(value) == reflect.Struct && !isExcludeStruct(realType(value)) {
			alias := ifElse(f.Anonymous, tableAlias, xormNames.Obj2Table(f.Name))
			// already skipped invalid fields so don't care errors
			cond, _ := DeepCondAlias(toInterface(value), alias)
			groups = append(groups, &xbGroup{cond, tg.or})
		} else {
			// collect common fields
			fields = append(fields, &field{&tg, &f, value})
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

func buildCond(fs []*field, alias string) builder.Cond {
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
		actualName = ifElse(alias != "", fmt.Sprintf("`%s`.`%s`", alias, actualName), actualName)
		var c builder.Cond
		if strings.ContainsAny(cmp, "&|") {
			c = arrayCond(cmp, actualName, f.val, f.tag)
		} else {
			c = getCond(cmp, actualName, f.val)
		}
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

func arrayCond(cmps string, key string, refVal reflect.Value, tg *sqlTag) builder.Cond {
	var temp builder.Cond
	cmp, idx, splitter := make([]rune, 0, 2), 0, rune(0)
	refVal = reflect.Indirect(refVal)
	for _, r := range cmps {
		if isSplitter(r) {
			elemVal := refVal.Index(idx)
			idx++
			if !isNilPtr(elemVal) && (tg.zero || !elemVal.IsZero()) {
				temp = condBySplitter(temp, getCond(string(cmp), key, elemVal), splitter)
			}
			splitter, cmp = r, cmp[0:0]
		} else {
			cmp = append(cmp, r)
		}
	}
	elemVal := refVal.Index(idx)
	if !isNilPtr(elemVal) && (tg.zero || !elemVal.IsZero()) {
		return condBySplitter(temp, getCond(string(cmp), key, elemVal), splitter)
	}
	return orElse(temp == nil, builder.NewCond(), temp).(builder.Cond)
}

func isSplitter(r rune) bool {
	return r == '&' || r == '|'
}

func condBySplitter(cond1, cond2 builder.Cond, splitter rune) builder.Cond {
	if cond1 == nil {
		return cond2
	}
	switch splitter {
	case '&':
		cond1 = cond1.And(cond2)
	case '|':
		cond1 = cond1.Or(cond2)
	default:
		panic("unknown splitter: " + string(splitter))
	}
	return cond1
}

func getCond(cmp string, key string, refVal reflect.Value) builder.Cond {
	value := toInterface(refVal)
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
		realVal := reflect.Indirect(refVal)
		return builder.Between{
			Col:     key,
			LessVal: toInterface(realVal.Index(0)),
			MoreVal: toInterface(realVal.Index(1)),
		}
	}
	panic("unknown " + cmp)
}
