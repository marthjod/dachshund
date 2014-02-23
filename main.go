package main

import (
	"errorcategory"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"patternchecker"
	"spellchecker"
	"sync"
)

func main() {
	var (
		err                 error
		input               []byte
		spellLang           string
		w                   sync.WaitGroup
		spellChecker        *spellchecker.SpellChecker
		patternChecker      *patternchecker.PatternChecker
		validExprsFile      string
		mistakePatternsFile string
		verbose             bool
		errorCategories     map[string]*errorcategory.ErrorCategory
	)

	// flag.StringVar(&spellLang, "spell-lang", "de_DE", "GNU aspell language code")
	// flag.Parse()
	// TODO aspell-go panics when en_US not present, f.ex.
	flag.StringVar(&validExprsFile, "valid-exprs", "none", `File containing key=value pairs of tokens not considered spelling mistakes; e.g. 'NUMBER=\d+'`)
	flag.StringVar(&mistakePatternsFile, "mistake-patterns", "none", `File containing key=value pairs for mistake patterns; e.g. 'DASS_OHNE_KOMMA=(?<!, )(?<!, ohne )(?<!, so)(?<!, so )(dass)'`)
	flag.BoolVar(&verbose, "v", false, "Verbose mode")
	flag.Parse()

	spellLang = "de_DE"

	if input, err = ioutil.ReadAll(os.Stdin); err != nil {
		fmt.Errorf("Error while reading from stdin: %s", err.Error())
		os.Exit(1)
	}

	errorCategories = make(map[string]*errorcategory.ErrorCategory)
	errorCategories["SPELLING"] = errorcategory.NewErrorCategory("SPELLING", "@{y}", "@{!y}")
	errorCategories["PATTERN"] = errorcategory.NewErrorCategory("PATTERN", "@{b}", "@{!b}")

	go func(in, validExprsFile string, verbose bool) {
		spellChecker = spellchecker.NewSpellChecker(spellLang, []string{",", "."}, errorCategories["SPELLING"], validExprsFile, verbose)
		spellChecker.Check(in)
		// goroutine finished
		w.Done()
	}(string(input), validExprsFile, verbose)
	w.Add(1)

	go func(in, patternsFile string, verbose bool) {
		patternChecker = patternchecker.NewPatternChecker(errorCategories["PATTERN"], mistakePatternsFile, verbose)
		patternChecker.Check(in)
		// goroutine finished
		w.Done()
	}(string(input), mistakePatternsFile, verbose)
	w.Add(1)

	// wait for all goroutines to finish
	w.Wait()
}
