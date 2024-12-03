package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"

	_ "github.com/go-mysql-org/go-mysql/driver"
)

type server struct {
	hostname string
	port     int
	status   string
}

func (s *server) String() string {
	return fmt.Sprintf("%s:%d", s.hostname, s.port)
}

// checkHealth is checking the HTTP status API on http://127.0.0.1:10080/status
// Note that we need to calculate the HTTP API port based on the SQL port.
func (s *server) checkHealth() bool {
	offset := s.port - 4000
	httpPort := 10080 + offset
	url := fmt.Sprintf("http://%s:%d/status", s.hostname, httpPort)

	r, err := http.Head(url)
	if err != nil {
		return false
	}

	if r.StatusCode == 200 {
		return true
	}

	return false
}

func (s *server) offline(ctx context.Context, db *sql.DB) {
	fmt.Printf("offlining %s\n", s)

	// ProxySQL doesn't seem to like prepared statements for some reason
	query := fmt.Sprintf("UPDATE mysql_servers SET status='OFFLINE_SOFT' WHERE hostname='%s' AND port=%d",
		s.hostname, s.port)
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		panic(err)
	}
	db.ExecContext(ctx, "LOAD MYSQL SERVERS TO RUNTIME")
	db.ExecContext(ctx, "SAVE MYSQL SEVERSS TO DISK")
}

func main() {
	var dsn = flag.String("dsn", "admin:admin@127.0.0.1:6032", "proxysql admin dsn")
	flag.Parse()

	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "SELECT hostname,port,status FROM mysql_servers")
	if err != nil {
		panic(err)
	}

	servers := make([]server, 0)
	for rows.Next() {
		var srv server
		rows.Scan(&srv.hostname, &srv.port, &srv.status)
		servers = append(servers, srv)
	}
	rows.Close()

	for _, srv := range servers {
		fmt.Printf("Checking %s\n", srv.String())
		if !srv.checkHealth() {
			srv.offline(ctx, db)
		}
	}
}
