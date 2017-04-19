package test

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

type Tags struct {
	ArticleId int64  `json:article_id`
	Name      string `json:name`
}

type RowMaper interface {
	//返回表名和字段映射
	RowMap() (string, map[string]interface{})
}

func (tags *Tags) RowMap() (tablename string, mapper map[string]interface{}) {
	return "tags", map[string]interface{}{
		"article_id": &tags.ArticleId,
		"name":       &tags.Name}
}

//将sql.Rows中的值根据 下划线命名-驼峰命名的映射关系 scan到提供的各RowMaper的字段中
//TODO  表/类型中有重名字段的问题
func MapRow(row *sql.Rows, mapper RowMaper) error {
	sqlRows, err := row.Columns()
	if err != nil {
		return err
	}
	var params []interface{}
	_, colMap := mapper.RowMap()
	for _, v := range sqlRows {
		tmp := colMap[v]
		if tmp == nil {
			var i interface{}
			tmp = &i
		}
		params = append(params, tmp)
	}
	return row.Scan(params...)
}

func InsertAndGetId(db *sql.DB, mapper RowMaper, idCol string) (id int64, e error) {
	table, m := mapper.RowMap()
	sql1 := "INSERT " + table + "("
	sql2 := ") values("
	var params []interface{}
	for k, v := range m {
		if k != idCol {
			sql1 += k + ", "
			sql2 += "?, "
			params = append(params, v)
		}
	}
	sql1 = sql1[:len(sql1)-2]
	sql2 = sql2[:len(sql2)-2]
	sql := sql1 + sql2 + ")"
	stmt, err := db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(params...)
	if err != nil {
		return -1, err
	}
	return res.LastInsertId()
}

func UpdateById(db *sql.DB, mapper RowMaper, idCol string) (rowsAffected int64, e error) {
	table, m := mapper.RowMap()
	sql := "UPDATE " + table + " set "
	var params []interface{}
	for k, v := range m {
		if k != idCol {
			sql += (k + " = ?, ")
			params = append(params, v)
		}
	}
	sql = sql[:len(sql)-2]
	if idCol != "" {
		sql += (" where " + idCol + " = ?")
		params = append(params, m[idCol])
	}
	stmt, err := db.Prepare(sql)
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(params...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}
func QueryUnique(db *sql.DB, mapper RowMaper, key string, v interface{}) error {
	table, _ := mapper.RowMap()
	rows, err := db.Query("select * from "+table+" where  "+key+" =?", v)
	if err != nil {
		return err
	}
	if rows.Next() {
		return MapRow(rows, mapper)
	} else {
		return errors.New("no rows returned")
	}

}
func DeleteById(db *sql.DB, table string, key string, v interface{}) (affectedRows int64, e error) {
	ret, err := db.Exec("delete from "+table+" where "+key+"=?", v)
	if err != nil {
		return -1, err
	} else {
		return ret.RowsAffected()
	}

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

	//验证插入
	var tag1 *Tags = &Tags{ArticleId: 2, Name: "te"}
	var tag2 *Tags = &Tags{Name: "te"}
	id, err := InsertAndGetId(db, tag1, "")
	fmt.Printf("%d---%v\n", id, err)
	id, err = InsertAndGetId(db, tag2, "aritcle_id")
	fmt.Printf("%d---%v\n", id, err)
	/*var tag *Tags = &Tags{}
	rows, _ := db.Query("select * from tags")
	defer rows.Close()
	rows.Next()
	MapRow(rows, tag)
	var tag2 *Tags = &Tags{}
	MapRow(rows, tag2)
	tag2.ArticleId = 2
	tag2.Name = "bbb"
	fmt.Printf("%d---%s\n", tag.ArticleId, tag.Name)
	fmt.Printf("%d---%s\n", tag2.ArticleId, tag2.Name)
	*/
}
