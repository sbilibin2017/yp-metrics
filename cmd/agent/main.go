package main

import (
	"context"
)

func main() {
	config, err := parseFlags()

	if err != nil {
		panic(err)
	}

	err = run(context.Background(), config)
	if err != nil {
		panic(err)
	}
}
