package main

import (
	"TPark_Secure/server"
	"context"
	"fmt"
	"github.com/jackc/pgx/pgxpool"
	"log"
)

const Port = ":8080"
const dsn = `pool_max_conns=30 host=localhost port=5432 user=usr password=postgres dbname=films sslmode=disable`

func InitDatabase() (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return pgxpool.ConnectConfig(context.Background(), config)
}

func main() {
	fmt.Println("Hello")
	connection, err := InitDatabase()
	if err != nil {
		log.Fatal("Can not connect to database: ", err)
	}

	server := server.NewServer(Port, connection)
	//log.Fatal(server.NewServer(Port).ListenAndServeTLS("./server.pem", "./server.key"))
	log.Fatal(server.Serv.ListenAndServe())
}
