package io

import (
	"fmt"
	"os"
	"bufio"
	"strings"
)

func checkValidPath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	return err
}

func ParseInputPath() (string, error) {
	reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter path to file: ")
    path, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	path = strings.TrimRight(path, "\n")
    err = checkValidPath(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

func PrintResult(res []string) {
	for _, r := range res {
		fmt.Println(r)
	}
}