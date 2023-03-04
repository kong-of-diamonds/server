package main

import (
	"kod/server/internal/server"
	"log"
)

func main() {
	log.Println(server.NewServer().Run())
}
