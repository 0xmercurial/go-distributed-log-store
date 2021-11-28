package main

import (
	"log"
	"logstore/handler"
)

func main() {
	srv := handler.NewHTTPLogServer(":8000")
	log.Fatal(srv.ListenAndServe())
}
