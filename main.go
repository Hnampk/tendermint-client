package main

import (
	"fmt"
	"os"

	"github.com/Hnampk/tendermint-client/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
