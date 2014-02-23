# dachshund

**Parses textual content from stdin and hints at potential {typo,ortho}graphical or stylistic errors contained in it by appropriate highlighting of the output.**
(_dachshund_ is a phonetic pun on "_docshint_" which would be a concise way to describe the tool's function.)

### Checks

- Rules for checks come from external sources, _dachshund_ merely acts as a glorified wrapper.
- Optionally add `-v` for verbose mode.
- Checks run in separate goroutines, so each checker has its own output.
- External pattern files allow blank and comment lines.

#### Spell checker

- Uses the _go-aspell_ library to invoke the local system's _aspell_ spell checker while scanning the input. Spelling mistakes are highlighted in the output.
- Tokens which should not be considered spelling mistakes but instead ignored (proper names, product names, e.g.) can be defined as valid expressions in an external file.
- Example: If the following entry were defined in _ignore.patterns_:

    ```
# Don't consider numbers spelling mistakes.
NUMBER=\d+
    ```

    ```bash
 echo "1234" | ./dachshund -valid-exprs=ignore.patterns
    ```

    would not highlight "1234" in the error type's color, whereas without this rule, the spell checker would highlight it as a possible mistake.

- In verbose mode, each ignored pattern is marked with a hint to its appropriate ignore rule (primarily useful for debugging patterns), thus:


```
    Spell check
    Ignoring matches for these expressions:
    NUMBER=\d+

    {NUMBER} 1234
```

#### Pattern checker

- Scans input (sentence-wise, i.e. by splitting at `.`), trying to match mistake patterns (described in the form of regular expressions). Patterns have a descriptive short name used to categorize the respective mistake type and mark it in the output.
- Uses an external file describing mistake patterns.
- Example: If the following entry were defined in _mistake.patterns_:

    ```ini
# !( ", dass" | ", so dass" | ", sodass" | ", ohne dass" )
DASS_OHNE_KOMMA=(?<!, )(?<!, ohne )(?<!, so)(?<!, so )(dass)
    ```

    ```bash
 echo "so sehr dass" | ./dachshund -mistake-patterns=mistake.patterns
    ```

 would produce the following output (color-highlighted if run in a shell):

    ```
 so sehr {DASS_OHNE_KOMMA} dass
    ```

- Patterns may have flags


| Flag | Function | Example |
|:--------|:------------|:-----------|
| `multi-word:` | Tells the highlighter that the expression stretches more than one token. | `multi-word:KOMPOSITUM_LEERZEICHEN=([A-ZÄÖÜ][A-ZÄÖÜa-zäöüß]+ ?){2,}` |
| `mid-sentence:`| Tells the highlighter to ignore a pattern match if it occurs at the beginning of a sentence. | `mid-sentence:KOMPOSITUM_LEERZEICHEN=([A-ZÄÖÜ][A-ZÄÖÜa-zäöüß]+ ?){2,}` |

- Flags may be combined: `multi-word:mid-sentence:KOMPOSITUM_LEERZEICHEN=([A-ZÄÖÜ][A-ZÄÖÜa-zäöüß]+ ?){2,}` finds all two or more successive words starting with a capital letter, except after `.`.

## Dependencies

- [glenn-brown/golang-pkg-pcre/src/pkg/pcre](https://github.com/glenn-brown/golang-pkg-pcre/) because the standard [_regexp_](http://code.google.com/p/re2/wiki/Syntax) package does not allow negative lookbehinds in regexes
- [trustmaster/go-aspell](https://github.com/trustmaster/go-aspell) for bindings to the system's _GNU aspell_ libraries
- [wsxiaoys/terminal/color](https://github.com/wsxiaoys/terminal) for colored terminal output


## Building _dachshund_

- `go get` the dependencies.
- Build, install and run _dachshund_:

```bash
for p in errorcategory patternchecker spellchecker
do
    mkdir -p $GOPATH/src/$p && \
    go build $p.go && \
    cp $p.go $GOPATH/src/$p && \
    go install $p
done
go build -o dachshund main.go

./dachshund -h                                                                                                                                                                                          
Usage of ./dachshund:
  -mistake-patterns="none": File containing key=value pairs for mistake patterns; e.g. 'DASS_OHNE_KOMMA=(?<!, )(?<!, ohne )(?<!, so)(?<!, so )(dass)'
  -v=false: Verbose mode
  -valid-exprs="none": File containing key=value pairs of tokens not considered spelling mistakes; e.g. 'NUMBER=\d+'
```

## Limitations

- Language: `de_DE`; providing other language codes as command-line args is not implemented yet because the _go-aspell_ library panics if a respective language package is missing on the local system.

## TODOs

- godoc
- struct methods' visibility
- Check for problems with umlauts etc.
- Rename packages to play nicely with $GOPATH.
- Find more and tweak patterns.
