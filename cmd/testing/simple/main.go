package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	fmt.Println("Starting test run")

	remainingTests := []string{}
	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, "ok") || strings.HasPrefix(text, "?") {
			continue
		}
		fmt.Println(text)
		remainingTests = append(remainingTests, fmt.Sprintf("^%s$", strings.TrimSpace(text)))
	}
	fmt.Printf("Found %d tests\n", len(remainingTests))
	args := createGoTestArgs(remainingTests, false, true)
	cmd := exec.Command("go", args...)

	fmt.Println(cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	if err := cmd.Start(); err != nil {
		fmt.Errorf("command run failed wtih error: %v", err)
	}
	defer cmd.Wait()
	time.Sleep(5 * time.Second)
}

func createGoTestArgs(tests []string, constructList bool, debug bool) []string {
	goTestArgs := []string{"test"}
	if constructList {
		goTestArgs = append(goTestArgs, "./...")
	} else {
		goTestArgs = append(goTestArgs, "github.com/ddelnano/terraform-provider-xenorchestra/cmd/testing/example")
	}
	goTestArgs = append(goTestArgs, "-v", "-count=1", "-json") //, "-parallel", "2")
	if !debug {
		goTestArgs = append(goTestArgs, "-sweep=true")
	}
	if constructList {
		goTestArgs = append(goTestArgs, fmt.Sprintf(`-run='%s'`, strings.Join(tests, `|`)))
	}
	return goTestArgs
}
