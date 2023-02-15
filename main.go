package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"golang.org/x/exp/slices"
)

type Reminder struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

const SockAddr = "/tmp/remindme.sock"

var data = []Reminder{}

func main() {
	args := os.Args[1:]

	switch args[0] {
	case "in":
		r := Reminder{
			Time:    parseTimeIn(args[1]),
			Message: strings.Join(args[2:], " "),
		}

		b, err := json.Marshal(&r)
		if err != nil {
			log.Panic(err)
		}

		send(b)
	case "at":
		r := Reminder{
			Time:    parseTimeAt(args[1]),
			Message: strings.Join(args[2:], " "),
		}

		b, err := json.Marshal(&r)
		if err != nil {
			log.Panic(err)
		}

		send(b)
	case "--watch":
		watch()
	default:
		fmt.Println("use 'in' or 'at'")
	}
}

func send(b []byte) {
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
				if strings.HasSuffix(v.Message, "!") {
					err := beeep.Alert("Reminder", strings.TrimSuffix(v.Message, "!"), "assets/warning.png")
					if err != nil {
						panic(err)
					}
				} else {
					err := beeep.Notify("Reminder", v.Message, "assets/information.png")
					if err != nil {
						panic(err)
					}
				}

				data = slices.Delete(data, k, k+1)
			}
		}
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
