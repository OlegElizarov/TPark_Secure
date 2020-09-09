package main

import (
	"TPark_Secure/server"
	"fmt"
	"log"
)

const Port = ":8080"

func main() {
	fmt.Println("Hello")
	//log.Fatal(server.NewServer(Port).ListenAndServeTLS("./server.pem", "./server.key"))
	log.Fatal(server.NewServer(Port).ListenAndServe())
}
