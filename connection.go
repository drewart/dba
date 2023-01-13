package dba

// interface to the db
// also a way to parse .cnf file
// Connection

type Connection struct {
	Name string
}

type ConnectionConfig struct {
	Name string
	File string
}

func (c ConnectionConfig) Connect() Connection {
	// open config file

	return Connection{}
}
