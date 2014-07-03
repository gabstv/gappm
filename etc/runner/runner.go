package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("No programs to run...")
		os.Exit(1)
	}
	var t0, t1 float64
	flag.Float64Var(&t0, "wait0", 30, "Time in seconds to wait before executing the program.")
	flag.Float64Var(&t1, "wait1", 30, "Time in seconds to wait after executing the program.")
	flag.Parse()
	programp := flag.Arg(0)
	//
	a0 := extractArgs(programp)
	a0[0] = path.Clean(a0[0])
	cm0 := exec.Command(a0[0], a0[1:]...)
	time.Sleep(time.Duration(int64(t0 * 1000 * 1000 * 1000)))
	if er0 := cm0.Start(); er0 != nil {
		log.Println(er0)
	}
	time.Sleep(time.Duration(int64(t1 * 1000 * 1000 * 1000)))
}

func extractArgs(val string) []string {
	args := make([]string, 0)
	var buffer bytes.Buffer
	inquot := false
	var last rune
	for _, v := range val {
		if v == '"' {
			if last == '\\' {
				buffer.WriteRune(v)
			} else {
				if inquot {
					inquot = false
				} else {
					inquot = true
				}
			}
			last = '"'
		} else if v == ' ' {
			if !inquot {
				args = append(args, buffer.String())
				buffer.Truncate(0)
			} else {
				buffer.WriteRune(v)
			}
			last = ' '
		} else if v == '\\' {
			if last == '\\' {
				buffer.WriteRune(v)
			}
		} else {
			buffer.WriteRune(v)
			last = v
		}
	}
	if buffer.Len() > 0 {
		args = append(args, buffer.String())
		buffer.Truncate(0)
	}
	return args
}
