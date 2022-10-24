# XormBuilder

[![Go Reference](https://pkg.go.dev/badge/github.com/838239178/xbuild.svg)](https://pkg.go.dev/github.com/838239178/xbuild)

SQL builder base on xorm. Using a struct to build query sql.

## Config tags

**Using tag 'sql'. Split by ','. Key-value like 'k=v'**

| Option  | Meaning                                   |
| ------- | ----------------------------------------- |
| zero    | Allowed zero value                        |
| -       | skip field                                |
| no-null | Concat 'AND xx IS NOT NULL' when building |
| or      | Concat previous condition by 'OR'         |
| opt=?   | eq/in/gt... default is eq                 |
| col=?   | define column name, default is field name |
| func=?   | the function will be applied to column. Ex. TIMESTAMP |

> FieldName Operation judging has been deprecated

### Supported opt

| Opt        | Meaning           |
| ---------- | ----------------- |
| eq         | =                 |
| neq        | !=                |
| gt         | >                 |
| ge         | >=                |
| lt         | <                 |
| le         | <=                |
| in         | IN                |
| not-in/nin | NOT IN            |
| like       | LIKE %value%      |
| like-l     | LIKE %value       |
| like-r     | LIKE value%       |
| btw        | BETWEEN v1 AND v2 |

> `btw` panic if value is not array or slice or containing nil element

#### Group opt

Using '&' or '|' to group conditions with array parameter. Ex:

```go	
type Cond struct {
  Age  *[2]int 				`sql:"opt=gt&le"`       //zero value element will be ignored
  Date *[2]time.Time  `sql:"opt=gt|lt,zero"`  //allow zero value element
}
// age >= ? AND age < ? AND (date >= ? OR date < ?) 
```

> using others splitter will cause panic

## Example

See [test file](orm_builder_test.go)

**Note**: Anonymous field thinks as a part of main table and named field is a joined table. All un-exported or nil field will be ignored.

## Benchmark

Simple benchmark:

```shell
goos: darwin
goarch: amd64
pkg: github.com/838239178/xbuild
cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
BenchmarkDeepCondAlias
BenchmarkDeepCondAlias-4   	  106696	     11351 ns/op	    5889 B/op	     161 allocs/op
```

