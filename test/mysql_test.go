package test

import (
	"fmt"
	"github.com/firmianavan/go-type-generator/mysql"
	"testing"
)

func TestMysql(t *testing.T) {
	/*cases := []struct {
		ori string
		des string
	}{
		{"", ""},
		{"abc", "Abc"},
		{"abc_de", "AbcDe"},
		{"a_d", "AD"},
		{"你", "你"},
	}*/

	var connStr string = "van:123456@tcp(127.0.0.1:3306)/wmp"
	mysql.Construct(connStr)
	defer mysql.Close()
	fmt.Println(mysql.GenTypeFromTable("tags"))
}
