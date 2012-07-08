package main

import "grope/lib"
import "os"

func main() {
  lib.New(os.Args[0]).Main(os.Args[1:])
}
