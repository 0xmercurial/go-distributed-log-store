package main

import (
	"log"
	"logpack/handler"
)

func main() {
	srv := handler.NewHTTPLogServer(":8000")
	log.Fatal(srv.ListenAndServe())
}
