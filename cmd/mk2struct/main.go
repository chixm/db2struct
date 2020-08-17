package main

import (
	"fmt"
	"os"
	_ "strconv"

	"github.com/chixm/db2struct"
	goopt "github.com/droundy/goopt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/howeyc/gopass"
)

var mariadbHost = os.Getenv("MYSQL_HOST")
var mariadbHostPassed = goopt.String([]string{"-H", "--host"}, "", "Host to check mariadb status of")
var mariadbPort = goopt.Int([]string{"--mysql_port"}, 3306, "Specify a port to connect to")
var mariadbDatabase = goopt.String([]string{"-d", "--database"}, "nil", "Database to for connection")
var mariadbPassword *string
var mariadbUser = goopt.String([]string{"-u", "--user"}, "user", "user to connect to database")

var packageName = goopt.String([]string{"--package"}, "", "name to set for package")
var gormAnnotation = goopt.Flag([]string{"--gorm"}, []string{}, "Add gorm annotations (tags)", "")
var dbAnnotation = goopt.Flag([]string{"--db"}, []string{}, "Add db annotations (tags)", "")
var gureguTypes = goopt.Flag([]string{"--guregu"}, []string{}, "Add guregu null types", "")

var targetFile = goopt.String([]string{"--target"}, "", "Save file path")

func init() {
	fmt.Println(`Initialize mk2struct`)
	goopt.OptArg([]string{"-p", "--password"}, "", "Mysql password", getMariadbPassword)
	// Setup goopts
	goopt.Description = func() string {
		return "Mariadb http Check"
	}
	goopt.Version = "0.0.2"
	goopt.Summary = "db2struct [-H] [-p] [-v] --package pkgName --struct structName --database databaseName -u user"
	//Parse options
	goopt.Parse(nil)
}

func main() {
	fmt.Println(`run mk2struct`)

	var writer *TargetWriter
	if targetFile != nil && *targetFile != "" {
		file, err := os.OpenFile(*targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Open File fail: " + err.Error())
			return
		}
		defer file.Close()
		writer = &TargetWriter{file: file}
	} else {
		writer = &TargetWriter{file: nil}
	}

	tables, err := db2struct.GetMySQLTableNames(*mariadbUser, *mariadbPassword, *mariadbHostPassed, *mariadbPort, *mariadbDatabase)
	if err != nil {
		fmt.Println(`Failed to get Table List.`)
		return
	}
	// Write package once
	writer.Write([]byte(fmt.Sprintf("package %s\n", *packageName)))

	// Write All Tables in Database to output
	for _, t := range tables {
		columnDataTypes, err := db2struct.GetColumnsFromMysqlTable(*mariadbUser, *mariadbPassword, *mariadbHostPassed, *mariadbPort, *mariadbDatabase, t)
		if err != nil {
			fmt.Println("Error in getting columns from information_schema: " + err.Error())
			return
		}
		struc, err := db2struct.GenerateWithoutPackage(*columnDataTypes, t, t, false, *gormAnnotation, *dbAnnotation, *gureguTypes)
		if err != nil {
			fmt.Println("Error in creating struct from json: " + err.Error())
			return
		}
		// 書き込み
		writer.Write(struc)

		writer.Write([]byte("\n"))
	}
}

func getMariadbPassword(password string) error {
	mariadbPassword = new(string)
	*mariadbPassword = password
	return nil
}

// TargetWriter 書き込み先のstruct
type TargetWriter struct {
	file *os.File
}

func (m *TargetWriter) Write(struc []byte) {
	if m.file != nil {
		_, err := m.file.WriteString(string(struc))
		if err != nil {
			fmt.Println("Save File fail: " + err.Error())
			return
		}
	} else {
		fmt.Println(string(struc))
	}
}
