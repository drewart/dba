# dba 

a tools for working with databases in the shell

## dba

A shell tool for connecting to mysql databases and running tools/commands, also works with wps go tool to create wps entries when creating users.

### install

```bash
git clone git@github.com:drewart/dba.git
cd dba
go install ./cmd/dba/.

# install mycli mysql connection tool
pip3 install mycli
```

### shell commands

 - list : lists servers from $HOME/.*.cnf files and $(pwd)/.*.cnf
 - list dev : lists dev servers or servers with dev in the name
 - pick digit : picks server number for current context
 - mpick digit,digit : picks many server number for current context used with mcli*
 - digit : picks the server context 
 - database context commands
   - cli : (mac/linux gnome) opens new terminal with mycli
   - cli2 : opens mycli connection in dba shell
   - mcli : (mac/linux gnome) opens new terminal with mycli
   - my :  (mac/linux gnome) opens mysql client in new terminal
   - my2 : opens mysql client connection in dba shell
   - cu foobar : creates user foobar
   - cur foobar : creates read only user foobar
   - init foobar : WIP creates database foobar create user foobar and foobar_ro
   - databases : list database on server
   - tables : list tables for current database selected
   - use foobar : switches database context for current server

* mcli needs tmux + xpanes installed

### future

 - select prod: auto selects prod databases
 - combine cli mcli base on connection context
 - cache obj storage for server -> databases -> tables, user -> grants, refresh
 - find table
 - find column
 - mysqldump schema and/or data
 - auto fill database/tables from like mycli based on commands 
 - auto fill users like mycli based on employee data or mysql.user data
 - easy standard grant ui
 - add mongo support or mongo mode

 ## web server

 list databases

 
### future 
   
- mysqladmin like tool but for user management
- list of users and database they are in
- list databases 
- tabs tables
- find user 
- add user
- update passwords
- remove passwords