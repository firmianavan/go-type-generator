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
    RowMap() (string, map[string]Column)
}

type ColumnMata struct {
    Field   string         //数据库字段名
    Type    string         //数据库的字段类型, 可以用来做长度验证
    Null    string         //是否可为空"YES"or"NO"
    Key     string         //"PRI","UNI"
    Default string         //默认值
    Extra   string         //
    GoType  string         //go中对应字段使用的类型
    GoField string         //go中对应字段名
}
type Column struct {
    Meta *ColumnMata
    V interface{}
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
    tmpString := make(map[string]*sql.NullString) //用来存放数据库得到的NullString类型, 便于后面统一装换为string类型
    _, colMap := mapper.RowMap()
    for _, v := range sqlRows {
        tmp, exists := colMap[v]
        if !exists {
            var i interface{}
            params = append(params, &i) //舍弃
        } else if tmp.Meta.GoType == "string" && tmp.Meta.Null == "YES" { //先scan到sql.NullString, 在转换到对象中的string
            tmpNullString := sql.NullString{}
            tmpString[v] = &tmpNullString
            params = append(params, &tmpNullString)
        } else {
            params = append(params, tmp.V)
        }

    }
    err = row.Scan(params...)
    if err != nil {
        return err
    }
    for k, v := range tmpString {
        if strPtr, ok := (colMap[k].V).(*string); ok {
            *strPtr = v.String
        }
    }
    return err
}

//insert, return an error if exists. if you want to abandon some columns, add their name into "ommit" param
func Insert( mapper RowMaper, ommit ...string) error {
    db:=GetDB()
    table, m := mapper.RowMap()
    sql1 := "INSERT " + table + "("
    sql2 := ") values("
    var params []interface{}
    for k, v := range m {
        if !contains(k, ommit...) {
            sql1 += k + ", "
            sql2 += "?, "
            params = append(params, v.V)
        }
    }
    sql1 = sql1[:len(sql1)-2]
    sql2 = sql2[:len(sql2)-2]
    sql := sql1 + sql2 + ")"
    stmt, err := db.Prepare(sql)
    if err != nil {
        return err
    }
    _, err = stmt.Exec(params...)
    return err
}

