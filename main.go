package main

import (
	"net/http"

	"github.com/coduno/compute/runner"
)

func main() {
	http.Handle("/simple", http.HandlerFunc(runner.Simple))
	http.Handle("/javaut", http.HandlerFunc(runner.JavaUnitTest))
	http.Handle("/outputtest", http.HandlerFunc(runner.OutputTest))
	http.ListenAndServe(":8081", nil)
}
