package main

import (
  "regexp"
  "os"
  "io/ioutil"
)

func main() {
  regexp := regexp.MustCompile(os.Args[1])
  var b []byte
  var out *os.File
  var replacement []byte
  var err error
  var infile *string
  inplace := false

  args := os.Args[0:]

  out = os.Stdout

  if len(args) > 1 && args[1] == "-i" {
    inplace = true
    args = args[1:]
  }

  if len(args) > 3 {
    replacement = []byte(args[2])
    infile = &args[3]
    b, _ = ioutil.ReadFile(*infile)
  } else if len(args) > 2 {
    infile = &args[2]
    b, err = ioutil.ReadFile(*infile)
    if os.IsNotExist(err) {
      infile = nil
      // file does not exist, treat arg #2 as a replacement string
      // use Stdin as input
      replacement = []byte(args[2])
      b, _ = ioutil.ReadAll(os.Stdin)
    }
  } else {
    b, _ = ioutil.ReadAll(os.Stdin)
  }

  if infile != nil && inplace {
    out, _ = os.OpenFile(*infile, os.O_WRONLY, 644)
  }

  if replacement != nil {
    out.Write(regexp.ReplaceAll(b, replacement))
  } else {
    result := regexp.FindAll(b, 100)
    for _, match := range result {
      out.Write(match)
      out.Write([]byte("\n"))
    }
  }
}
