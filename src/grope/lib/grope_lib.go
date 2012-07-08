package lib

import (
  "flag"
  "fmt"
  "regexp"
  "os"
  "path"
  "io/ioutil"
  "log"
)

const BAD_ARGS = 255
const INPUT_ERR = 1
const OUTPUT_ERR = 2
const BAD_INPLACE_FLAG = 3
const BAD_TEMPLATE_FLAG = 4

const MAX_INT = int(^uint(0) >> 1)

type NullWriter struct {}
func (w *NullWriter) Write(data []byte) (int, error) { return len(data), nil }

type read_file func() ([]byte, error)

func print_help() {
  program := path.Base(os.Args[0])
  fmt.Fprintf(os.Stderr, "Usage: %s [-d] [-i|-I] <regex> [replacement] [file file file]\n", program)
  fmt.Fprintf(os.Stderr, "       %s --help\n", program)
  flag.PrintDefaults()
  os.Exit(BAD_ARGS)
}

func fail(msg string, code int) {
  os.Stderr.WriteString(msg + "\n")
  os.Exit(code)
}

func handle_file_error(err error, file string, msg string, code int) {
  if err != nil {
    fail(msg + ": " + file, code)
  }
}

func is_file(file string) bool {
  isfile := false
  info, err := os.Stat(file)
  if err != nil {
    if os.IsNotExist(err) {
      log.Print("No such file: " + file)
    } else {
      log.Print(err.Error())
    }
  } else if info.IsDir() {
    log.Print(file + " is not a file")
  } else {
    isfile = true
  }
  return isfile
}

func parse_file_args(args []string) (replacement *string, files []string) {
  replacement = nil
  files = args
  if len(args) > 0 {
    if !is_file(args[0]) {
      replacement = &args[0]
      log.Printf("Interpreting '%s' as replacement string", *replacement)
      files = args[1:]
    }
  }
  return
}

func write_match(out *os.File, with_filename bool, file string, match []byte) {
  if with_filename {
    out.WriteString(file + ": ")
  }
  out.Write(match)
  out.WriteString("\n")
}

func replace(re *regexp.Regexp, input []byte, replacement *string, inplace bool, file string, with_filename bool, out *os.File) *os.File {
  log.Printf("Processing input file: %s", file)
  if (inplace) {
    if (os.Stdin.Name() == file) {
      os.Stderr.WriteString("Warning: not replacing Stdin in-place\n")
    } else {
      var err error
      out, err = os.OpenFile(file, os.O_WRONLY, 644)
      handle_file_error(err, "Error opening output file", file, OUTPUT_ERR)
    }
  }
  write_match(out, with_filename, file, re.ReplaceAll(input, []byte(*replacement)))
  return out
}

func find(re *regexp.Regexp, input []byte, file string, with_filename bool, out *os.File) {
  result := re.FindAll(input, MAX_INT)
  for _, match := range result {
    write_match(out, with_filename, file, match)
  }
}

func expand(re *regexp.Regexp, input []byte, template string, file string, with_filename bool, out *os.File) {
  result := re.FindAllSubmatchIndex(input, MAX_INT)
  for _, match := range result {
    dst := []byte{}
    expanded := re.Expand(dst, []byte(template), input, match)
    write_match(out, with_filename, file, expanded)
  }
}

func grope(re *regexp.Regexp, replacement *string, template bool, inplace bool, file string, with_filename bool, reader read_file) {
  input, err := reader()
  handle_file_error(err, "Error reading file", file, INPUT_ERR)

  out := os.Stdout
  if template {
    expand(re, input, *replacement, file, with_filename, out)
  } else if replacement != nil {
    out = replace(re, input, replacement, inplace, file, with_filename, out)
  } else {
    find(re, input, file, with_filename, out)
  }
  out.Sync()
  if (out != os.Stdout) {
    out.Close()
  }
}

func Main() {
  log.SetFlags(log.LstdFlags | log.Lshortfile)

  inplace := false
  inplace_many := false
  with_filename := false
  template := false
  help := false
  debug := false

  flag.BoolVar(&help, "help", false, "print help")
  flag.BoolVar(&inplace, "i", false, "perform replacement on one file in-place")
  flag.BoolVar(&inplace_many, "I", false, "perform replacement on multiple files in-place")
  flag.BoolVar(&with_filename, "H", false, "prefix matches with file name")
  flag.BoolVar(&template, "o", false, "expand replacement string as output template")
  flag.BoolVar(&debug, "d", false, "enable debug output")
  flag.Parse()

  args := flag.Args()

  if len(args) <= 0 || help {
    print_help()
  }

  inplace = inplace || inplace_many
  if (!debug) {
    log.SetOutput(new(NullWriter))
  }

  re := regexp.MustCompile(args[0])
  replacement, files := parse_file_args(args[1:])
  log.Printf("Number of files: %d", len(files))

  if replacement != nil {
    log.Printf("Replacement string: %s", *replacement)
  }

  if template && replacement == nil {
    fail("Must specify output string when using template expansion", BAD_TEMPLATE_FLAG)
  }

  if len(files) == 0 {
    grope(re, replacement, template, inplace, os.Stdin.Name(), with_filename, func() ([]byte, error) {
      return ioutil.ReadAll(os.Stdin)
    })
  } else {
    if len(files) > 1 && inplace && !inplace_many {
      fail("Operating on multiple files but -I option not specified", BAD_INPLACE_FLAG)
    }
    for _, file := range files {
      grope(re, replacement, template, inplace, file, with_filename, func() ([]byte, error) {
        return ioutil.ReadFile(file)
      })
    }
  }
}
