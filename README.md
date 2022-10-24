[English doc](./README_eng.md)

# XormBuilder

[![Go Reference](https://pkg.go.dev/badge/github.com/838239178/xbuild.svg)](https://pkg.go.dev/github.com/838239178/xbuild)

利用结构体TAG轻松构造XORM复杂的条件查询

## 可配置的TAG

**TAG名为‘sql’，使用‘，’分割**

| Option  | Meaning                                   |
| ------- | ----------------------------------------- |
| zero    | 允许零值，默认当字段为零值时忽略条件                        |
| -       | 跳过这个字段                                |
| no-null | 额外拼接 'AND xx IS NOT NULL' 到条件中 |
| or      | 使用 'OR' 与上一个字段拼接         |
| opt=?   | 比较条件 eq/in/gt... 默认为eq               |
| col=?   | 定义列名，默认为结构体字段名 |
| func=?   | 设置一个可以应用到列名的SQL函数 Ex. TIMESTAMP |

### 支持的查询条件

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

> 如果`btw`对应的字段不是切片或者数组，会发生PANIC

#### 组合条件

使用分隔符 '&' or '|' 和数组或切片字段形成一组会被括号起来的条件. Ex:

```go	
type Cond struct {
  Age  *[2]int 				`sql:"opt=gt&le"`       //zero value element will be ignored
  Date *[2]time.Time  `sql:"opt=gt|lt,zero"`  //allow zero value element
}
// age > ? AND age <= ? AND (date > ? OR date < ?) 
```

> 使用嵌套结构体也可以实现条件分组，具体看测试文件

> 使用其他分隔符会造成PANIC

## 例子

具体查看 [测试文件](orm_builder_test.go)

**Note**: 未导出字段和Nil指针将被忽略，匿名结构体（组合）被视为一张表的组合条件，但是具名结构体会被视为另一张表，表明与结构体字段的名称一致

## 基准测试

简易基准测试:

```shell
goos: darwin
goarch: amd64
pkg: github.com/838239178/xbuild
cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
BenchmarkDeepCondAlias
BenchmarkDeepCondAlias-4   	  106696	     11351 ns/op	    5889 B/op	     161 allocs/op
```

