package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	// Example 1: Basic usage with default error return
	data, _ := readFile("example.txt")//?
	fmt.Println(string(data))

	// Example 2: Custom return values
	value, err := processValue(42)//? value err
	fmt.Println(value)

	// Example 3: Multiple return values
	result, count, _ := complexOperation()//? result, count, data, err
	fmt.Printf("Result: %v, Count: %d\n", result, count)
}

func readFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func processValue(val int) (int, error) {
	if val < 0 {
		return 0, fmt.Errorf("negative value not allowed")
	}
	return val * 2, nil
}

func complexOperation() (string, int, error) {
	// Simulate a complex operation
	if _, err := os.Stat("/tmp"); err != nil {
		return "", 0, err
	}
	return "success", 42, nil
}
