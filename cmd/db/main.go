package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	db "github.com/drewart/dba"
	"github.com/drewart/dba/cli"
	"github.com/drewart/dba/config"
	_ "github.com/go-sql-driver/mysql"
)

// list connections
// pick connection
// auto complete example like docker/minikube
// perform database admin
// ldap intagrations to connect via preset ldap group to db mappings
// proxy queries
// mycli like tool?  auto complete
// shell like tool?
// collection of config list
// sort in json config
// need a way to load and store json config
// add config / add via glub

// https://github.com/jtblin/go-ldap-client  ldap

type DBUser struct {
	Host string `json:"host"`
	User string `json:"user"`
}

func (u *DBUser) Display() {
	fmt.Printf("user:%s host:%s", u.User, u.Host)
}

func help() {
	usage := `Command Help:
  findone <term0> <term1> ...  - finds top one config with terms
  find <term0> <term1> ... - finds db configs with terms
  list - list database configs
  exec - runs mycli fineone database
  dev|stage|prod -  runs exec to term=(dev|stage|prod) + terms 

Flags:
  -r  - include readonly connection
`
	fmt.Println(usage)
}

func main() {
	showReadOnly := flag.Bool("r", false, "show readonly")

	flag.Parse()

	var picker db.Picker

	picker.FindConfigs()

	command := ""
	var args []string
	if len(os.Args) > 1 {
		command = os.Args[1]
		args = os.Args[2:]
	}

	limit := -1
	run := false

	switch command {
	case "dev":
		fallthrough
	case "stage":
		fallthrough
	case "prod":
		run = true
		limit = 1
		args = append(args, command)
		fallthrough
	case "exec":
		run = true
		fallthrough
	case "findone":
		limit = 1
		fallthrough
	case "list":
		fallthrough
	case "find":
		var findList []config.Config
		for _, cnf := range picker.AllConfigs {
			// filter out readonly hosts
			if !*showReadOnly && (strings.Contains(cnf.Host, "-read") ||
				strings.Contains(cnf.Host, "-ro")) {
				continue
			}
			match := true
			// search for matching args in host
			for _, arg := range args {
				if !strings.Contains(cnf.Host, arg) {
					match = false
				}
			}
			if match {
				findList = append(findList, cnf)
			}
		}
		var exeConfig config.Config

		for i, cnf := range findList {

			exeConfig = cnf
			if !run {
				fmt.Println(cnf.FilePath)
			}
			if limit > 0 && i >= limit-1 {
				break
			}
		}
		if run && exeConfig.FilePath != "" {
			cli.MyCLI(&exeConfig)
		} else {
			fmt.Println("config not found")
		}
	case "help":
		help()
	default:
		help()
	}
}
