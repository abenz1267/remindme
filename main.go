package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
	"golang.org/x/exp/slices"
)

var formats = map[string]string{
	"h": "4",
	"m": "0",
}

type Reminder struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

func main() {
	args := os.Args[1:]

	switch args[0] {
	case "in":
		r := Reminder{
			Time:    parseTimeIn(args[1]),
			Message: strings.Join(args[2:], " "),
		}

		addToJson(r)
	case "at":
		r := Reminder{
			Time:    parseTimeAt(args[1]),
			Message: strings.Join(args[2:], " "),
		}

		addToJson(r)
	case "--watch":
		watch()
	default:
		fmt.Println("use 'in' or 'at'")
	}
}

func watch() {
	deleteOld()

	for {
		time.Sleep(time.Minute * 1)

		_, file := paths()

		b, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		rs := []Reminder{}

		err = json.Unmarshal(b, &rs)
		if err != nil {
			panic(err)
		}

		for k, v := range rs {
			now := time.Now()

			if now.After(v.Time) || now.Equal(v.Time) {
				err := beeep.Notify("Remindme", v.Message, "assets/information.png")
				if err != nil {
					log.Println(err)
				}

				rs = slices.Delete(rs, k, k+1)
			}
		}

		b, err = json.Marshal(&rs)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(file, b, PermFile)
		if err != nil {
			panic(err)
		}
	}
}

func deleteOld() {
	_, file := paths()

	b, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	rs := []Reminder{}

	err = json.Unmarshal(b, &rs)
	if err != nil {
		panic(err)
	}

	for k, v := range rs {
		now := time.Now()

		if now.After(v.Time) {
			rs = slices.Delete(rs, k, k+1)
		}
	}

	b, err = json.Marshal(&rs)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(file, b, PermFile)
	if err != nil {
		panic(err)
	}
}

const (
	PermFile   = 0o644
	PermFolder = 0o755
)

func paths() (string, string) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(home, ".remindme")
	file := filepath.Join(dir, "reminders.json")

	return dir, file
}

func addToJson(r Reminder) {
	dir, file := paths()

	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, PermFolder)
		if err != nil {
			panic(err)
		}
	}

	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		rs := []Reminder{}

		b, err := json.Marshal(&rs)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile(file, b, PermFile)
		if err != nil {
			panic(err)
		}
	}

	rs := []Reminder{}

	b, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(b, &rs)
	if err != nil {
		panic(err)
	}

	rs = append(rs, r)

	b, err = json.Marshal(&rs)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(file, b, PermFile)
	if err != nil {
		panic(err)
	}
}

func parseTimeAt(val string) time.Time {
	time, err := time.Parse("15:04", val)
	if err != nil {
		panic(err)
	}

	return time
}

func parseTimeIn(val string) time.Time {
	switch {
	case strings.HasSuffix(val, "h"):
		val = strings.TrimSuffix(val, "h")

		toAdd, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		now := time.Now()
		now = now.Add(time.Hour * time.Duration(toAdd))

		return now
	case strings.HasSuffix(val, "m"):
		val = strings.TrimSuffix(val, "m")

		toAdd, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		now := time.Now()
		now = now.Add(time.Minute * time.Duration(toAdd))

		return now
	default:
		panic("can't parse time")
	}
}
