package sql

import (
	fmt "fmt"
	"strings"
)

type DSN struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
	SSLMode  string
	Charset  string
}

func (that *DSN) String() string {
	var args []string
	if that.Host != "" {
		args = append(args, fmt.Sprintf("host=%s", that.Host))
	}

	if that.Port != 0 {
		args = append(args, fmt.Sprintf("port=%d", that.Port))
	}

	if that.Database != "" {
		args = append(args, fmt.Sprintf("dbname=%s", that.Database))
	}

	if that.User != "" {
		args = append(args, fmt.Sprintf("user=%s", that.User))
	}

	if that.Password != "" {
		args = append(args, fmt.Sprintf("password=%s", that.Password))
	}

	if that.SSLMode != "" {
		args = append(args, fmt.Sprintf("sslmode=%s", that.SSLMode))
	}

	if that.Charset != "" {
		args = append(args, fmt.Sprintf("charset=%s", that.Charset))
	}

	return strings.Join(args, " ")
}

func DefaultDSN() DSN {
	return DSN{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
	}
}
