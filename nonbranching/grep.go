package nonbranching

import (
  "slices"
  "os"
  "fmt"
  "bufio"
  "strings"
)

type stringMatcher func(line string) bool 

type lineProcessor func(lineno int, line string) (result string, done bool)

func Search(pattern string, flags, files []string) []string {
  matchBuilder := newMatchBuilder(pattern)
  if slices.Contains(flags, "-i") {
    matchBuilder = matchBuilder.caseInsensitive()
  }
  if slices.Contains(flags, "-v") {
    matchBuilder = matchBuilder.invert()
  }
  if slices.Contains(flags, "-x") {
    matchBuilder = matchBuilder.exact()
  }
  var matcher = matchBuilder.build();
  
  outputFilenameOnly := slices.Contains(flags, "-l")
  prependLineNumber := slices.Contains(flags, "-n")
  prependFilename := len(files) > 1

  var output []string

  lineProcessor := newLineProcessor(outputFilenameOnly, prependLineNumber, prependFilename, matcher)
  
  for _, filename := range files {
    processor := lineProcessor(filename)  

    if result, err := scanFile(filename, processor); err != nil {
      panic(err)
    } else {
      output = append(output, result...)
    }
  } 
	return output
}

func newLineProcessor(outputFilenameOnly, prependLineNumber, prependFilename bool, matched stringMatcher) func(filename string) lineProcessor {

  if outputFilenameOnly {
    return func(filename string) lineProcessor {
      return func(lineno int, line string) (result string, done bool) {
        if matched(line) {
          result = filename
          done = true
        }
        return
      }
    }
  }

  if prependFilename {
    if prependLineNumber {
      return func(filename string) lineProcessor {
        return func(lineno int, line string) (result string, done bool) {
          if matched(line) {
            result = fmt.Sprintf("%s:%d:%s", filename, lineno, line)
          }
          return
        }
      }
    } else {
      return func(filename string) lineProcessor {
        return func(lineno int, line string) (result string, done bool) {
          if matched(line) {
            result = fmt.Sprintf("%s:%s", filename, line)
          }
          return
        }
      }
    }
  }

  if prependLineNumber {
    return func(filename string) lineProcessor {
      return func(lineno int, line string) (result string, done bool) {
        if matched(line) {
          result = fmt.Sprintf("%d:%s", lineno, line)
        }
        return
      }
    }
  }

  return func(filename string) lineProcessor {
    return func(lineno int, line string) (result string, done bool) {
      if matched(line) {
        result = line
      }
      return
    }
  }
}

func scanFile(filename string, process lineProcessor) ([]string, error) {
  file, err := os.Open(filename)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var output []string

  var lineno int
  var line string

  scanner := bufio.NewScanner(file)

  var result string
  var done bool

  for scanner.Scan() && !done {
    lineno++
    line = scanner.Text()
    if err := scanner.Err(); err != nil {
      panic(err)
    }
    result, done = process(lineno, line)
    if len(result) > 0 {
      output = append(output, result) 
    }
  } 

  return output, nil
}

type matchBuilder struct {
  ignoreCase bool
  negate bool
  exactMatch bool
  pattern string
}
func newMatchBuilder(pattern string) *matchBuilder {
  return &matchBuilder{ignoreCase: false, negate: false, exactMatch: false, pattern: pattern}
}
func (b *matchBuilder) caseInsensitive() *matchBuilder {
   b.ignoreCase = true
   return b
}
func (b *matchBuilder) invert() *matchBuilder {
   b.negate = true
   return b
}
func (b *matchBuilder) exact() *matchBuilder {
  b.exactMatch = true
  return b
}
func (b *matchBuilder) build() (match stringMatcher) {
  if b.ignoreCase && b.exactMatch {
    lowerPattern := strings.ToLower(b.pattern)
    if b.negate {
      match = func(line string) bool {
        return lowerPattern != strings.ToLower(line)
      } 
    } else {
      match = func(line string) bool {
        return lowerPattern == strings.ToLower(line)
      } 
    }
  } else if b.ignoreCase {
    lowerPattern := strings.ToLower(b.pattern)
    if b.negate {
      match = func(line string) bool {
        return !strings.Contains(strings.ToLower(line), lowerPattern)
      } 
    } else {
      match = func(line string) bool {
        return strings.Contains(strings.ToLower(line), lowerPattern)
      } 
    }
  } else if b.exactMatch {
    if b.negate {
      match = func(line string) bool {
        return line != b.pattern
      } 
    } else {
      match = func(line string) bool {
        return line == b.pattern
      } 
    }
  } else {
    if b.negate {
      match = func(line string) bool {
        return !strings.Contains(line, b.pattern)
      }
    } else {
      match = func(line string) bool {
        return strings.Contains(line, b.pattern)
      }
    }
  }
  return
}
