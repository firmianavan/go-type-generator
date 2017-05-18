package mysql

var ExportTemplate string =`
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

var EmptyResultError error = errors.New("no rows returned")

//select where key=v, if return more than one row, only the first mapped into param mapper
func QueryUnique(mapper RowMaper, key string, v interface{}) error {
    db := GetDB()
    table, _ := mapper.RowMap()
    rows, err := db.Query("select * from "+table+" where  "+key+" =?", v)
    if err != nil {
        return err
    }
    if rows.Next() {
        return MapRow(rows, mapper)
    } else {
        return EmptyResultError
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