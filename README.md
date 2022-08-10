# XormBuilder

SQL builder base on xorm. Using a struct to build query sql.

## Config tags

**Using tag 'sql'. Split by ','. Key-value like 'k=v'**

| Option  | Meaning                                                    |
| ------- | ---------------------------------------------------------- |
| zero    | Allowed zero value                                         |
| no-null | Concat 'AND xx IS NOT NULL' when building                  |
| or      | Concat previous condition by 'OR'                          |
| opt     | eq/in/gt... If this doesn't exist, judging from field name |

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

> `btw` panic if value is not array or slice or contains nil element

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
BenchmarkDeepCondAlias-4   	   88532	     14760 ns/op
```

