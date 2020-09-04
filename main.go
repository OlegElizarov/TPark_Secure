package main

import (
	"TPark_Secure/server"
	"fmt"
	"github.com/labstack/echo"
	"log"
)

const Port = ":8080"

func main() {
	fmt.Println("Hello")
	e := echo.New()
	srv0 := server.NewServer(Port, e)
	log.Fatal(srv0.ListenAndServe())
}
