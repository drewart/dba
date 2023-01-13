package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/drewart/dba"
	"github.com/drewart/dba/cli"
	"github.com/drewart/dba/config"
	_ "github.com/go-sql-driver/mysql"

	//"subsplash.io/drew/redeem"
	//"subsplash.io/drew/wps"

	"github.com/c-bata/go-prompt"
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

var (
	history     []string
	picker      dba.Picker
	sqlCommands = []string{"select", "create", "update", "insert", "grant", "describe"}
	reDigit, _  = regexp.Compile(`^\d+$`)
	user        *dba.User
)

func (u *DBUser) Display() {
	fmt.Printf("user:%s host:%s", u.User, u.Host)
}

func help() {
	usage := `
Help:
  quit   - exit
  list   - list databases
  servers - server list
  databases - list databases
  tables - list tables
  users <filter> - lists user and perms
  my - opens mysql client in new terminal
  my2 - opens mysql in shell client 
  cli - opens mycli with current config
  cli2 - opens in shell mycli with current config

  <digit> - connection/server number
`

	fmt.Print(usage)
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	word := d.GetWordBeforeCursor()
	s := []prompt.Suggest{}
	if word == "" || !strings.Contains(d.Text, " ") {
		if picker.CurrentDB == nil || len(picker.MultiCurrentDB) == 0 {
			s = []prompt.Suggest{
				{Text: "list|servers", Description: "lists servers to connect to"},
				{Text: "pick <digit>|<digit>", Description: "pick <digit> switches server connections"},
				{Text: "mpick <digit,digit>|<digit,digit,...>", Description: "mpick <digit,digit,...> switches server connections"},
				{Text: "find-user", Description: "find user "},
			}
		}
		if strings.Contains(word, "create-user") {
			s3 := prompt.Suggest{Text: "<username>", Description: "username to create i.e. john.doe"}
			s = append(s, s3)
		} else if picker.CurrentDB != nil {
			s2 := []prompt.Suggest{
				{Text: "dbs", Description: "list databases for current connection"},
				{Text: "databases", Description: "list databases for current connection"},
				{Text: "tables", Description: "lists tables for current database"},
				{Text: "use", Description: "use <database> switches database for connection"},
				{Text: "init", Description: "init <project name> - generates database and service user"},
				{Text: "my", Description: "starts mysql client in new terminal"},
				{Text: "my2", Description: "starts mysql client in shell"},
				{Text: "cli", Description: "starts mycli in new termainal"},
				{Text: "cli2", Description: "starts mycli in shell"},
				{Text: "mcli", Description: "starts mycli in new termainal xpanes with many connetion using mycli"},
				{Text: "cur", Description: "creates readonly user"},
				{Text: "cu", Description: "creates user with read/write"},
			}
			s = append(s, s2...)
		} else if len(picker.MultiCurrentDB) > 0 {
			s2 := []prompt.Suggest{
				{Text: "mcli", Description: "starts mycli in new termainal xpanes with many connetion using mycli"},
			}
			s = append(s, s2...)

		}
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// createServiceUser user
func createServiceUser(admin *dba.Admin, user *dba.User) error {
	defaultUserHost := "172.20.%.%"
	if user.Host == "" {
		user.Host = defaultUserHost
	}

	grantFormat := "grant insert,update,select,execute ON `%s`.* TO '%s'@'%s'"

	if strings.Contains(user.Username, "_ro") {
		grantFormat = "grant select ON `%s`.* TO '%s'@'%s'"
	}

	grantStr := fmt.Sprintf(grantFormat, admin.Database, user.Username, user.Host)

	if user.Password == "" {
		user.Password = admin.GenPassword()
	}

	var grants []dba.Grant

	grant := dba.Grant{Raw: grantStr}

	grants = append(grants, grant)
	user.Grants = grants

	err := admin.AddUser(*user)
	if err != nil {
		return err
	}
	err = admin.AddGrant(grant)
	if err != nil {
		return err
	}
	return nil
}

func createUser(admin *dba.Admin, user *dba.User, readOnly bool) error {
	defaultUserHost := "172.20.%.%"
	if user.Host == "" {
		user.Host = defaultUserHost
	}

	grantFormat := "grant insert,update,select,execute ON `%%`.* TO '%s'@'%s'"

	if readOnly {
		grantFormat = "grant select ON `%%`.* TO '%s'@'%s'"
	}

	grantStr := fmt.Sprintf(grantFormat, user.Username, user.Host)

	if user.Password == "" {
		user.Password = admin.GenPassword()
	}

	var grants []dba.Grant

	grant := dba.Grant{Raw: grantStr}

	grants = append(grants, grant)
	user.Grants = grants

	err := admin.AddUser(*user)
	if err != nil {
		return err
	}
	err = admin.AddGrant(grant)
	if err != nil {
		return err
	}
	return nil
}

func listUser(picker dba.Picker, withGrants bool, filter string) {
	admin := dba.Admin{Database: picker.CurrentDB.Database}
	admin.SetConn(picker.GetCurrentConn())
	users := admin.GetUsersWithGrants()

	fmt.Printf("%s50s\t%s30s\n", "username", "host")
	for _, u := range users {

		if filter != "" && !strings.Contains(u.Username, filter) {
			continue
		}

		fmt.Printf("%50s\t%30s\n", u.Username, u.Host)
		if withGrants {
			for _, g := range u.Grants {
				fmt.Print("\t", g.Raw)
			}
		}
		fmt.Println()
	}
}

func listGrants(picker dba.Picker, filter string) {
	admin := dba.Admin{Database: picker.CurrentDB.Database}
	admin.SetConn(picker.GetCurrentConn())
	users := admin.GetUsersWithGrants()

	for _, u := range users {
		for _, g := range u.Grants {
			if filter != "" && !strings.Contains(g.Raw, filter) {
				continue
			}
			fmt.Println(g.Raw)
		}
	}
}

func pingAll() {
	for _, cfg := range picker.AllConfigs {
		go func(conf config.Config) {
			dsn := conf.ToDSN()
			conn, err := sql.Open("mysql", dsn)
			if err != nil {
				fmt.Println(err)
			}
			timeStr := dba.GetTime(conn)
			fmt.Println(conf.FilePath, conf.Host, timeStr)
			conn.Close()
		}(cfg)
	}
}

// initDB set up project database.
func initDB(picker dba.Picker, project string, env string) error {
	databaseName := project
	host := picker.CurrentDB.Host
	// replace project dash with underscore
	if strings.Contains(project, "-") {
		databaseName = strings.Replace(project, "-", "_", -1)
	}

	// TODO move prefix
	user := dba.User{
		Username: "sa_" + strings.ToLower(databaseName),
		Host:     "10.0.%",
	}
	userRO := dba.User{
		Username: "sa_" + strings.ToLower(databaseName) + "_ro",
		Host:     "10.0.%",
	}

	admin := dba.Admin{Database: databaseName}
	admin.SetConn(picker.GetCurrentConn())

	databases := dba.GetDatabases(picker.GetCurrentConn())

	dbExists := false
	for _, name := range databases {
		if name == databaseName {
			dbExists = true
		}
	}

	// if database hasn't been created
	if !dbExists {
		err := admin.CreateDB(databaseName)
		if err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		fmt.Println("Database ", databaseName, " exists skipping creation")
	}

	users := admin.GetUsersWithGrants()

	hasUser := false
	hasUserRO := false
	for _, u := range users {
		if u.Username == user.Username && u.Host == user.Host {
			hasUser = true
		}
		if u.Username == userRO.Username && u.Host == userRO.Host {
			hasUserRO = true
		}
	}

	// assume if non readonly user exists this has been setup
	if hasUser && hasUserRO {
		fmt.Println("user and ro user exists skipping")
		return nil
	}

	// TODO write to json file
	fmt.Println("Wrote user to json file: ", host, user)

	return nil
}

// createDBUser creates a database user
func createDBUser(picker dba.Picker, username string, env string, readOnly bool) error {
	fmt.Println("Created user", username)

	admin := dba.Admin{Database: picker.CurrentDB.Database}
	admin.SetConn(picker.GetCurrentConn())

	user := dba.User{
		Username: username,
		Host:     "10.0.%",
	}

	hasUser := false
	users := admin.GetUsers()
	for _, u := range users {
		if u.Username == user.Username && u.Host == user.Host {
			hasUser = true
		}
	}

	// assume if non readonly user exists this has been setup
	if hasUser {
		fmt.Printf("user %s exists skipping\n", username)
		return nil
	}

	err := createUser(&admin, &user, readOnly)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("User created:", user.Username)
	fmt.Println("Password:", user.Password)

	// TODO setup config for password management tool
	openbrowser("https://drewart.com/password/")

	return nil
}

// switch
func switchConn(picker *dba.Picker, env *string, connNum string) {
	if i, err := strconv.Atoi(connNum); err == nil {
		picker.SetCurrent(i)
	} else {
		fmt.Println(err)
	}
	// figure out env
	host := strings.ToLower(picker.CurrentDB.Host)
	if strings.Contains(host, "dev") {
		*env = "dev"
	} else if strings.Contains(host, "stage") {
		*env = "stage"
	} else {
		*env = "prod"
	}
}

func findDatabaseUser(in UserFindInput) UserFindResult {
	var result UserFindResult
	result.DBIndex = in.DBIndex
	result.Found = false
	conn, err := sql.Open("mysql", in.DSN)
	if err != nil {
		result.Err = err
		return result
	}
	admin := dba.Admin{}
	admin.SetConn(conn)
	users := admin.GetUsers()
	// log.Printf("found %d users\n", len(users))
	for _, u := range users {
		if u.Username == in.Username {
			result.Found = true
		}
	}
	defer admin.Close()
	return result
}

type UserFindResult struct {
	DBIndex int
	Found   bool
	Err     error
}

type UserFindInput struct {
	DSN      string
	DBIndex  int
	Username string
}

// findUser with goroutine workers
func findUser(picker *dba.Picker, username string) []UserFindResult {
	var userResults []UserFindResult
	jobs := make(chan UserFindInput, len(picker.AllConfigs))
	results := make(chan UserFindResult, len(picker.AllConfigs))

	// spin up 2 goroutines
	go findUserWorker(jobs, results)
	go findUserWorker(jobs, results)

	for i, cfg := range picker.AllConfigs {
		jobs <- UserFindInput{cfg.ToDSN(), i, username}
	}
	close(jobs)

	for i := 0; i < len(picker.AllConfigs); i++ {
		r := <-results

		fmt.Println(picker.AllConfigs[r.DBIndex].Host, " Found:", r.Found)
		userResults = append(userResults, r)
	}
	return userResults
}

func findUserWorker(jobs <-chan UserFindInput, results chan<- UserFindResult) {
	for j := range jobs {
		results <- findDatabaseUser(j)
	}
}

type Exit int

func exit(_ *prompt.Buffer) {
	panic(Exit(0))
}

func handleExit() {
	switch v := recover().(type) {
	case nil:
		return
	case Exit:
		os.Exit(int(v))
	default:
		fmt.Println(v)
		fmt.Println(string(debug.Stack()))
	}
}

func executer(in string) {
	isQuery := false
	command := strings.TrimRight(in, ";\n")
	history = append(history, command)

	cmdParts := strings.SplitN(in, " ", 2)
	cmd := cmdParts[0]
	cmd = strings.ToLower(cmd)
	args := ""
	if len(cmdParts) > 1 {
		args = cmdParts[1]
	}
	// fmt.Println("cmd:", cmd, "args:", args)

	switch cmd {
	case "exit":
		fallthrough
	case "quit":
		os.Exit(0)
	case "servers":
		fallthrough
	case "list":
		if args != "" {
			picker.ListFilter(args)
		} else {
			picker.ListDatabases()
		}
	case "tables":
		if picker.IsCurrentSet() {
			tables := dba.GetTables(picker.GetCurrentConn())
			for _, dbName := range tables {
				fmt.Println(dbName)
			}
		}
	case "dbs":
		fallthrough
	case "databases":
		if picker.IsCurrentSet() {
			databases := dba.GetDatabases(picker.GetCurrentConn())
			for _, dbName := range databases {
				fmt.Println(dbName)
			}
		}
	case "init":
		fmt.Println("init")
		if args != "" {
			project := args
			err := initDB(picker, project, ShellState.Env)
			if err != nil {
				log.Println(err)
			}

		} else {
			fmt.Println("init <database name>")
		}
	case "info":
		if picker.CurrentDB != nil {
			fmt.Println("current: ", picker.CurrentDB.Host, picker.CurrentDB.FilePath)
		}

	case "current":
		// list current database
		fmt.Println(picker.CurrentDB.Host, picker.CurrentDB.Database)
	case "search":
		// search
		fmt.Println("search", args)
	case "pick":
		if args != "" {
			switchConn(&picker, &ShellState.Env, args)
		} else {
			fmt.Println("pick <int> - connection number for current list")
		}
	case "mpick":
		if args != "" {
			multiPickConn(&picker, &ShellState.Env, args)
		} else {
			fmt.Println("mpick <int,int,...>")
		}
	case "clear":
		picker.ClearMultiCurrent()
		picker.CurrentDB = nil
	case "cli":
		if picker.IsCurrentSet() {
			cli.OpenMyCLI(picker.CurrentDB, "mycli")
		}
	case "mcli":
		if len(picker.MultiCurrentDB) > 0 {
			var connections []config.Config
			for _, index := range picker.MultiCurrentDB {
				connections = append(connections, picker.AllConfigs[index])
			}

			cli.OpenMyCLIMany(connections, "mycli")
		}
	case "cli2":
		if picker.IsCurrentSet() {
			cli.OpenCLIInShell(picker.CurrentDB, "mycli")
		}
	case "my":
		if picker.IsCurrentSet() {
			cli.OpenMyCLI(picker.CurrentDB, "mysql")
		}
	case "my2":
		if picker.IsCurrentSet() {
			cli.OpenCLIInShell(picker.CurrentDB, "mysql")
		}

	case "users":
		listGrants := true
		filter := ""
		if args != "" {
			filter = args
		}
		listUser(picker, listGrants, filter)
	case "find":
		fallthrough
	case "find-user":
		if args == "" {
			fmt.Println("find-user <username")
		} else {
			var notFound []int
			results := findUser(&picker, args)
			fmt.Println("found")
			for _, r := range results {
				if r.Found {
					fmt.Println(r.DBIndex, picker.AllConfigs[r.DBIndex].Host)
				} else {
					notFound = append(notFound, r.DBIndex)
				}
			}
			fmt.Println("not found")
			for _, i := range notFound {
				fmt.Println(i, picker.AllConfigs[i].Host)
			}
		}
	case "cur":
		if picker.IsCurrentSet() {
			if args == "" {
				fmt.Println("cur <username>")
			} else {
				readOnly := true
				createDBUser(picker, args, ShellState.Env, readOnly)
			}
		}
	case "cu":
		if picker.IsCurrentSet() {
			if args == "" {
				fmt.Println("cu <username>")
			} else {
				fmt.Println("cu <username>")

				readOnly := false
				createDBUser(picker, args, ShellState.Env, readOnly)
			}
		}
	case "cup":
		fmt.Println("create user with password")
	case "curp":
		fmt.Println("create user readonly with password")
	case "fp":
		fallthrough
	case "find-pass":
		if args == "" {
			fmt.Println("find-pass <term>")
		} else {
			//findPassword(args)
			fmt.Print("TODO find password tool")
		}
	case "clear-pass":
		user = nil
	case "grants":
		filter := ""
		if args != "" {
			filter = args
		}
		listGrants(picker, filter)
	case "help":
		help()
	case "ping-all":
		pingAll()
	case "use":
		// set current database
		if args != "" {
			picker.CurrentDB.Database = args
			dba.UseDB(picker.GetCurrentConn(), args)
		}
	case "env":
		ShellState.Env = args
	default:
		// fmt.Println("in default")
		for _, sqlVerb := range sqlCommands {
			if strings.Contains(cmd, sqlVerb) {
				isQuery = true
			}
		}

		if isQuery && picker.IsCurrentSet() {

			err := dba.Run(picker.GetCurrentConn(), command)
			if err != nil {
				fmt.Println(err)
			}

		} else if reDigit.MatchString(cmd) {
			switchConn(&picker, &ShellState.Env, cmd)
		} else {
			fmt.Println("unknown:", cmd)
			help()
		}
	}
}

func multiPickConn(picker *dba.Picker, env *string, args string) {
	var connections []int
	parts := strings.Split(args, ",")
	if len(parts) > 0 {
		for i := 0; i < len(parts); i++ {
			num, err := strconv.Atoi(parts[i])
			if err != nil {
				log.Println(err)
				continue
			}
			connections = append(connections, num)
		}
		if len(connections) > 0 {
			picker.CurrentDB = nil
			picker.SetMultiCurrent(connections)
		} else {
			fmt.Println("error unable to set connections")
		}
	}
}

var ShellState struct {
	Database string
	Env      string
}

// returns prompt text
func promptTxt() (string, bool) {
	promptDBText := "(:)"
	env := ShellState.Env
	if len(picker.MultiCurrentDB) > 0 {
		dilim := ""
		promptDBText = "("
		for _, index := range picker.MultiCurrentDB {
			promptDBText += fmt.Sprintf("%s%d", dilim, index)
			dilim = ","
		}
		promptDBText += ")"
		env = "many"
	} else if picker.IsCurrentSet() && picker.CurrentDB.Host != "" {
		promptDBText = fmt.Sprintf("(%s:%s)", picker.CurrentDB.Host, picker.CurrentDB.Database)
	}
	passSetStr := ""
	if user != nil {
		passSetStr = "$p"
	}
	promptTxt := fmt.Sprintf("dba %s:%s %s > ", env, promptDBText, passSetStr)
	return promptTxt, true
}

func shellNew() {
	defer handleExit()

	p := prompt.New(executer, completer, prompt.OptionHistory(history), prompt.OptionLivePrefix(promptTxt), prompt.OptionPrefixTextColor(prompt.Blue))

	p.Run()
}

func shell() {
	promptDBName := "(:)"
	env := ""
	// reDigit, _ := regexp.Compile(`^\d+$`)
	promptTxt := ">"
	reader := bufio.NewReader(os.Stdin)
	for {
		if len(picker.MultiCurrentDB) > 0 {
			dilim := ""
			for _, index := range picker.MultiCurrentDB {
				promptDBName += fmt.Sprintf("%s(%s:%s)", dilim, picker.AllConfigs[index].Host, picker.AllConfigs[index].Database)
				env = "many"
				dilim = ",\n"
			}
		} else if picker.IsCurrentSet() && picker.CurrentDB.Host != "" {
			promptDBName = fmt.Sprintf("(%s:%s)", picker.CurrentDB.Host, picker.CurrentDB.Database)
		}
		passSetStr := ""
		if user != nil {
			passSetStr = "$p"
		}

		var command string
		promptTxt = fmt.Sprintf("dba %s:%s %s > ", env, promptDBName, passSetStr)
		println(promptTxt)

		/*i, err := fmt.Scan(&command)
		if err != nil {
			fmt.Println(err)
			return
		}*/
		command, _ = reader.ReadString('\n')
		// command = prompt.Input(promptTxt, completer)

		executer(command)
	}
}

func main() {
	picker.FindConfigs()
	picker.ListDatabases()
	defer picker.Close()

	// shell()
	shellNew()
}
