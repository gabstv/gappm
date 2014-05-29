package main

import (
	"bytes"
	"fmt"
	"github.com/ActiveState/tail"
	"github.com/ActiveState/tail/watch"
	"github.com/gabstv/cfg"
	"github.com/gabstv/gappm/gappm"
	"io"
	"io/ioutil"
	"launchpad.net/tomb"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	cfgm    map[string]string
	apps    []gappm.Appdef
	stop    bool
	restart bool
	pipef   *os.File
	stdo    *os.File
	sigk    os.Signal
	SIGTERM os.Signal = syscall.SIGTERM
	SIGQUIT os.Signal = syscall.SIGQUIT
	//
	lastDay int
	c       chan os.Signal
	tomb0   tomb.Tomb
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

	c = make(chan os.Signal, 1)
	if runtime.GOOS == "windows" {
		signal.Notify(c, os.Interrupt, os.Kill, SIGTERM, SIGQUIT)
	} else {
		signal.Notify(c, os.Interrupt, os.Kill)
	}
	go func() {
		<-c
		for k := range apps {
			apps[k].Stop = true
			if apps[k].Command != nil {
				if apps[k].Command.Process != nil {
					apps[k].Command.Process.Signal(sigk)
				}
			}
		}
		stop = true
	}()
	if cfgm["webconsole"] != "false" {
		go gappm.StartWS()
		go startClientWS()
	}
	execApps()
	go pipeRoutine()
	go watchApps()

	if cfgm["webconsole"] != "false" {
		go gappm.StartHTML()
	}
	//if cfgm["watch_config"] == "true" {
	//go watchConfig()
	//}
	for !stop {
		time.Sleep(time.Millisecond * 250)
		ld := time.Now().Day()
		if !stop && ld != lastDay {
			lastDay = ld
			// RELOG
			ldays, _ := strconv.Atoi(cfgm["delete_logs_older_than_x_days"])
			fmt.Println("RE LOGGING", ldays)
			for k := range apps {
				apps[k].ReLog(ldays)
			}
		}
	}
	log.Println("[[[ [[[ [[[ [[[ [[[ [[[ [[[[ EXIT ]]]] ]]] ]]] ]]] ]]] ]]] ]]]")
}

func watchApps() {
	for !stop {
		for k, _ := range apps {
			time.Sleep(time.Millisecond * 100)
			if len(apps[k].UpdatePath) > 0 && !apps[k].Stop {
				if _, err := os.Stat(apps[k].UpdatePath); err == nil {
					apps[k].Stop = true
					fmt.Println("Updating", apps[k].Path, ":", "COPYING OLD VERSION")
					f, err := os.OpenFile(apps[k].Path+".old", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0744)
					if err == nil {
						forig, err := os.Open(apps[k].Path)
						if err == nil {
							io.Copy(f, forig)
							forig.Close()
						}
						f.Close()
					}
					if apps[k].Command != nil {
						if apps[k].Command.Process != nil {
							fmt.Println("Updating", apps[k].Path, ":", "KILLING PROCESS")
							apps[k].Command.Process.Signal(sigk)
							time.Sleep(time.Millisecond * 1500)
						}
					}
					f, err = os.OpenFile(apps[k].UpdatePath, os.O_RDONLY, 0744)
					if err == nil {
						fmt.Println("Updating", apps[k].Path, ":", "UPDATING FILE")
						fupd, err := os.OpenFile(apps[k].Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0744)
						if err == nil {
							io.Copy(fupd, f)
							f.Close()
							f = nil
							fupd.Close()
							fupd = nil
							time.Sleep(time.Second * 1)
							err = os.Remove(apps[k].UpdatePath)
							for err != nil {
								fmt.Println("[FATAL] [FATAL] [FATAL] ERROR ON DELETING UPDATE FILE:", err.Error())
								time.Sleep(time.Second * 2)
								err = os.Remove(apps[k].UpdatePath)
							}
						} else {
							f.Close()
							f = nil
						}
					}
					fmt.Println("Updating", apps[k].Path, ":", "RESTARTING PROCESS")
					apps[k].Stop = false
					time.Sleep(time.Millisecond * 100)
					go apps[k].Run()
					fmt.Println("Updating", apps[k].Path, ":", "DONE")
				}
			}
			if apps[k].Cron.UseTime {
				if apps[k].CronTest() {
					if !apps[k].IsRunning() {
						fmt.Println("[CRON] Starting", apps[k].Path, "(cron job)")
						apps[k].Stop = false
						go apps[k].Run()
						time.Sleep(time.Microsecond * 10)
					}
				} else {
					if apps[k].IsRunning() {
						er90 := apps[k].Command.Process.Signal(sigk)
						if er90 == nil {
							fmt.Println("[CRON] Killing", apps[k].Path, "(cron job)")
							apps[k].Stop = true
							time.Sleep(time.Millisecond * 500)
							apps[k].Stop = false
						}
					}
				}
			}
		}
		time.Sleep(time.Second * 30)
	}
}

