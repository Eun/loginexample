package main

import (
	"log"

	gogenapi "github.com/Eun/go-gen-api"
)

func main() {
	type User struct {
		ID       int64
		Name     string
		Password string
	}

	type Token struct {
		ID     int64
		UserID int64
	}

	err := gogenapi.Generate(&gogenapi.Config{
		Structs:    []interface{}{&User{}, &Token{}},
		OutputPath: "gogenapi",
	})
	if err != nil {
		log.Fatal(err)
	}
}
