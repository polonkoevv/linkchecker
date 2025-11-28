package main

import (
	"fmt"

	"github.com/polonkoevv/linkchecker/internal/config"
)

func main() {

	cfg := config.MustLoad()
	fmt.Println(cfg)
	fmt.Println("Hello World!")
}
