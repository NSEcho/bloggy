package main

import (
	"fmt"
	"os"

	"github.com/lateralusd/bloggy/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error ocurred: %+v\n", err)
	}
}
