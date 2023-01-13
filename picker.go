package dba

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/drewart/dba/config"
	_ "github.com/go-sql-driver/mysql"
)

type Picker struct {
	conn           *sql.DB
	CurrentDB      *config.Config
	MultiCurrentDB []int
	AllConfigs     []config.Config
}

func (p *Picker) IsCurrentSet() bool {
	return p.CurrentDB != nil && p.CurrentDB.Host != ""
}

func (p *Picker) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}

func (p *Picker) GetCurrentConn() *sql.DB {
	var err error
	if p.CurrentDB.Host != "" && p.conn == nil {
		dsn := p.CurrentDB.ToDSN()
		p.conn, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Println(err)
			return p.conn
		}
	}

	return p.conn
}

func (p *Picker) SetCurrent(i int) {
	if p.conn != nil {
		p.conn.Close()
	}
	p.conn = nil
	p.CurrentDB = &p.AllConfigs[i]
}

func (p *Picker) SetMultiCurrent(indexes []int) {
	p.MultiCurrentDB = append(p.MultiCurrentDB, indexes...)
}

func (p *Picker) ClearMultiCurrent() {
	p.MultiCurrentDB = nil
}

// FilteredFiles returns file from a directory based on filter string i.e. .cnf
func ListFiles(dir string, filter string) ([]string, error) {
	var files []string

	dirList, err := ioutil.ReadDir(dir)
	if err != nil {
		return files, err
	}

	for _, file := range dirList {
		if file.IsDir() == false && strings.HasSuffix(file.Name(), filter) {
			filePath := path.Join(dir, file.Name())
			files = append(files, filePath)
		}
	}

	return files, nil
}

func (p *Picker) FindConfigs() {
	usr, _ := user.Current()

	files, err := ListFiles(usr.HomeDir, ".cnf")
	if err != nil {
		log.Fatal(err)
	}

	// get current directory .cnf files too
	currentDir, _ := os.Getwd()
	if currentDir != usr.HomeDir {
		currentDirFiles, err := ListFiles(currentDir, ".cnf")
		if err != nil {
			log.Fatal(err)
		}
		files = append(files, currentDirFiles...)
	}

	for _, file := range files {

		cnf, err := config.NewConfig(file)
		if err != nil {
			log.Println(err)
		}
		p.AllConfigs = append(p.AllConfigs, cnf)
	}
}

func (p *Picker) ListDatabases() {
	for i, cnf := range p.AllConfigs {
		fmt.Println(i, ":", cnf.Host, cnf.Database)
	}
}

func (p *Picker) ListFilter(filter string) {
	filterReadOnly := false
	if strings.Contains(filter, " -ro") {
		filter = strings.ReplaceAll(filter, " -ro", "")
		filterReadOnly = true
	}
	for i, cnf := range p.AllConfigs {
		if filterReadOnly && strings.Contains(cnf.Host, "-ro") || strings.Contains(cnf.Host, "-read") {
			continue
		}
		if strings.Contains(cnf.Host, filter) || strings.Contains(cnf.Database, filter) {
			fmt.Println(i, ":", cnf.Host, cnf.Database)
		}
	}
}
