package main

import (
	"context"
)

func main() {
	config := parseFlags()
	err := run(context.Background(), config)
	if err != nil {
		panic(err)
	}
}
