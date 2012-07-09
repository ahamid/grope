package lib

import "testing"
import "os"
import "regexp"

const CORPUS = "../../../test/corpus.txt"

type RecordingWriter struct {
  msgs [][]byte
}

func NewRecordingWriter() *RecordingWriter {
  r := new(RecordingWriter)
  r.msgs = make([][]byte, 0, 5)
  return r
}

func (r *RecordingWriter) Emit(out *os.File, with_filename bool, file string, match []byte) {
  r.msgs = append(r.msgs, append([]byte(nil), match...))
}

func TestSearch(test *testing.T) {
  e := NewRecordingWriter()
  grope := New("test")
  grope.emitter = e
  grope.files = []string{ CORPUS }
  grope.re = regexp.MustCompile("Never gonna.*?\n")
  grope.Exec()
  
  expected := []string {
    "Never gonna give you up,\n",
    "Never gonna let you down\n",
    "Never gonna run around and desert you\n",
    "Never gonna make you cry,\n",
    "Never gonna say goodbye\n",
    "Never gonna tell a lie and hurt you\n" }

  if len(e.msgs) != len(expected) {
    test.Error("wrong results length")
  }
  for i, match := range expected {
    if string(e.msgs[i]) != match {
      test.Error("wrong results")
    }
  }
}

func TestExpand(test *testing.T) {
  e := NewRecordingWriter()
  grope := New("test")
  grope.emitter = e
  grope.template = true
  grope.files = []string{ CORPUS }
  grope.re = regexp.MustCompile("Never gonna (.*?)[[:punct:]]*\n")
  grope.Exec()
  
  expected := []string {
    "give you up",
    "let you down",
    "run around and desert you",
    "make you cry",
    "say goodbye",
    "tell a lie and hurt you" }

  if len(e.msgs) != len(expected) {
    test.Error("wrong results length")
  }
  for i, match := range expected {
    if string(e.msgs[i]) != match {
      test.Error("wrong results")
    }
  }
}

func TestExpandWithTemplate(test *testing.T) {
  e := NewRecordingWriter()
  grope := New("test")
  grope.emitter = e
  grope.template = true
  grope.files = []string{ CORPUS }
  grope.re = regexp.MustCompile("Never gonna (.*?)[[:punct:]]*\n")
  replacement := "Always gonna $1"
  grope.replacement = &replacement
  grope.Exec()
  
  expected := []string {
    "Always gonna give you up",
    "Always gonna let you down",
    "Always gonna run around and desert you",
    "Always gonna make you cry",
    "Always gonna say goodbye",
    "Always gonna tell a lie and hurt you" }

  if len(e.msgs) != len(expected) {
    test.Error("wrong results length")
  }
  for i, match := range expected {
    if string(e.msgs[i]) != match {
      test.Error("wrong results")
    }
  }
}

func TestReplace(test *testing.T) {
  e := NewRecordingWriter()
  grope := New("test")
  grope.emitter = e
  grope.files = []string{ "../../../test/test.txt" }
  grope.re = regexp.MustCompile("(\\w+) fish")
  replacement := "pill $1"
  grope.replacement = &replacement
  grope.Exec()
  
  expected := "pill one pill two pill red pill blue\n"

  if len(e.msgs) != 1 {
    test.Error("wrong results length")
  }

  if string(e.msgs[0]) != expected {
    test.Error("wrong results")
  }
}

func TestDirIsNotFile(test *testing.T) {
  path, _ := os.Getwd()
  if (New("test").isFile(path)) {
    test.Error("Directory detected as file")
  }
}

func TestMissingIsNotFile(test *testing.T) {
  if (New("test").isFile("i don't exist")) {
    test.Error("Non-existent file detected as file")
  }
}

func TestParseValidFileArgs(test *testing.T) {
  args := []string{ CORPUS, "../../../test/test.txt"}
  replacement, files := New("test").parseFileArgs(args)
  if replacement != nil {
    test.Error("first file interpreted as replacement string")
  }
  if len(files) != len(args) ||
     files[0] != args[0] ||
     files[1] != args[1] {
    test.Error("parsed incorrect file args")
  }
}

func TestParseFileArgsWithReplacement(test *testing.T) {
  args := []string{"abcd", "../../../test/test.txt"}
  replacement, files := New("test").parseFileArgs(args)
  if replacement == nil || *replacement != "abcd" {
    test.Error("first file not interpreted as replacement string")
  }
  if len(files) != 1 ||
     files[0] != args[1] {
    test.Error("parsed incorrect file args")
  }
}
