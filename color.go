package main

import (
	"fmt"
)

func green(str string) string {
	return fmt.Sprintf("\033[0;32m%s\033[0m",str);
}
func red(str string) string {
	return fmt.Sprintf("\033[0;31m%s\033[0m", str);
}
