package lib

import "testing"
import "os"

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
  args := []string{"../../../test/test1.txt", "../../../test/test2.txt"}
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
  args := []string{"abcd", "../../../test/test2.txt"}
  replacement, files := New("test").parseFileArgs(args)
  if replacement == nil || *replacement != "abcd" {
    test.Error("first file not interpreted as replacement string")
  }
  if len(files) != 1 ||
     files[0] != args[1] {
    test.Error("parsed incorrect file args")
  }
}
