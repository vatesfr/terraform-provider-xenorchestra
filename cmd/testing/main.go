package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"os/exec"
	"strings"
	"sync"
)

type goTestOutput struct {
	Time   time.Time
	Action string
	Test   string
	Output string
}

var debug bool

func init() {

	_, exists := os.LookupEnv("TF_ACC")
	debug = !exists
}

func main() {
	fmt.Println("Starting XOA acceptance test runner")
	// TODO: Determine how to do this from within the test runner

	// goTestListArgs := []string{"test", "./...", "-list='.*'"}
	// cmd := exec.Command("go", goTestListArgs...)
	// cmd.Env = append(os.Environ(), "PWD=/home/ddelnano/go/src/github.com/ddelnano/terraform-provider-xenorchestra")

	// fmt.Println(cmd.String())

	// stdout, err := cmd.CombinedOutput()
	// fmt.Printf("stdout\n%s\n", stdout)
	// if err != nil {
	// 	log.Fatal(fmt.Sprintf("failed to list available tests with err: %v", err))
	// }

	// cmd.Wait()

	// r := bytes.NewReader(stdout)
	// scanner := bufio.NewScanner(r)

	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))

	remainingTests := []string{}
	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, "ok") || strings.HasPrefix(text, "?") {
			continue
		}
		fmt.Println(text)
		remainingTests = append(remainingTests, fmt.Sprintf("^%s$", strings.TrimSpace(text)))
	}

	fmt.Printf("Found %d tests\n", len(remainingTests))

	testRun := 1
	passedTests := []string{}
	for len(remainingTests) != 0 {
		fmt.Printf("Starting test run %d\n", testRun)

		goTestArgs := createGoTestArgs(remainingTests, debug)
		cmd := exec.Command("go", goTestArgs...)
		cmd.Env = append(os.Environ())

		fmt.Println(cmd.String())

		r, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("What is the type of StdoutPipe: %T\n", r)

		decoder := json.NewDecoder(r)
		commandFinished := make(chan bool)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			for {
				select {
				case <-commandFinished:
					fmt.Println("Stopping go routine for 'go test'")
					wg.Done()
				default:
					// fmt.Println("Starting next decoder loop")
					for decoder.More() {
						var data goTestOutput
						if err := decoder.Decode(&data); err != nil {
							if err == io.EOF {
								return
							}
							log.Printf("Failed to decode JSON: %v", err)
						}
						// fmt.Printf("Received event '%v' '%v'\n", data.Action, data.Output)
						switch data.Action {
						case "pass":
							if data.Test == "" {
								continue
							}
							passedTests = append(passedTests, data.Test)
							fmt.Printf("Found test %s passedTests: %v\n", data.Test, passedTests)
						}
						// fmt.Print(data.Output)
					}
				}
			}
		}()

		if err := cmd.Run(); err != nil {
			fmt.Errorf("command run failed wtih error: %v", err)
		}
		fmt.Println("Command Start() finished")
		// if err := cmd.Run(); err != nil {
		// 	fmt.Errorf("command run failed wtih error: %v", err)
		// }
		// fmt.Println("Command Run() finished")
		time.Sleep(35 * time.Second)
		// if err := cmd.Wait(); err != nil {
		// 	fmt.Errorf("command failed wtih error: %v", err)
		// }
		// fmt.Println("Command Wait() finished")
		commandFinished <- true
		wg.Wait()

		fmt.Printf("%d passed out of %d\n", len(passedTests), len(remainingTests))
		remainingTests = findRemainingTests(passedTests, remainingTests)
		testRun = testRun + 1
	}

}

// struct testRunReporter {
//     testsStarted []string
//     testsSkipped []string
//     testsFailed []string
//     testsSucceeded []string
// }

func createGoTestArgs(tests []string, debug bool) []string {
	goTestArgs := []string{"test", "./...", "-v", "-count=1", "-json", "-parallel", "2"}
	if !debug {
		goTestArgs = append(goTestArgs, "-sweep=true")
	}
	goTestArgs = append(goTestArgs, fmt.Sprintf(`-run='%s'`, strings.Join(tests, `|`)))
	return goTestArgs
}

func findRemainingTests(passedTests, remainingTests []string) []string {
	testsToRetry := []string{}

	for _, test := range remainingTests {
		found := false
		for _, successfulTest := range passedTests {
			if test == successfulTest {
				found = true
			}
		}
		if !found {
			testsToRetry = append(testsToRetry, test)
		}
	}
	return testsToRetry
}
