package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	FilePath string
	Database string
	Host     string
	User     string
	Pass     string
}

func NewConfig(fileStr string) (Config, error) {
	c := Config{FilePath: fileStr}

	file, err := os.Open(fileStr)
	if err != nil {
		return c, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	var config map[string]string

	config = make(map[string]string)
	for scanner.Scan() {
		line = scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			key := strings.Trim(parts[0], " ")
			value := strings.Trim(parts[1], " ")
			config[key] = value
		}
	}

	for k, v := range config {
		switch k {
		case "database":
			c.Database = v
		case "host":
			c.Host = v
		case "user":
			c.User = v
		case "password":
			c.Pass = v
		}
	}
	if err := scanner.Err(); err != nil {
		return c, err
	}
	return c, nil
}

func (c Config) ToDSN() string {
	return fmt.Sprintf("%s:%s@(%s)/%s", c.User, c.Pass, c.Host, c.Database)
}

func FileToDSN(fileStr string) (string, error) {
	cnf, err := NewConfig(fileStr)
	if err != nil {
		return "", nil
	}

	dsn := cnf.ToDSN()

	if err != nil {
		return "", nil
	}

	return dsn, nil
}
