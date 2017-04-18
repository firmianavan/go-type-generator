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
	buffer.WriteString(fmt.Sprintf("func (%v *%v) MapRow(row *sql.Rows) {\n", tableName, camel))
	buffer.WriteString("    sqlRows,err:=row.Columns()\n")
	buffer.WriteString("    if err!=nil {\n")
	buffer.WriteString("        log.Fatal(err)\n")
	buffer.WriteString("    }\n")
	buffer.WriteString("    var params []interface{}\n")
	buffer.WriteString("    var colMap = map[string]interface{}{\n")
	for i, v := range cols {
		if i != 0 {
			buffer.WriteString(",\n")
		}
		buffer.WriteString(fmt.Sprintf("\"%s\": &%s.%s", v.Field, tableName, v.GoField))
	}
	buffer.WriteString("    }\n")
	buffer.WriteString("    for _,v:= range sqlRows {\n")
	buffer.WriteString("        tmp :=colMap[v]\n")
	buffer.WriteString("        if tmp ==nil {\n")
	buffer.WriteString("            var i int\n")
	buffer.WriteString("            tmp =&i\n")
	buffer.WriteString("        }\n")
	buffer.WriteString("        params = append(params,tmp)\n")
	buffer.WriteString("    }\n")
	buffer.WriteString("    row.Scan(params...)\n}\n\n\n")
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