//insert, return generated id by db and an error if exists. if you want to abandon some columns, add their name into "ommit" param
func InsertAndGetId(mapper RowMaper, idCol string, ommit ...string) (id int64, e error) {
    db:=GetDB()
    table, m := mapper.RowMap()
    sql1 := "INSERT " + table + "("
    sql2 := ") values("
    var params []interface{}
    for k, v := range m {
        if k != idCol && !contains(k, ommit...) {
            sql1 += k + ", "
            sql2 += "?, "
            params = append(params, v.V)
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

//update ** where idCol=?, return rows affected and an error if exists. if you want to abandon some columns, add their name into "ommit" param
func UpdateById( mapper RowMaper, idCol string, ommit ...string) (rowsAffected int64, e error) {
    db:=GetDB()
    table, m := mapper.RowMap()
    sql := "UPDATE " + table + " set "
    var params []interface{}
    for k, v := range m {
        if k != idCol && !contains(k, ommit...) {
            sql += (k + " = ?, ")
            params = append(params, v.V)
        }
    }
    sql = sql[:len(sql)-2]
    if idCol != "" {
        sql += (" where " + idCol + " = ?")
        params = append(params, m[idCol].V)
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

func contains(target string, sets ...string) bool {
    for _, s := range sets {
        if target == s {
            return true
        }
    }
    return false
}

//select where key=v, if return more than one row, only the first mapped into param mapper
func QueryUnique( mapper RowMaper, key string, v interface{}) error {
    db:=GetDB()
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

//delete ** where key=v, return rows affected and an error if exists.
func DeleteById( table string, key string, v interface{}) (affectedRows int64, e error) {
    db:=GetDB()
    ret, err := db.Exec("delete from "+table+" where "+key+"=?", v)
    if err != nil {
        return -1, err
    } else {
        return ret.RowsAffected()
    }

}

`

var DataSource string = `

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "log"
    "sync"
)

var db *sql.DB
var once sync.Once
var connStr string = "%s?parseTime=true&loc=UTC"

func GetDB() *sql.DB {
    once.Do(func() {
        var err error
        db, err = sql.Open("mysql", connStr)
        if err != nil {
            log.Fatal("program exit: failed to connect the database %s", connStr)
        }
    })
    return db
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
	constVar := ""
	metaSingleton := ""
	for i, _ := range cols {
		cols[i].GoType = mapFromSqlType(cols[i].Type, cols[i].Null)
		cols[i].GoField = common.ChangeNameToCamel(cols[i].Field, "_")
		constVar += ifEnum(cols[i].Type, camel, cols[i].GoField)
		metaSingleton += fmt.Sprintf("var %s%s ColumnMata = ColumnMata{ Field:\"%s\",Type:\"%s\",Null:\"%s\",Key:\"%s\",Default:\"%s\",Extra:\"%s\",GoType:\"%s\",GoField:\"%s\"}\n",
			tableName, cols[i].GoField, cols[i].Field, cols[i].Type, cols[i].Null, cols[i].Key, cols[i].Default.String, cols[i].Extra, cols[i].GoType, cols[i].GoField)
		buffer.WriteString(cols[i].String())
	}
	buffer.WriteString("\n}\n")
	buffer.WriteString(constVar)
	buffer.WriteString(metaSingleton)
	buffer.WriteString(fmt.Sprintf("func (%v *%v) RowMap()(tableName string, colMap map[string]Column) {\n", tableName, camel))
	buffer.WriteString("    colMap = map[string]Column{\n")
	for i, v := range cols {
		if i != 0 {
			buffer.WriteString(",\n")
		}
		buffer.WriteString(fmt.Sprintf("\"%s\": Column{Meta : &%s%s, V : &%s.%s}", v.Field, tableName, v.GoField, tableName, v.GoField))
	}
	buffer.WriteString("    }\n")
	buffer.WriteString("    return \"" + tableName + "\",colMap\n")
	buffer.WriteString("    }\n\n\n")

	return buffer.String()
}

func mapFromSqlType(sqlType string, nullAble string) string {
	ifNull := func(currentType string) string {
		if "yes" == strings.ToLower(nullAble) {
			return "sql.Null" + strings.Title(currentType)
		} else {
			return currentType
		}
	}
	if strings.Contains(sqlType, "int") {
		return ifNull("int64")
	} else if strings.Contains(sqlType, "char") || strings.Contains(sqlType, "text") {
		return "string"
	} else if strings.Contains(sqlType, "date") || strings.Contains(sqlType, "timestamp") {
		return "time.Time"
	} else if strings.Contains(sqlType, "float") || strings.Contains(sqlType, "double") {
		return ifNull("float64")
	} else if strings.Contains(sqlType, "enum") {
		return "string"
	} else {
		return "[]byte"
	}
}

func ifEnum(sqlType, typeName, fieldName string) string {
	ret := ""
	if strings.Contains(sqlType, "enum") {
		tmp := sqlType[strings.Index(sqlType, "('")+2 : strings.LastIndex(sqlType, "')")]
		t := strings.Split(tmp, "','")
		for i := 0; i < len(t); i++ {
			des := strings.Replace(t[i], "-", "_", -1) //变量命名不可以有中划线, 会被认为是减号
			ret += fmt.Sprintf("const Const_%s_%s_%s string = \"%s\"\n", typeName, fieldName, strings.Title(des), des)
		}
		//fmt.Println(ret)
	}
	return ret
}

func (v *MetaCol) String() string {
	return fmt.Sprintf("\n    %s %s `json:%s`", v.GoField, v.GoType, v.Field)
}
