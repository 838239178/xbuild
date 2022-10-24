package xbuild

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
)

type TestBean struct {
	TestGroup3
	ID []int64 `sql:"opt=in"`
}

type TestGroup3 struct {
	Major      TestGroup
	TestGroup2 `sql:"or"` //contact previous condition with 'OR'
}

type TestGroup2 struct {
	Name       string        `sql:"zero"`                        //zero means allow zero value
	PostDate   time.Time     `sql:"zero,no-null,func=TIMESTAMP"` //means '(TIMESTAMP(post_date) = 'xx' AND post_date IS NOT NULL)'
	CreateDate *[2]time.Time `sql:"opt=btw"`                     //btw means 'create_date BETWEEN ? AND ?'; nil value cause panic; won't ignore zero value even if there has the `zero` tag
	Info       string        `sql:"opt=like-r"`                  //like-r means 'info LIKE ?%'; like-l means '%?'; like means '%?%';
}

type TestGroup struct {
	Age *[2]int `sql:"opt=gt&le"` // zero value in array will be ignored if no `zero` tag; `&` means 'AND'; `|` means 'OR'
}

type TestSplitter struct {
	Age  *[2]int    `sql:"opt=gt&le"`
	Date *[2]string `sql:"opt=ge|le"`
	//StartDate string     `sql:"opt=gt,col=date"`	//col define the actual column name
	//EndDate   string     `sql:"opt=lt,col=date"`
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
	xormDB.SetMapper(names.GonicMapper{})
	xormDB.ShowSQL(true)
}

func TestXormBuilderSplitter(t *testing.T) {
	cond, _ := DeepCond(&TestSplitter{
		Age:  &[2]int{10, 20},
		Date: &[2]string{"2022-11-01", ""},
	})
	// SELECT `id`, `age`, `major_id`, `post_date` FROM `tb` WHERE age>? AND age<=? AND date>=? LIMIT 1
	_, _ = xormDB.Where(cond).Get(&TestTable{})
}

func TestXormBuilder(t *testing.T) {
	cond, _ := DeepCondAlias(&TestBean{
		ID: []int64{1, 2, 3},
		TestGroup3: TestGroup3{
			Major: TestGroup{
				Age: &[2]int{10, 20},
			},
			TestGroup2: TestGroup2{
				Name:       "Yes",
				PostDate:   time.Now(),
				CreateDate: &[2]time.Time{time.Now(), time.Now()},
				Info:       "search",
			},
		},
	}, "tb")
	// SELECT tb.id, tb.age, tb.major_id
	// FROM `tb` INNER JOIN `major` ON major.id = tb.major_id
	// WHERE `tb`.`id` IN (?,?,?)
	// AND (
	//	(`major`.`age`>? AND `major`.`age`<=?)
	//	OR (
	//		`tb`.`name`=?
	//		AND TIMESTAMP(`tb`.`post_date`) = ?
	//		AND `tb`.`post_date` IS NOT NULL
	//		AND `tb`.`create_date` BETWEEN ? AND ?
	//		AND `tb`.`info` LIKE ?
	//	)
	//)
	_, _ = xormDB.Table(&TestTable{}).Alias("tb").Where(cond).
		Select("tb.id, tb.age, tb.major_id").
		Join("INNER", "major", "major.id = tb.major_id").
		Get(&TestTable{})
}

func BenchmarkDeepCondAlias(b *testing.B) {
	bean := &TestBean{
		ID: []int64{1, 2, 3},
		TestGroup3: TestGroup3{
			Major: TestGroup{
				Age: &[2]int{10, 20},
			},
			TestGroup2: TestGroup2{
				Name:       "Yes",
				PostDate:   time.Now(),
				CreateDate: &[2]time.Time{time.Now(), time.Now()},
				Info:       "search",
			},
		},
	}
	for n := 0; n < b.N; n++ {
		_, _ = DeepCondAlias(bean, "tb")
	}
}

func TestReflectValue(t *testing.T) {
	bean := &TestTable{
		Age: 20,
	}
	fn := func(t interface{}) {
		switch t := t.(type) {
		case TestTable:
			t.Age += 10
		case *TestTable:
			t.Age -= 10
		}
	}
	beanVal := reflect.Indirect(reflect.ValueOf(bean))
	fn(beanVal.Addr().Interface())
	t.Logf("%#v", bean)
}
