package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("Starting XOA acceptance test runner")
	goTestListArgs := []string{"test", "./...", "-list='.*'"}
	cmd := exec.Command("go", goTestListArgs...)
	cmd.Env = append(os.Environ(), "PWD=/home/ddelnano/go/src/github.com/ddelnano/terraform-provider-xenorchestra")

	fmt.Println(cmd.String())

	stdout, err := cmd.CombinedOutput()
	fmt.Printf("stdout\n%s\n", stdout)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to list available tests with err: %v", err))
	}

	r := bytes.NewReader(stdout)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
