# go-type-generator
从数据库表生成对应的go类型,对应的result maper以及一些常用的数据库操作,作为ORM外的另一个选择, 目前仅支持mysql.(generate go types and mappers and some common db operations from tables of database, currentlt on mysql is supported)

### install
go install github.com/firmianavan/go-type-generator

### Usage
this is a command-line tool,if env `GOPATH` is set and `$GOPATH/bin` is add to path, type `$ go-type-generator --help` you will see:
```
    -c string
        connect str, should be matched with driver, a mysql conn str:'user:pwd@tcp(127.0.0.1:3306)/db' (default "van:123456@tcp(127.0.0.1:3306)/wmp")
    -d string
        your project root, if not set , current path will be used. The tool will generate code in entity package and test code in test package (default "/home/van/archive/workspace/go/src/github.com/firmianavan/go-type-generator")
    -driver string
        driver name, such as '-d mysql' (default "mysql")
    -table string
        tables to be generate code by, seperate by ',', if omit, will use all the tables in current db
    -withmethod
        if generate exported methods , default is no, that prevents overwriting you modify on those methods. when you first generate your code, you should add this flag
```
