package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/lateralusd/bloggy/cmd"
)

//go:embed static/* templates/*
var content embed.FS

func main() {
	if err := cmd.Execute(content); err != nil {
		fmt.Fprintf(os.Stderr, "Error ocurred: %+v\n", err)
	}
}
