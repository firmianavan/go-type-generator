package test

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"testing"
)

type Tags struct {
	ArticleId int64  `json:article_id`
	Name      string `json:name`
}

func (tags *Tags) MapRow(row *sql.Rows) {
	sqlRows, err := row.Columns()
	if err != nil {
		log.Fatal(err)
	}
	var params []interface{}
	var colMap = map[string]interface{}{
		"article_id": &tags.ArticleId,
		"name":       &tags.Name}
	for _, v := range sqlRows {
		tmp := colMap[v]
		if tmp == nil {
			var i int
			tmp = &i
		}
		params = append(params, tmp)
	}
	row.Scan(params...)
}
func TestGenerated(t *testing.T) {
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
	db, _ := sql.Open("mysql", connStr)
	defer db.Close()
	var tag *Tags = &Tags{}
	rows, _ := db.Query("select * from tags")
	defer rows.Close()
	rows.Next()
	tag.MapRow(rows)
	fmt.Printf("%d---%s", tag.ArticleId, tag.Name)
}
