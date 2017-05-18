package main

import (
	"flag"
	"github.com/firmianavan/go-type-generator/mysql"
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

	//准备参数
	packageName := "entity"
	prepareDirectory(*directory, packageName)
	var tables []string
	if *table != "" {
		tables = strings.Split(*table, ",")
	}

	if *driver == "mysql" {
		mysql.GenCode(*conn, packageName, tables, *withMethod)
	} else { //目前仅支持mysql
		log.Fatal("unsupported driver")
	}

}

//验证并切换到生产代码的路径, 即 root/packageName
func prepareDirectory(root, packageName string) {
	if !exist(root) {
		log.Fatal(root + " is not exist")
	}
	err := os.Chdir(root)
	if err != nil {
		log.Fatal(err)
	}

	if !exist(packageName) {
		if err := os.Mkdir(packageName, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	os.Chdir(packageName)
}

func exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
