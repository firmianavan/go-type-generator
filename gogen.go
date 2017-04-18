package main

import (
	"flag"
	"fmt"
	"github.com/firmianavan/go-type-generator/mysql"
	"log"
	"strings"
)

//default value
var dbType string = "mysql"
var connStr string = "van:123456@tcp(127.0.0.1:3306)/wmp"

func main() {
	driver := flag.String("d", dbType, "driver name, such as '-d mysql'")
	conn := flag.String("c", connStr, "connect str, should be matched with driver, a mysql conn str:'user:pwd@tcp(127.0.0.1:3306)/db'")
	table := flag.String("table", "", "tables to be generate code by, seperate by ',', if omit, will use all the tables in current db")
	flag.Parse()
	flag.Usage = func() {
		fmt.Println("a simple tool to generate code from database model:")
		fmt.Println("\t-d which driver you are using, default is mysql")
		fmt.Println("\t-c connect string to destination db, should be matched with driver, a mysql conn str look like: user:pwd@tcp(127.0.0.1:3306)/db")
		fmt.Println("\t-table tables to be generate code by, seperate by ',', if omit, all the tables in current db will be used")
	}
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
			fmt.Println(tmp)
		}
	} else {
		log.Fatal("unsupported driver")
	}

}
