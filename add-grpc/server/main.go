package main

import (
	"add-grpc/handler"
	"log"
)

func main() {
	srv := handler.NewHTTPLogServer(":8000")
	log.Fatal(srv.ListenAndServe())
}