func getExtension() string {
	if runtime.GOOS == "windows" {
		return ".bat"
	}
	return ".sh"
}

func watchConfig() {
	fmt.Print("\x07")
	time.Sleep(time.Second * 15)
	fmt.Fprintln(pipef, time.Now().String(), "~~~~~ WATCHING CONFIG FILE ~~~~~")
	w := watch.NewPollingFileWatcher(findConfig())
	w.BlockUntilExists(&tomb0)
	fmt.Fprintln(pipef, time.Now().String(), "~~~~~ CONFIG FILE CHANGED ~~~~~")
	if len(cfgm["restart_command"]) > 1 {
		tempdir := os.TempDir()
		tf := path.Join(tempdir, "gappm_restart"+getExtension())
		os.Remove(tf)
		fupd, _ := os.OpenFile(tf, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0744)
		fupd.Write([]byte(cfgm["restart_command"]))
		fupd.Close()
		cmd0 := exec.Command(tf, os.Args[0])
		cmd0.Start()
	}
}

func execApps() {
	for k, _ := range apps {
		go func(n int) {
			apps[n].Run()
		}(k)
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
		if !strings.Contains(line.Text, "<br>") && len(line.Text) > 10 {
			if cfgm["webconsole"] != "false" {
				gappm.Publish(line.Text + "<br>\n")
			}
		} else {
			if cfgm["webconsole"] != "false" {
				gappm.Publish(line.Text)
			}
		}
	}
	stdo.WriteString("PIPE ENDED\n")
}

func loadConfig() {
	lastDay = time.Now().Day()
	sigk = os.Interrupt
	if runtime.GOOS == "windows" {
		sigk = os.Kill
	}
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
			log.Println("APP: ", k, v)
			appd := gappm.Appdef{}
			args := extractArgs(v)
			appd.Path = args[0]
			k2 := "update-" + k[5:]
			if len(cfgm[k2]) > 0 {
				appd.UpdatePath = cfgm[k2]
				if strings.HasPrefix(appd.UpdatePath, "\"") {
					appd.UpdatePath = appd.UpdatePath[1:]
				}
				if strings.HasSuffix(appd.UpdatePath, "\"") {
					appd.UpdatePath = appd.UpdatePath[:len(appd.UpdatePath)-1]
				}
			}
			// cron (time)
			k2 = "time-" + k[5:]
			cfgtime := cfgm[k2]
			if len(cfgtime) > 0 {
				be := strings.Split(cfgtime, " ")
				h0, m0, ok := gappm.ParseHM(be[0])
				if !ok || len(be) != 2 {
					fmt.Println("INVALID TIME FORMAT [", k[5:], "]: ", cfgtime)
				} else {
					h1, m1, ok := gappm.ParseHM(be[1])
					if !ok {
						fmt.Println("INVALID TIME FORMAT [", k[5:], "]: ", be[1])
					} else {
						appd.Cron.UseTime = true
						appd.Cron.StartHour = h0
						appd.Cron.StartMinute = m0
						appd.Cron.StopHour = h1
						appd.Cron.StopMinute = m1
					}
				}
			}
			appd.Args = args[1:]
			apps = append(apps, appd)
		}
	}
	for k := range apps {
		log.Println("APP[2] [", k, "] ", apps[k].Path)
		_, fn := path.Split(apps[k].Path)
		apps[k].LogPath = path.Join(cfgm["logpath"], fn+".log")
		if cfgm["beep_on_failure"] == "true" {
			apps[k].BeepOnFailure = true
		}
		if cfgm["mirror_app_logs"] == "true" {
			apps[k].Writer2 = os.Stdout
		}
	}
	if len(cfgm["restart_command"]) > 1 {
		cfgm["restart_command"] = strings.Replace(cfgm["restart_command"], "\n", "\r\n", -1)
	}
}

func findConfig() string {
	d, _ := path.Split(os.Args[0])
	if _, er2 := os.Stat(path.Join(d, "gappm.main.conf")); er2 == nil {
		return path.Join(d, "gappm.main.conf")
	}
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
