package tmpl

var DataSourceTemplage string = `
package {{.PackageName}}
import (
    "database/sql"
    _ "{{.DriverName}}"
    "log"
    "sync"
)

var db *sql.DB
var once sync.Once
var connStr string = "{{.ConnStr}}?parseTime=true&loc=UTC"

func GetDB() *sql.DB {
    once.Do(func() {
        var err error
        db, err = sql.Open("mysql", connStr)
        if err != nil {
            log.Fatal("program exit: failed to connect the database %!s(MISSING)", connStr)
        }
    })
    return db
}

func Release() {
    if db != nil {
        db.Close()
    }
}
`
