package main

import (
	"flag"

	"github.com/codecrafters-io/http-server-starter-go/app/server"
)

func main() {
	directory := flag.String("directory", "", "the absolute path of the directory where files are stored")
	flag.Parse()
	server := server.NewServer(*directory)
	server.Run()

}
