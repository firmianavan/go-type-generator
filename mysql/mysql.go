package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/firmianavan/go-type-generator/common"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

type MetaCol struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
	GoType  string
	GoField string
}

var db *sql.DB

func Construct(connStr string) {
	var err error
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

func Close() {
	db.Close()
}

var InterfaceDefination string = `
type RowMaper interface {
    //返回表名和字段映射
    RowMap() (string, map[string]interface{})
}
`
var ExportedMethods string = `
//将sql.Rows中的值根据 下划线命名-驼峰命名的映射关系 scan到提供的RowMaper的各字段中
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
`

func GetTables() []string {
	rows, err := db.Query("show tables")
	if err != nil {
		log.Fatal("failed to list tables, err is: %v", err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		err := rows.Scan(&t)
		if err != nil {
			log.Fatal(err)
		}
		tables = append(tables, t)
	}
	return tables
}

func GenTypeFromTable(tableName string) string {
	rows, err := db.Query("desc " + tableName)
	if err != nil {
		log.Fatal("failed to query table %s, err is: %v", tableName, err)
	}
	defer rows.Close()
	var table []MetaCol
	for rows.Next() {
		col := MetaCol{}
		err := rows.Scan(&col.Field, &col.Type, &col.Null, &col.Key, &col.Default, &col.Extra)
		if err != nil {
			log.Fatal(err)
		}
		table = append(table, col)
	}
	return genText(tableName, table)
}

func genText(tableName string, cols []MetaCol) string {
	var buffer bytes.Buffer
	//buffer.WriteString("package entity\n\n")
	camel := common.ChangeNameToCamel(tableName, "_")
	buffer.WriteString("type " + camel + " struct {")
	for i, _ := range cols {
		cols[i].GoType = mapFromSqlType(cols[i].Type)
		cols[i].GoField = common.ChangeNameToCamel(cols[i].Field, "_")
		buffer.WriteString(cols[i].String())
	}
	buffer.WriteString("\n}\n")
	buffer.WriteString(fmt.Sprintf("func (%v *%v) RowMap()(tableName string, colMap map[string]interface{}) {\n", tableName, camel))
	buffer.WriteString("    var colMap = map[string]interface{}{\n")
	for i, v := range cols {
		if i != 0 {
			buffer.WriteString(",\n")
		}
		buffer.WriteString(fmt.Sprintf("\"%s\": &%s.%s", v.Field, tableName, v.GoField))
	}
	buffer.WriteString("    }\n")
	buffer.WriteString("    return " + tableName + ",colMap\n")
	buffer.WriteString("    }\n\n\n")

	return buffer.String()
}

func mapFromSqlType(sqlType string) string {
	if strings.Contains(sqlType, "int") {
		return "int64"
	} else if strings.Contains(sqlType, "char") || strings.Contains(sqlType, "text") {
		return "string"
	} else if strings.Contains(sqlType, "date") || strings.Contains(sqlType, "timestamp") {
		return "time.Time"
	} else if strings.Contains(sqlType, "float") || strings.Contains(sqlType, "double") {
		return "float64"
	} else {
		return "[]byte"
	}
}

func (v *MetaCol) String() string {
	return fmt.Sprintf("\n    %s %s `json:%s`", v.GoField, v.GoType, v.Field)
}
