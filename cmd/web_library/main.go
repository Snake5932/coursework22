package main

import (
	server_gen "DB_coursework/internal/server"
	"DB_coursework/internal/web_library"
	"os"
)

func main() {
	server_gen.Serve(web_library.NewServer(os.Args[1]))
}
