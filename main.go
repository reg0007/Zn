package main

import (
	"fmt"
	"os"

	"github.com/reg0007/Zn/cmd/zn"
)

func main() {
	if err := zn.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
