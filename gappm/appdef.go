package gappm

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

type Appdef struct {
	Path          string
	UpdatePath    string
	Args          []string
	Command       *exec.Cmd
	LogPath       string
	BeepOnFailure bool
	Stop          bool
	Writer2       io.Writer
}

func (a *Appdef) Run() {
	if a.Writer2 == nil {
		// /dev/null
		a.Writer2 = ioutil.Discard
	}
	a.Command = exec.Command(a.Path, a.Args...)
	// check log file
	var f *os.File
	if _, err := os.Stat(a.LogPath); err != nil {
		f, err = os.OpenFile(a.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(a.Path, a.LogPath, err)
		}
	} else {
		f, _ = os.OpenFile(a.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	w := io.MultiWriter(a.Writer2, f)
	a.Command.Stdout = w
	a.Command.Stderr = w

	err := a.Command.Start()
	if err != nil {
		if a.BeepOnFailure {
			fmt.Print("\x07")
		}
		fmt.Fprintln(w, a.Path, err.Error(), "\n")
		time.Sleep(time.Second * 30)
		go func() {
			if !a.Stop {
				a.Run()
			}
		}()
		return
	}
	fmt.Fprintln(w, time.Now().String(), a.Path, "[[[started]]].")
	time.Sleep(time.Second * 3)
	if a.Command.Process != nil {
		stat, err := a.Command.Process.Wait()
		if err != nil {
			fmt.Fprintln(w, a.Path, err.Error())
		} else {
			if a.BeepOnFailure {
				fmt.Print("\x07")
			}
			fmt.Fprintln(w, "[[[Program]]]", a.Path, "[[[exited]]].", time.Now().String())
			fmt.Fprintln(w, stat.SystemTime(), stat.UserTime())
		}
	} else {
		time.Sleep(time.Second * 25)
	}
	go func() {
		if !a.Stop {
			a.Run()
		}
	}()
}
