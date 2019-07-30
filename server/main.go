package main

import (
	"github.com/sergivb01/forums/service"
)

func main() {
	s, err := service.NewServer("./config.yml")
	if err != nil {
		panic(err)
	}

	s.Listen(":8080")
}
