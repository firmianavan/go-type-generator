package tmpl

var EntityTemplate string = `
//this file is generated by go-type-generater, you should not modify it.
//any change will be lost after rerun go-type-generator
package {{.PackageName}}
import ({{range .Packages}}
    "{{.}}"{{end}}
)

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
{{range .Types}}
{{$tableName := .TableName }}
type {{.Camel}} struct {
    {{range .Columns}}{{.GoField}} {{.GoType}} ` + "`json:\"{{.Field}}\" bson:\"{{.Field}}\"`" + `
    {{end}}
    }

    {{range .Consts}}const Const_{{.TypeName}}_{{.FieldName}}_{{.Val}} string = "{{.Val}}"
    {{end}}

    {{range .Columns}}var {{$tableName}}{{.GoField}} ColumnMata = ColumnMata{ Field:"{{.Field}}",Type:"{{.Type}}",Null:"{{.Null}}",Key:"{{.Key}}",Default:"{{.Default}}",Extra:"{{.Extra}}",GoType:"{{.GoType}}",GoField:"{{.GoField}}"}
    {{end }}
    
    func ({{.TableName}} *{{.Camel}}) RowMap()(tableName string, colMap map[string]Column) {
        colMap = map[string]Column{
        {{range .Columns}}"{{.Field}}": Column{Meta : &{{$tableName}}{{.GoField}}, V : &{{$tableName}}.{{.GoField}}},
        {{end}}
        }
    return "{{.TableName}}",colMap
    }


{{end}}
`
