package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/firmianavan/go-type-generator/mysql"
	"go/format"
	"log"
	"os"
	"strings"
)

//default value
var dbType string = "mysql"
var connStr string = "van:123456@tcp(127.0.0.1:3306)/wmp"

func main() {
	driver := flag.String("driver", dbType, "driver name, such as '-d mysql'")
	conn := flag.String("c", connStr, "connect str, should be matched with driver, a mysql conn str:'user:pwd@tcp(127.0.0.1:3306)/db'")
	table := flag.String("table", "", "tables to be generate code by, seperate by ',', if omit, will use all the tables in current db")
	withMethod := flag.Bool("withmethod", false, "if generate exported methods , default is no, that prevents overwriting you modify on those methods. when you first generate your code, you should add this flag")
	cur, _ := os.Getwd()
	directory := flag.String("d", cur, "your project root, if not set , current path will be used. The tool will generate code in entity package and test code in test package")

	flag.Parse()
	/*flag.Usage = func() {
		fmt.Println("a simple tool to generate code from database model:")
		fmt.Println("\t-d which driver you are using, default is mysql")
		fmt.Println("\t-c connect string to destination db, should be matched with driver, a mysql conn str look like: user:pwd@tcp(127.0.0.1:3306)/db")
		fmt.Println("\t-table tables to be generate code by, seperate by ',', if omit, all the tables in current db will be used")
	}*/
	var ret bytes.Buffer
	var code string
	if *driver == "mysql" {
		mysql.Construct(*conn)
		defer mysql.Close()
		var tables []string
		if *table != "" {
			tables = strings.Split(*table, ",")
		} else {
			tables = mysql.GetTables()
		}
		for _, t := range tables {
			tmp := mysql.GenTypeFromTable(t)
			ret.WriteString(tmp)
		}
		code = ret.String()

	} else {
		log.Fatal("unsupported driver")
	}

	if !exist(*directory) {
		log.Fatal(*directory + " is not exist")
	}
	err := os.Chdir(*directory)
	if err != nil {
		log.Fatal(err)
	}

	//包名
	packageName := "entity"

	if !exist(packageName) {
		if err := os.Mkdir(packageName, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	os.Chdir(packageName)
	f1, err := os.OpenFile("entity.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer f1.Close()
	fmt.Fprintln(f1, "//this file is generated by go-type-generater, you should not modify it.\n//any change will be lost after rerun go-type-generator\npackage "+packageName)
	fmt.Fprintln(f1, "import (\n    \"database/sql\"")
	if strings.Contains(code, "time.Time") {
		fmt.Fprintln(f1, "    \"time\"")
	}
	fmt.Fprintln(f1, ")")
	fmt.Fprintln(f1, mysql.InterfaceDefination)
	formated, err := format.Source([]byte(code))
	if err != nil {
		log.Fatal("failed to format code, error is:" + err.Error() + "\n origin code:\n" + code)
	}
	fmt.Fprintln(f1, string(formated))

	//datasource.go
	ds, err := os.OpenFile("datasource.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	defer ds.Close()
	ds.WriteString("package " + packageName + "\n")
	ds.WriteString(fmt.Sprintf(mysql.DataSource, *conn))

	//exports.go
	if *withMethod {
		f2, err := os.OpenFile("exports.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		defer f2.Close()
		f2.WriteString("package " + packageName + "\n")
		f2.WriteString("import (\n    \"database/sql\"\n    \"errors\"\n)\n")
		content, err := format.Source([]byte(mysql.ExportedMethods))
		f2.Write(content)
	}
}
func exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
