package procedural

import (
  "os"
  "fmt"
  "bufio"
  "strings"
)

type stringMatcher func(line string) bool 

type lineProcessor func(lineno int, line string) (result string, done bool)

func Search(pattern string, flags, files []string) []string {
  var ignoreCase, exact, invert, outputFilenamesOnly, prependLineNumbers, prependFilename bool
  for _, flag := range flags {
    switch flag {
    case "-i":
      ignoreCase = true
      pattern = strings.ToLower(pattern)
    case "-v":
      invert = true
    case "-x":
      exact = true
    case "-l":
      outputFilenamesOnly = true
    case "-n":
      prependLineNumbers = true
    }
  }
  prependFilename = len(files) > 1

  output := []string{}

  for _, filename := range files {
    file, err := os.Open(filename)
    if err != nil {
      panic(err)
    }

    var lineno int
    var line string
    var match bool
    var matchLine string

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
      line = scanner.Text()
      lineno++
      if err := scanner.Err(); err != nil {
        panic(err)
      }
      if ignoreCase {
        matchLine = strings.ToLower(line)
      } else {
        matchLine = line
      }
      match = false
      if exact {
        match = matchLine == pattern
      } else {
        match = strings.Contains(matchLine, pattern)
      }
      if invert {
        match = !match
      }
      if !match {
        continue
      }
      
      if outputFilenamesOnly {
        output = append(output, filename)
        break
      }
      
      if prependFilename {
        if prependLineNumbers {
          output = append(output, fmt.Sprintf("%s:%d:%s", filename, lineno, line))
        } else {
          output = append(output, fmt.Sprintf("%s:%s", filename, line))
        }
      } else {
        if prependLineNumbers {
          output = append(output, fmt.Sprintf("%d:%s", lineno, line))
        } else {
          output = append(output, line)
        }
      }
    }
    file.Close()
  }
  return output
}
