package main

import (
	"bytes"
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/gabstv/cfg"
	"github.com/gabstv/gappm/gappm"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"time"
)

var (
	cfgm  map[string]string
	apps  []gappm.Appdef
	stop  bool
	pipef *os.File
	stdo  *os.File
)

func main() {
	loadConfig()
	// remove temporary stdout/stderr file
	defer func() {
		fn := pipef.Name()
		pipef.Close()
		os.Stdout = stdo
		os.Stderr = stdo
		pipef = nil
		os.Remove(fn)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		for k := range apps {
			apps[k].Stop = true
			if apps[k].Command != nil {
				if apps[k].Command.Process != nil {
					apps[k].Command.Process.Signal(os.Interrupt)
				}
			}
		}
		stop = true
	}()
	go gappm.StartWS()
	go startClientWS()
	execApps()
	go pipeRoutine()
	go gappm.StartHTML()
	for !stop {
		time.Sleep(time.Millisecond * 250)
	}
}

func execApps() {
	for k, _ := range apps {
		go func() {
			apps[k].Run()
		}()
	}
}

func startClientWS() {
	time.Sleep(time.Second * 1)
	gappm.ClientConnect()
	time.Sleep(time.Second * 5)
	fmt.Fprintln(pipef, time.Now().String(), "WEB CONSOLE STARTED")
}

func pipeRoutine() {
	tail, _ := tail.TailFile(pipef.Name(), tail.Config{Follow: true})
	for line := range tail.Lines {
		if !strings.HasSuffix(line.Text, "\n") {
			stdo.WriteString(line.Text + "\n")
		} else {
			stdo.WriteString(line.Text)
		}
		if !strings.HasSuffix(line.Text, "<br>") && !strings.HasSuffix(line.Text, "<br>\n") {
			gappm.Publish(line.Text + "<br>\n")
		} else {
			gappm.Publish(line.Text)
		}
	}
	stdo.WriteString("PIPE ENDED\n")
}

func loadConfig() {
	stdo = os.Stdout
	var err error
	pipef, err = ioutil.TempFile("", "gappm_")
	if err != nil {
		log.Fatal("Temp file creation error!!! (", err.Error(), ")\n")
	}
	os.Stdout = pipef
	os.Stderr = pipef
	if _, err = os.Stat(findConfig()); err != nil {
		log.Fatal("Config file not found! (", findConfig(), ")\n")
	}
	cfgm, err := cfg.ParseFile(findConfig())
	if err != nil {
		log.Fatal("Config file error!", err, "\n")
	}
	apps = make([]gappm.Appdef, 0)
	for k, v := range cfgm {
		if strings.HasPrefix(k, "exec-") {
			appd := gappm.Appdef{}
			args := extractArgs(v)
			appd.Path = args[0]
			appd.Args = args[1:]
			apps = append(apps, appd)
		}
	}
	for k := range apps {
		_, fn := path.Split(apps[k].Path)
		apps[k].LogPath = path.Join(cfgm["logpath"], fn+".log")
		if cfgm["beep_on_failure"] == "true" {
			apps[k].BeepOnFailure = true
		}
		if cfgm["mirror_app_logs"] == "true" {
			apps[k].Writer2 = os.Stdout
		}
	}
}

func findConfig() string {
	d := ""
	switch runtime.GOOS {
	case "darwin":
		d = path.Join("/Library", "Preferences", "gappm")
	case "windows":
		d = path.Join("C:\\Documents and Settings\\All Users\\Application Data", "gappm")
	case "linux":
		d = path.Join("/etc", "gappm")
	}
	return path.Join(d, "gappm.main.conf")
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
