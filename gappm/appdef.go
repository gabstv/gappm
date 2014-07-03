package gappm

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	Cron          AppDefCron
	Writer2       io.Writer
	//
	lf *os.File
}

type AppDefCron struct {
	UseTime bool
	UseDay  bool
	//
	StartHour   int
	StartMinute int
	StopHour    int
	StopMinute  int
	//
	StartDays int
	Entries   []TzEntry
}

func dayOfWeek() int {
	return int(time.Now().Weekday())
}

type TzEntry struct {
	Day         int
	StartHour   int
	StartMinute int
	StopHour    int
	StopMinute  int
}

type KV map[string]string

func (a *Appdef) DailyLogPath() string {
	n := time.Now()
	tail := "_" + strconv.Itoa(n.Year()) + "-" + strconv.Itoa(int(n.Month())) + "-" + strconv.Itoa(n.Day()) + ".log"
	return strings.Replace(a.LogPath, ".log", tail, 1)
}

func (a *Appdef) oldLogPath(days int) string {
	n := time.Now()
	n.Add(time.Hour * -24 * time.Duration(days))
	tail := "_" + strconv.Itoa(n.Year()) + "-" + strconv.Itoa(int(n.Month())) + "-" + strconv.Itoa(n.Day()) + ".log"
	return strings.Replace(a.LogPath, ".log", tail, 1)
}

func (a *Appdef) Run() {
	// CHECK CRON
	if !a.CronTest() {
		return
	}
	if a.Writer2 == nil {
		// /dev/null
		a.Writer2 = ioutil.Discard
	}
	a.Command = exec.Command(a.Path, a.Args...)
	// check log file
	if _, err := os.Stat(a.DailyLogPath()); err != nil {
		a.lf, err = os.OpenFile(a.DailyLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(a.Path, a.DailyLogPath(), err)
		}
	} else {
		a.lf, _ = os.OpenFile(a.DailyLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	w := io.MultiWriter(a.Writer2, a.lf)
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

func (a *Appdef) ReLog(ldays int) {

	buff := new(bytes.Buffer)
	w := io.MultiWriter(a.Writer2, buff)
	if a.Command != nil {
		a.Command.Stdout = w
		a.Command.Stderr = w
	}

	a.lf.Close()
	a.lf, _ = os.OpenFile(a.DailyLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	w = io.MultiWriter(a.Writer2, a.lf)
	if a.Command != nil {
		a.Command.Stdout = w
		a.Command.Stderr = w
	}

	a.lf.Write(buff.Bytes())
	//
	// gambits do Elio
	if ldays > 0 {
		for i := ldays + 1; i < ldays+60; i++ {
			if _, err := os.Stat(a.oldLogPath(i)); err == nil {
				os.Remove(a.oldLogPath(i))
			}
		}
	}
}

func (a *Appdef) CronTest() bool {
	if !a.Cron.UseTime {
		return true
	}
	now := time.Now()
	nowhms := now.Hour()*60 + now.Minute()
	if a.Cron.Entries != nil {
		today := dayOfWeek()
		for _, v := range a.Cron.Entries {
			if v.Day == today {
				startms := v.StartHour*60 + v.StartMinute
				stopms := v.StopHour*60 + v.StopMinute
				if startms > stopms {
					if nowhms >= startms || nowhms < stopms {
						return true
					}
				} else {
					if nowhms >= startms && nowhms < stopms {
						return true
					}
				}
			}
		}
	} else {
		startms := a.Cron.StartHour*60 + a.Cron.StartMinute
		stopms := a.Cron.StopHour*60 + a.Cron.StopMinute
		if startms > stopms {
			if nowhms >= startms || nowhms < stopms {
				return true
			}
		} else {
			if nowhms >= startms && nowhms < stopms {
				return true
			}
		}
	}
	return false
}

func (a *Appdef) IsRunning() bool {
	if a.Command == nil {
		return false
	}
	if a.Command.Process == nil {
		return false
	}
	return true
}

func ParseHM(str string) (hour, minute int, ok bool) {
	ok = false
	hm := strings.Split(str, ":")
	if len(hm) < 2 {
		return
	}
	str_hour := strings.TrimSpace(hm[0])
	str_minute := strings.TrimSpace(hm[1])
	h0, err := strconv.ParseInt(str_hour, 10, 32)
	if err != nil {
		return
	}
	hour = int(h0)

	m0, err := strconv.ParseInt(str_minute, 10, 32)
	if err != nil {
		return
	}
	minute = int(m0)
	ok = true
	return
}
