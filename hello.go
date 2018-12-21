package main

import (
	"fmt"

	"github.com/fabiosussetto/hello/utils"
)

func main() {
	results := utils.Run()

	for _, res := range results {
		fmt.Printf("Res: %s \n", res)
	}
}
