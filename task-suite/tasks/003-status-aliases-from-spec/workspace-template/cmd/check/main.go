package main

import (
	"fmt"

	"task/statusaliases/internal/status"
)

func main() {
	fmt.Println(status.NormalizeStatus("enabled"))
	fmt.Println(status.NormalizeStatus("disabled"))
	fmt.Println(status.NormalizeStatus(" ACTIVE "))
	fmt.Println(status.NormalizeStatus("mystery"))
}
