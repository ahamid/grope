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
const LOG_FLAGS = log.LstdFlags | log.Lshortfile

type NullWriter struct {}
func (w *NullWriter) Write(data []byte) (int, error) { return len(data), nil }

type ReadFunc func() ([]byte, error)
type Emitter interface {
  Emit(out *os.File, with_filename bool, file string, match []byte)
}

type Grope struct {
  program string
  inplace,
  inplace_many,
  with_filename,
  template,
  help,
  debug bool
  log *log.Logger
  flagSet *flag.FlagSet
  re *regexp.Regexp
  replacement *string
  files []string
  default_out *os.File
  emitter Emitter
}

func New(program string) *Grope {
  grope := Grope{}
  grope.program = path.Base(program)
  grope.log = log.New(new(NullWriter), "", LOG_FLAGS)
  grope.flagSet = flag.NewFlagSet("grope", flag.ExitOnError)
  grope.flagSet.BoolVar(&grope.help, "help", false, "print help")
  grope.flagSet.BoolVar(&grope.inplace, "i", false, "perform replacement on one file in-place")
  grope.flagSet.BoolVar(&grope.inplace_many, "I", false, "perform replacement on multiple files in-place")
  grope.flagSet.BoolVar(&grope.with_filename, "H", false, "prefix matches with file name")
  grope.flagSet.BoolVar(&grope.template, "o", false, "expand replacement string as output template")
  grope.flagSet.BoolVar(&grope.debug, "d", false, "enable debug output")
  //grope.flagSet.Usage = PrintHelpAndExit
  grope.default_out = os.Stdout
  grope.emitter = &grope
  return &grope
}

func (grope *Grope) PrintHelpAndExit() {
  fmt.Fprintf(os.Stderr, "Usage: %s [-d] [-i|-I] <regex> [replacement] [file file file]\n", grope.program)
  fmt.Fprintf(os.Stderr, "       %s --help\n", grope.program)
  grope.flagSet.PrintDefaults()
  os.Exit(BAD_ARGS)
}

func (grope *Grope) ParseArgs(args []string) {
  err := grope.flagSet.Parse(args)
  if err != nil {
    fail(err.Error(), BAD_ARGS)
  }

  args = grope.flagSet.Args()

  if len(args) <= 0 || grope.help {
    grope.PrintHelpAndExit()
  }

  grope.inplace = grope.inplace || grope.inplace_many

  if (grope.debug) {
    grope.log = log.New(os.Stderr, "", LOG_FLAGS)
  }

  grope.re = regexp.MustCompile(args[0])

  grope.replacement, grope.files = grope.parseFileArgs(args[1:])

}

func (grope *Grope) Main(args []string) {
  grope.ParseArgs(args)
  grope.Exec()
}

func (grope *Grope) Exec() {
  if grope.replacement != nil {
    grope.log.Printf("Replacement string: %s", *grope.replacement)
  }

  if len(grope.files) == 0 {
    grope.GropeFile(os.Stdin.Name(), func() ([]byte, error) {
      return ioutil.ReadAll(os.Stdin)
    })
  } else {
    if len(grope.files) > 1 && grope.inplace && !grope.inplace_many {
      fail("Operating on multiple files but -I option not specified", BAD_INPLACE_FLAG)
    }
    for _, file := range grope.files {
      grope.GropeFile(file, func() ([]byte, error) {
        return ioutil.ReadFile(file)
      })
    }
  }
}

func (grope *Grope) GropeFile(file string, read_func ReadFunc) {
  input, err := read_func()
  handle_file_error(err, "Error reading file", file, INPUT_ERR)

  out := grope.default_out

  if grope.template {
    grope.Expand(input, file, out)
  } else if grope.replacement != nil {
    out = grope.Replace(input, file, out)
  } else {
    grope.Find(input, file, out)
  }

  out.Sync()
  if (out != os.Stdout) {
    out.Close()
  }
}

func (grope *Grope) Replace(input []byte, file string, out *os.File) *os.File {
  grope.log.Printf("Processing input file: %s", file)
  if (grope.inplace) {
    if (os.Stdin.Name() == file) {
      os.Stderr.WriteString("Warning: not replacing Stdin in-place\n")
    } else {
      var err error
      out, err = os.OpenFile(file, os.O_WRONLY, 644)
      handle_file_error(err, "Error opening output file", file, OUTPUT_ERR)
    }
  }
  replaced := grope.re.ReplaceAll(input, []byte(*grope.replacement))
  grope.emitter.Emit(out, grope.with_filename, file, replaced)
  return out
}

func (grope *Grope) Find(input []byte, file string, out *os.File) {
  result := grope.re.FindAll(input, MAX_INT)
  for _, match := range result {
    grope.emitter.Emit(out, grope.with_filename, file, match)
  }
}

func (grope *Grope) Expand(input []byte, file string, out *os.File) {
  if (grope.replacement != nil) {
    result := grope.re.FindAllSubmatchIndex(input, MAX_INT)
    for _, match := range result {
      dst := []byte{}
      expanded := grope.re.Expand(dst, []byte(*grope.replacement), input, match)
      grope.emitter.Emit(out, grope.with_filename, file, expanded)
    }
  } else {
    result := grope.re.FindAllSubmatch(input, MAX_INT)
    for _, match := range result {
      for _, submatch := range match[1:] {
        grope.emitter.Emit(out, grope.with_filename, file, submatch)
      }
    }
  }
}

func (grope *Grope) Emit(out *os.File, with_filename bool, file string, match []byte) {
  if with_filename {
    out.WriteString(file + ": ")
  }
  out.Write(match)
  out.WriteString("\n")
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

func (grope *Grope) isFile(file string) bool {
  isfile := false
  info, err := os.Stat(file)
  if err != nil {
    if os.IsNotExist(err) {
      grope.log.Print("No such file: " + file)
    } else {
      grope.log.Print(err.Error())
    }
  } else if info.IsDir() {
    grope.log.Print(file + " is not a file")
  } else {
    isfile = true
  }
  return isfile
}

func (grope *Grope) parseFileArgs(args []string) (replacement *string, files []string) {
  replacement = nil
  files = args
  if len(args) > 0 {
    if !grope.isFile(args[0]) {
      replacement = &args[0]
      grope.log.Printf("Interpreting '%s' as replacement string", *replacement)
      files = args[1:]
    }
  }
  return
}


