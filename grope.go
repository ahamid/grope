package main

import (
  "regexp"
  "os"
  "io/ioutil"
)

func main() {
  var re *regexp.Regexp
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

  re= regexp.MustCompile(args[1])
 
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
    os.Stderr.Write([]byte("Setting out to " + *infile + "\n"))
    out, err = os.OpenFile(*infile, os.O_WRONLY, 644)
    if (err != nil) {
      os.Stderr.Write([]byte("Error setting outputfile"))
    }
  }

  os.Stderr.Write([]byte(args[1]))
  if replacement != nil {
    os.Stderr.Write([]byte("WRITING REPLACEMENT " + string(replacement) + "\n"))
    os.Stderr.Write(b)
    os.Stderr.Write(re.ReplaceAll(b, replacement))
    out.Write(re.ReplaceAll(b, replacement))
  } else {
    result := re.FindAll(b, 100)
    for _, match := range result {
      os.Stderr.Write([]byte("Writing match " + string(match)))
      out.Write(match)
      out.Write([]byte("\n"))
    }
  }
  out.Sync()
}
