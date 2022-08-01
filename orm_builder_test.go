package main

import (
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
	"xorm.io/xorm"
)

type TestBean struct {
	TestGroup3
	IDIN []int64
}

type TestGroup3 struct {
	Major      TestGroup
	TestGroup2 `sql:"or"` //与前一个组合通过OR拼接
}

type TestGroup2 struct {
	Name     string    `sql:"zero"`         //zero 允许零值
	PostDate time.Time `sql:"zero,no-null"` //no-null将会生成一个组合判断：'(post_date = 'xx' AND post_date IS NOT NULL)'
}

type TestGroup struct {
	AgeGE int
	AgeLT int //默认使用and拼接并忽略零值
}

type TestTable struct {
	ID       int64
	Age      int
	MajorID  int64
	PostDate time.Time
}

var xormDB *xorm.Engine

func init() {
	var dsn []byte
	file, err := os.Open("dsn.txt")
	if err != nil {
		panic(err)
	}
	if dsn, err = ioutil.ReadAll(file); err != nil {
		panic(err)
	}
	if xormDB, err = xorm.NewEngine("mysql", strings.TrimSpace(string(dsn))); err != nil {
		panic(err)
	}
}

func TestXormBuilder(t *testing.T) {
	cond, _ := DeepCondAlias(TestBean{
		IDIN: []int64{1, 2, 3},
		TestGroup3: TestGroup3{
			Major: TestGroup{
				AgeGE: 12,
				AgeLT: 100,
			},
			TestGroup2: TestGroup2{
				Name:     "Yes",
				PostDate: time.Now(),
			},
		},
	}, "test_table")
	xormDB.ShowSQL(true)
	_, _ = xormDB.Where(cond).
		Select("id,age,major_id").
		Join("INNER", "major", "major.id = test_table.major_id").
		Get(&TestTable{})
}
