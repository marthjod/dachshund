package patternchecker

import (
	"bufio"
	"errorcategory"
	"fmt"
	// Standard package "regexp" does not support negative lookbehinds,
	// cf. godoc regexp/syntax
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"os"
	"strings"
)

const (
	// Highlighter must know if pattern may be stretched across multiple tokens.
	multiWordMarker = "multi-word:"
	// Should we mark pattern only if not at beginning of sentence?
	midSentenceMarker = "mid-sentence:"
)

type MistakePattern struct {
	Name    string
	Pattern string
	Attrs   map[string]bool
}

func NewMistakePattern(name, pattern string, attrs map[string]bool) *MistakePattern {
	m := new(MistakePattern)
	m.Name = name
	m.Pattern = pattern
	m.Attrs = attrs
	return m
}

type PatternChecker struct {
	scanner         *bufio.Scanner
	errCat          *errorcategory.ErrorCategory
	verbose         bool
	mistakePatterns []MistakePattern
	Matches         int
}

func NewPatternChecker(errCat *errorcategory.ErrorCategory, mistakePatternsFile string, verbose bool) *PatternChecker {

	c := new(PatternChecker)
	c.errCat = errCat
	c.verbose = verbose
	c.readPatterns(mistakePatternsFile)
	c.Matches = 0
	return c
}

func (c *PatternChecker) split(input string) []string {
	var (
		lines      []string
		splitLines []string
		i          int
	)

	splitLines = make([]string, 0)

	lines = strings.Split(input, ".")
	for i = range lines {
		lines[i] = strings.Trim(lines[i], " \n")
		if lines[i] != "" {
			splitLines = append(splitLines, lines[i])
		}
	}

	return splitLines
}

func (c *PatternChecker) Check(input string) {

	var (
		splitLines       []string
		re               pcre.Regexp
		reErr            *pcre.CompileError
		matcher          *pcre.Matcher
		line             string
		pattern          MistakePattern
		patterns         []MistakePattern
		patternMatched   bool
		patternNames     []string
		currPatternName  string
		currPatternAttrs []string
	)

	patterns = c.mistakePatterns

	c.errCat.MarkError("\n# Pattern check\n")
	if c.verbose && len(patterns) > 0 {

		fmt.Printf("# Looking for %d mistake patterns:\n", len(patterns))

		for p := range patterns {

			currPatternName = "# " + patterns[p].Name

			if len(patterns[p].Attrs) > 0 {
				currPatternName += " ("
				for attrName := range patterns[p].Attrs {
					currPatternAttrs = append(currPatternAttrs, strings.Replace(attrName, ":", "", -1))
				}
				currPatternName += strings.Join(currPatternAttrs, ", ") + ")"
			}

			patternNames = append(patternNames, currPatternName)
		}

		fmt.Print(strings.Join(patternNames, "\n"))
		fmt.Println()
	}

	splitLines = c.split(input)
	// look for any mistake pattern in each line
	patternMatched = false

	for i := range splitLines {
		line = splitLines[i]

		for p := range patterns {
			pattern = patterns[p]

			if re, reErr = pcre.Compile(pattern.Pattern, 0); reErr == nil {
				if matcher = re.MatcherString(line, 0); matcher.Matches() {

					patternMatched = true
					c.Matches++

					// pattern found in line, scan line again
					// and highlight position
					if pattern.Attrs[multiWordMarker] {
						go c.markMultiTokens(line, pattern, matcher)
					} else {
						go c.markSingleTokens(line, pattern)
					}

				} else {
					patternMatched = false
				}
			} else {
				fmt.Errorf("Error compiling regex: %v", reErr.Message)
			}
		}

		// no patterns matches in current line
		if !patternMatched {
			// lines have originally been split by "."
			fmt.Println(line + ".")
		}

	}
}

func (c *PatternChecker) markMultiTokens(sentence string, pattern MistakePattern, matcher *pcre.Matcher) {
	var (
		scanner             *bufio.Scanner
		word                string
		highlightGroup      string
		atSentenceBeginning bool
	)

	hint := c.errCat.MarkHint
	mark := c.errCat.MarkError

	scanner = bufio.NewScanner(strings.NewReader(sentence))
	scanner.Split(bufio.ScanWords)
	for group := 0; group < matcher.Groups(); group++ {
		highlightGroup = ""
		atSentenceBeginning = true

		// while sentence.hasNext()
		for scanner.Scan() {
			word = scanner.Text()

			if strings.HasPrefix(matcher.GroupString(group), highlightGroup+word) {
				highlightGroup += word + " "
			} else {
				if highlightGroup != "" {
					if !atSentenceBeginning || !pattern.Attrs[midSentenceMarker] {
						hint("{" + pattern.Name + "} ")
						mark(highlightGroup)
					} else {
						fmt.Print(highlightGroup)
					}
					// reset
					highlightGroup = ""

				} else {
					fmt.Print(word + " ")
				}
				atSentenceBeginning = false
			}
		}

		fmt.Print("\b. ")
	}
}

func (c *PatternChecker) markSingleTokens(sentence string, pattern MistakePattern) {
	var (
		matcher *pcre.Matcher
		re      pcre.Regexp
		reErr   *pcre.CompileError
		scanner *bufio.Scanner
		word    string
	)

	hint := c.errCat.MarkHint
	mark := c.errCat.MarkError

	if re, reErr = pcre.Compile(pattern.Pattern, 0); reErr == nil {
		scanner = bufio.NewScanner(strings.NewReader(sentence))
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			word = scanner.Text()
			if matcher = re.Matcher([]byte(word), 0); matcher.Matches() {
				hint("{" + pattern.Name + "} ")
				mark(word + " ")
			} else {
				fmt.Print(word + " ")
			}
		}
		// kill last superfluous space and
		// restore sentence ending
		fmt.Print("\b. ")
	}
}

func (c *PatternChecker) readPatterns(path string) {
	var (
		file         *os.File
		scanner      *bufio.Scanner
		err          error
		name         string
		pattern      string
		patternAttrs map[string]bool
	)

	c.mistakePatterns = make([]MistakePattern, 0)

	if path == "none" {
		return
	}

	if file, err = os.Open(path); err != nil {
		fmt.Printf("ERROR: %v\n", err.Error())
		return
	}

	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		// ignore if commented out or empty line
		if len(text) > 0 && !strings.HasPrefix(text, "#") {
			// pattern
			patternAttrs = make(map[string]bool)

			expr := strings.Split(scanner.Text(), "=")
			name = expr[0]
			if strings.Contains(name, multiWordMarker) {
				patternAttrs[multiWordMarker] = true
				name = strings.Replace(name, multiWordMarker, "", 1)
			}
			if strings.Contains(name, midSentenceMarker) {
				patternAttrs[midSentenceMarker] = true
				name = strings.Replace(name, midSentenceMarker, "", 1)
			}
			pattern = expr[1]
			c.mistakePatterns = append(c.mistakePatterns,
				*NewMistakePattern(name,
					pattern, patternAttrs))
		}
	}

	//if c.verbose {
	//	fmt.Printf("Patterns read in: %v\n", c.mistakePatterns)
	//}
}
