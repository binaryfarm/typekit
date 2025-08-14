package common

import "fmt"

type Console struct{}

func (c *Console) Log(message string) {
	fmt.Println(message)
}
