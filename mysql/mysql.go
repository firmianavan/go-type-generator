package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/firmianavan/go-type-generator/common"
	"github.com/firmianavan/go-type-generator/tmpl"
	_ "github.com/go-sql-driver/mysql"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"
)

var driverName string = "github.com/go-sql-driver/mysql"

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
type ConstVar struct {
	StructName string
	FieldName  string
	Val        string
}
type SingleType struct {
	TableName string
	Camel     string
	Columns   []MetaCol
	Consts    []ConstVar
}
type TemplateContext struct {
	DriverName  string
	PackageName string
	ConnStr     string
	Packages    []string
	Types       []SingleType
}

var db *sql.DB

//入口方法
func GenCode(connStr, packageName string, tables []string, genMethods bool) {
	var err error
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	data := prepareData(connStr, packageName, tables)

	genFile(tmpl.EntityTemplate, "entity.go", data)
	genFile(tmpl.EntityTemplate, "datasource.go", data)
	if genMethods {
		genFile(ExportTemplate, "export.go", data)
	}
}

//将模板注入数据生成完整的代码并进行格式化后写到当前目录下的desFileName文件
func genFile(tmpl, desFileName string, data interface{}) {
	f1, err := os.OpenFile(desFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f1.Close()
	t := template.Must(template.New(desFileName).Parse(tmpl))
	var buffer bytes.Buffer
	t.Execute(&buffer, data)

	formated, err := format.Source(buffer.Bytes())
	if err != nil {
		log.Fatal("formate ", desFileName, " failed, err is "+err.Error(), "\n ori code is : \n", buffer.String())
	}
	f1.Write(formated)
}

func prepareData(connString, packageName string, tables []string) TemplateContext {
	if len(tables) == 0 {
		tables = GetTables()
	}
	context := TemplateContext{PackageName: packageName, ConnStr: connString, DriverName: driverName}
	context.Packages = []string{"database/sql"}
	for _, v := range tables {
		context.Types = append(context.Types, GenTypeFromTable(v))
	}

	return context

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

func GenTypeFromTable(tableName string) SingleType {
	rows, err := db.Query("desc " + tableName)
	if err != nil {
		log.Fatal("failed to query table %s, err is: %v", tableName, err)
	}
	defer rows.Close()
	var table []MetaCol
	var constVars []ConstVar
	camel := common.ChangeNameToCamel(tableName, "_")
	for rows.Next() {
		col := MetaCol{}
		err := rows.Scan(&col.Field, &col.Type, &col.Null, &col.Key, &col.Default, &col.Extra)
		if err != nil {
			log.Fatal(err)
		}
		col.GoType = mapFromSqlType(col.Type, col.Null)
		col.GoField = common.ChangeNameToCamel(col.Field, "_")
		constVars = ifEnum(constVars, col.Type, camel, col.GoField)
		table = append(table, col)
	}
	return SingleType{TableName: tableName, Camel: camel, Consts: constVars, Columns: table}
}

/*func genText(tableName string, cols []MetaCol) string {
	var buffer bytes.Buffer
	//buffer.WriteString("package entity\n\n")

	buffer.WriteString("type " + camel + " struct {")
	constVar := ""
	metaSingleton := ""
	for i, _ := range cols {
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
}*/

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

func ifEnum(vars []ConstVar, sqlType, typeName, fieldName string) []ConstVar {
	if strings.Contains(sqlType, "enum") {
		tmp := sqlType[strings.Index(sqlType, "('")+2 : strings.LastIndex(sqlType, "')")]
		t := strings.Split(tmp, "','")
		for i := 0; i < len(t); i++ {
			des := strings.Replace(t[i], "-", "_", -1) //变量命名不可以有中划线, 会被认为是减号
			vars = append(vars, ConstVar{StructName: typeName, FieldName: fieldName, Val: des})
		}
		//fmt.Println(ret)
	}
	return vars
}

func (v *MetaCol) String() string {
	return fmt.Sprintf("\n    %s %s `json:%s`", v.GoField, v.GoType, v.Field)
}
