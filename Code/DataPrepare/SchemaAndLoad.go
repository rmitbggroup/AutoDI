package DataPrepare

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"io/ioutil"
	"strings"
)

// Suppose the ";" separation.
func SchemaDefine(sqlPath string) {
	bytes, err := ioutil.ReadFile(sqlPath)
	if err != nil {
		return
	}
	sqlss := string(bytes)
	sqlsa := strings.Split(sqlss, ";")
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/test")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	for _, query := range sqlsa {
		if query != "" {
			query = strings.TrimSpace(query)+";"
			_, err := db.Exec(query)
			if err != nil {
				panic(err.Error())
				return
			}
			defer db.Close()
		}

	}
}

func LoadDataFromDir(dirPath, tableName string) {
	files, _ := ioutil.ReadDir(dirPath)
	for _, f := range files {
		if !(f.Name() == ".DS_Store") {
			println(f.Name())
			LoadDataFromFile(dirPath+f.Name(), tableName)
		}
	}
}

func LoadDataFromFile(dataPath, tableName string) {
	mysql.RegisterLocalFile(dataPath)
	query := fmt.Sprintf("load data local infile '%s' into table %s fields terminated by ',' enclosed by '\"' lines terminated by '\\n';", dataPath, tableName)
	println(query)
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/test")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	_, err = db.Exec(query)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
}
