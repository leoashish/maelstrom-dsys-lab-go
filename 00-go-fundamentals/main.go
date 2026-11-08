// Chapter 0: Go Fundamentals Refresher
//
// One consolidated chapter, read top to bottom or run section by
// section, covering: basics, structs & interfaces, error handling,
// collections & generics, JSON & app-development conventions, and
// concurrency (both the primitives and the named patterns built from
// them). See README.md in this folder for how the sections map to
// what's used later in the challenges.
//
// Run everything in order:
//
//	go run .
//
// Run just one section (useful once you've refreshed the earlier ones
// and want to jump straight to, say, concurrency):
//
//	go run . concurrency
//	go run . patterns
//
// Valid section names are the keys of the `sections` map below.
package main

import (
	"fmt"
	"os"
	"strings"
)

var sections = map[string]func(){
	"basics":      RunBasics,
	"structs":     RunStructsAndInterfaces,
	"errors":      RunErrorHandling,
	"collections": RunCollectionsAndGenerics,
	"json":        RunJSONAndAppDev,
	"concurrency": RunConcurrencyBasics,
	"patterns":    RunConcurrencyPatterns,
}

// orderedSectionNames fixes the run order -- map iteration in Go is
// randomized (see section 4), which would make "run everything" print
// in a different order every time if we ranged over `sections` directly.
var orderedSectionNames = []string{
	"basics", "structs", "errors", "collections", "json", "concurrency", "patterns",
}

func main() {
	if len(os.Args) > 1 {
		name := strings.ToLower(os.Args[1])
		fn, ok := sections[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown section %q. valid names: %s\n", os.Args[1], strings.Join(orderedSectionNames, ", "))
			os.Exit(1)
		}
		fn()
		return
	}

	for _, name := range orderedSectionNames {
		sections[name]()
		fmt.Println()
	}

	fmt.Println("That's the refresher. Next: run `go test ./...` in this folder for a")
	fmt.Println("few small test-first warm-up exercises (see exercises.go), then head")
	fmt.Println("to ../challenges/01-echo/ to start the lab itself.")
}

func section(title string) {
	fmt.Println(strings.Repeat("=", len(title)))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", len(title)))
}
