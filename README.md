# XormBuilder

SQL builder base on xorm. Using a struct to build query sql.

## Config tags

**using tag 'sql'**

| Option  | Meaning                                   |
| ------- | ----------------------------------------- |
| zero    | Allowed zero value                        |
| no-null | Concat 'AND xx IS NOT NULL' when building |
| or      | Concat previous condition by 'OR'          |

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

