package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

type Reminder struct {
	Time        time.Time `json:"time"`
	Message     string    `json:"message"`
	IsPomodoro  bool      `json:"is_pomodoro"`
	PomodoroNum int       `json:"pomodoro_num"`
}

const (
	SockAddr           = "/tmp/remindme.sock"
	UrgencyLow         = "low"
	UrgencyNormal      = "normal"
	UrgencyCritical    = "critical"
	BreakMsg           = "5 minute break"
	BreakFinishedMsg   = "break finished"
	LongBreakMsg       = "20 minute break"
	PomodoroStoppedMsg = "stopped"
	PomodoroStartedMsg = "started"
)

var (
	data   = []Reminder{}
	pCount = 0
	pNum   = 0
)

func main() {
	args := os.Args[1:]

	switch args[0] {
	case "in":
		send(Reminder{
			Time:    parseTimeIn(args[1]),
			Message: strings.Join(args[2:], " "),
		})
	case "at":
		send(Reminder{
			Time:    parseTimeAt(args[1]),
			Message: strings.Join(args[2:], " "),
		})
	case "p":
		switch args[1] {
		case "start":
			send(Reminder{
				Time:       time.Now(),
				Message:    PomodoroStartedMsg,
				IsPomodoro: true,
			})
		case "stop":
			send(Reminder{
				Time:       time.Now(),
				Message:    PomodoroStoppedMsg,
				IsPomodoro: true,
			})
		default:
			fmt.Println("use 'start' or 'stop'")
		}
	case "--watch":
		watch()
	default:
		fmt.Println("use 'in' or 'at'. Pomodoro with 'p start/stop'")
	}
}

func send(r Reminder) {
	b, err := json.Marshal(&r)
	if err != nil {
		log.Panic(err)
	}

	conn, err := net.Dial("unix", SockAddr)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	_, err = conn.Write(b)
	if err != nil {
		log.Panic(err)
	}
}

func watch() {
	os.Remove(SockAddr)

	go receive()

	for {
		time.Sleep(time.Second * 1)

		now := time.Now()

		for k, v := range data {
			if now.After(v.Time) {
				switch v.Message {
				case PomodoroStartedMsg:
					if v.PomodoroNum == pNum {
						pCount = 1

						data = append(data, Reminder{
							Time:        now.Add(time.Second * 25),
							Message:     BreakMsg,
							IsPomodoro:  true,
							PomodoroNum: pNum,
						})

						notify("Pomodoro", v.Message, UrgencyNormal)
					}
				case BreakMsg:
					if v.PomodoroNum == pNum {
						pause := 5
						message := BreakMsg

						if pCount == 4 {
							message = LongBreakMsg
							pause = 20
						}

						data = append(data, Reminder{
							Time:        now.Add(time.Second * time.Duration(pause)),
							Message:     BreakFinishedMsg,
							IsPomodoro:  true,
							PomodoroNum: v.PomodoroNum,
						})

						notify("Pomodoro", message, UrgencyCritical)
					}
				case BreakFinishedMsg:
					if v.PomodoroNum == pNum {
						pCount++

						if pCount > 4 {
							pCount = 1
						}

						data = append(data, Reminder{
							Time:        now.Add(time.Second * 25),
							Message:     BreakMsg,
							IsPomodoro:  true,
							PomodoroNum: v.PomodoroNum,
						})

						notify("Pomodoro", v.Message, UrgencyCritical)
					}
				case PomodoroStoppedMsg:
					pNum++
					notify("Pomodoro", v.Message, UrgencyNormal)
				default:
					if strings.HasSuffix(v.Message, "!") {
						notify("Reminder", strings.TrimSuffix(v.Message, "!"), UrgencyCritical)
					} else {
						notify("Reminder", v.Message, UrgencyNormal)
					}
				}

				data = slices.Delete(data, k, k+1)
			}
		}
	}
}

// urgency: low, normal, critical
func notify(title, message, urgency string) {
	cmd := exec.Command("notify-send", title, message, fmt.Sprintf("--urgency=%s", urgency), "--app-name=RemindMe")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Panic(string(out))
	}
}

func receive() {
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: SockAddr})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.AcceptUnix()
		if err != nil {
			log.Panic(err)
		}

		b := make([]byte, 1024)
		i, err := conn.Read(b)
		if err != nil {
			log.Panic(err)
		}

		var r Reminder
		err = json.Unmarshal(b[:i], &r)
		if err != nil {
			log.Panic(err)
		}

		data = append(data, r)
	}
}

func parseTimeAt(val string) time.Time {
	now := time.Now()

	parts := strings.Split(val, ":")

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}

	t := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	return t
}

var unitMap = map[string]time.Duration{
	"h": time.Hour,
	"m": time.Minute,
	"s": time.Second,
}

func addTime(val, unit string) time.Time {
	val = strings.TrimSuffix(val, unit)

	toAdd, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}

	now := time.Now()
	now = now.Add(unitMap[unit] * time.Duration(toAdd))

	return now
}

func parseTimeIn(val string) time.Time {
	switch {
	case strings.HasSuffix(val, "h"):
		return addTime(val, "h")
	case strings.HasSuffix(val, "m"):
		return addTime(val, "m")
	case strings.HasSuffix(val, "s"):
		return addTime(val, "s")
	default:
		panic("can't parse time")
	}
}
