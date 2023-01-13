package dba

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// admin functions

type Admin struct {
	conn     *sql.DB
	Database string
	UserList []User
}

type User struct {
	Username string
	Password string
	Host     string
	Grants   []Grant
}

func (u *User) AddGrant(g Grant) {
	u.Grants = append(u.Grants, g)
}

func (u *User) ToString() string {
	var grants []string
	for _, g := range u.Grants {
		grants = append(grants, g.Raw)
	}
	return fmt.Sprintf("user:%s, host:%s \n grants: %s\n", u.Username, u.Host, strings.Join(grants, "\n"))
}

type Grant struct {
	Raw string
}

func (a *Admin) SetConn(conn *sql.DB) {
	a.conn = conn
}

func (a *Admin) Close() {
	if a.conn != nil {
		a.conn.Close()
	}
}

func (a *Admin) CreateDB(name string) error {
	query := fmt.Sprintf(`create schema %s collate utf8mb4_unicode_ci`, name)
	results, err := a.conn.Exec(query)
	if err != nil {
		return err
	}
	rowCnt, _ := results.RowsAffected()
	log.Println("rows affected", rowCnt)
	return nil
}

func (a *Admin) AddUser(user User) error {
	query := fmt.Sprintf(`create user '%s'@'%s' identified by '%s'`, user.Username, user.Host, user.Password)
	results, err := a.conn.Exec(query)
	if err != nil {
		return err
	}
	rowCnt, _ := results.RowsAffected()
	log.Println("rows affected", rowCnt)
	return nil
}

func (a *Admin) AddGrant(g Grant) error {
	results, err := a.conn.Exec(g.Raw)
	if err != nil {
		return err
	}
	rowCnt, _ := results.RowsAffected()
	log.Println("rows affected", rowCnt)
	return nil
}

func (a *Admin) GenPassword() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUZVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789" + "#!@^&*+")
	var b strings.Builder
	for i := 0; i < 30; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func UseDB(conn *sql.DB, db string) {
	conn.Exec("use " + db)
}

// add users
func GetTables(conn *sql.DB) []string {
	var tables []string
	results, err := conn.Query("show tables")
	if err != nil {
		log.Println(err)
		return tables
	}
	defer results.Close()
	for results.Next() {
		var table string
		err = results.Scan(&table)
		if err != nil {
			log.Println(err)
		}
		tables = append(tables, table)
	}
	return tables
}

// add users
func GetDatabase(conn *sql.DB) string {
	var database string

	results, err := conn.Query("select database()")
	if err != nil {
		log.Println(err)
		return database
	}
	defer results.Close()
	for results.Next() {
		err = results.Scan(&database)
		if err != nil {
			log.Println(err)
		}
	}
	return database
}

// add users
func GetDatabases(conn *sql.DB) []string {
	var databases []string

	results, err := conn.Query("show databases")
	if err != nil {
		log.Println(err)
		return databases
	}
	defer results.Close()
	for results.Next() {
		var database string
		err = results.Scan(&database)
		if err != nil {
			log.Println(err)
		}
		databases = append(databases, database)
	}
	return databases
}

func (a *Admin) GetUsers() []User {
	var users []User

	results, err := a.conn.Query("select user,host from mysql.user")
	if err != nil {
		log.Fatal(err)
	}

	defer results.Close()

	for results.Next() {
		var user User
		err = results.Scan(&user.Username, &user.Host)
		if err != nil {
			break
		}

		// fmt.Println(user.Username, user.Host)
		users = append(users, user)
	}
	return users
}

// list users
func (a *Admin) GetUsersWithGrants() []User {
	users := a.GetUsers()

	for _, u := range users {
		query := fmt.Sprintf("show grants for '%s'@'%s';", u.Username, u.Host)
		fmt.Println(query)
		grantResults, e := a.conn.Query(query)
		if e != nil {
			grantResults.Close()
			log.Fatal(e)
		}
		var err error
		for grantResults.Next() {
			var grant Grant
			err = grantResults.Scan(&grant.Raw)
			if err != nil {
				fmt.Println(err)
			}
			// fmt.Println(grant.Raw)
			u.AddGrant(grant)
		}
		grantResults.Close()
	}
	return users
}

// list grants
func (a *Admin) ListGrants(user User) {
	query := fmt.Sprintf(`show grants for '{}'@'{}'`,
		user.Username, user.Host)
	fmt.Sprintf(query)
	fmt.Println("TODO")
}

// add users
func GetTime(conn *sql.DB) string {
	var timeStr string
	results, err := conn.Query("select now()")
	if err != nil {
		log.Println(err)
		return timeStr
	}
	defer results.Close()
	for results.Next() {
		err = results.Scan(&timeStr)
		if err != nil {
			log.Println(err)
		}
	}
	return timeStr
}

// RemoveGrants

func Run(conn *sql.DB, sql string) error {
	sqlLower := strings.ToLower(sql)

	hasRowOutput := false
	if strings.Contains(sqlLower, "select") || strings.Contains(sqlLower, "describe") {
		hasRowOutput = true
	}

	if !hasRowOutput {
		results, err := conn.Exec(sql)
		if err != nil {
			return err
		}
		rowCnt, _ := results.RowsAffected()
		log.Println("rows affected", rowCnt)
	} else {
		rows, _ := conn.Query(sql)
		columns, _ := rows.Columns()
		rowCnt := len(columns)
		values := make([]interface{}, rowCnt)
		valuePtrs := make([]interface{}, rowCnt)

		for _, col := range columns {
			fmt.Printf("%-40s", col)
		}
		fmt.Println()

		defer rows.Close()

		for rows.Next() {
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			rows.Scan(valuePtrs...)

			for i := range columns {
				val := values[i]

				b, ok := val.([]byte)
				var v interface{}
				if ok {
					v = string(b)
				} else {
					v = val
				}

				fmt.Printf("%-40v", v)
			}
			fmt.Println()
		}
	}
	return nil
}
