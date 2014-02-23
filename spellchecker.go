package spellchecker

import (
	"bufio"
	"errorcategory"
	"fmt"
	"github.com/trustmaster/go-aspell"
	"os"
	"regexp"
	"strings"
)

type SpellChecker struct {
	punctuationChars []string
	langCode         string
	scanner          *bufio.Scanner
	speller          aspell.Speller
	errCat           *errorcategory.ErrorCategory
	validExprs       map[string]string
	verbose          bool
}

func NewSpellChecker(langCode string, punctuationChars []string, errCat *errorcategory.ErrorCategory, validExprsFile string, verbose bool) *SpellChecker {
	var (
		err error
	)

	sc := new(SpellChecker)

	sc.verbose = verbose
	sc.punctuationChars = punctuationChars
	sc.langCode = langCode
	sc.errCat = errCat

	// Initialize the speller
	// TODO lang as CLI arg
	sc.speller, err = aspell.NewSpeller(map[string]string{
		"lang": langCode,
	})
	if err != nil {
		fmt.Errorf("-- aspell error: %s", err.Error())
	}
	defer sc.speller.Delete()

	sc.readValidExprs(validExprsFile)

	return sc
}

// https://github.com/trustmaster/go-aspell
func (s *SpellChecker) Check(input string) {

	var (
		err         error
		word        string
		trimmedWord string
		exprOK      bool
	)

	s.errCat.MarkError("\nSpell check\n")
	if s.verbose {
		if s.verbose && len(s.validExprs) > 0 {
			s.errCat.MarkHint("Ignoring matches for these expressions:\n")
			for key, val := range s.validExprs {
				s.errCat.MarkHint(key + "=" + val + "\n")
			}
		}
		fmt.Println()
	}

	s.scanner = bufio.NewScanner(strings.NewReader(input))
	// Set the split function for the scanning operation.
	s.scanner.Split(bufio.ScanWords)

	for s.scanner.Scan() {
		word = s.scanner.Text()
		exprOK = true

		// aspell fails to recognize valid words
		// when they have trailing punctuation
		trimmedWord = func(w string) string {
			for _, trimChar := range s.punctuationChars {
				w = strings.Trim(w, trimChar)
			}
			return w
		}(word)

		if !s.speller.Check(trimmedWord) {
			// check against any exceptions?
			if len(s.validExprs) > 0 {
				// expression OK if matching any valid expression
				for regexName, regex := range s.validExprs {
					match, _ := regexp.Match(regex, []byte(trimmedWord))
					if match {
						exprOK = true
						if s.verbose {
							s.errCat.MarkHint("{" + regexName + "} ")
						}
						break
					} else {
						exprOK = false
					}
				}
			} else {
				exprOK = false
			}
		}

		// we print the original unreduced token
		if exprOK {
			fmt.Print(word + " ")
		} else {
			s.errCat.MarkError(word + " ")
		}

	}
	if err = s.scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}
}

func (s *SpellChecker) readValidExprs(path string) {

	var (
		file    *os.File
		scanner *bufio.Scanner
		err     error
	)

	s.validExprs = make(map[string]string)

	if path == "none" {
		return
	}

	file, err = os.Open(path)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err.Error())
		return
	}

	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		// ignore if commented out
		if len(text) > 0 && !strings.HasPrefix(text, "#") {
			expr := strings.Split(scanner.Text(), "=")
			s.validExprs[expr[0]] = expr[1]
		}
	}
}
