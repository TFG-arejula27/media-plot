package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

type Log struct {
	Message string  `csv:"Message"`
	Power   float64 `csv:"Average power(Watts)"`
	Time    float64 `csv:"time"`
}

func main() {
	file, err := os.OpenFile("fr-1.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	logs := []*Log{}

	if err := gocsv.UnmarshalFile(file, &logs); err != nil { // Load clients from file
		panic(err)
	}
	logsRealPower := []*Log{}
	for i, log := range logs {
		if log.Message != "idle" {
			newRealLog := &Log{Message: log.Message, Power: log.Power - logs[i-1].Power, Time: log.Time}

			logsRealPower = append(logsRealPower, newRealLog)
		}
	}
	m := make(map[string][]Log)
	for _, log := range logsRealPower {
		m[log.Message] = append(m[log.Message], Log{Power: log.Power, Time: log.Time})

	}
	logsAverageRealPower := []*Log{}
	for k, v := range m {
		logsAverageRealPower = append(logsAverageRealPower, calcAverage(v, k))
	}
	sort.Slice(logsAverageRealPower, func(i, j int) bool {
		msgi := logsAverageRealPower[i].Message
		msgj := logsAverageRealPower[j].Message
		frzi, _ := strconv.Atoi(strings.Split(msgi, "-")[0])
		frzj, _ := strconv.Atoi(strings.Split(msgj, "-")[0])
		erri, _ := strconv.Atoi(strings.Split(msgi, "-")[2])
		errj, _ := strconv.Atoi(strings.Split(msgj, "-")[2])

		return (frzi == frzj && erri < errj) || frzi < frzj

	})
	csvContent, err := gocsv.MarshalString(&logsAverageRealPower)
	fmt.Println(csvContent)
	// Display all clients as CSV string

}

func calcAverage(logs []Log, m string) *Log {
	var pwr float64 = 0
	var time float64 = 0
	for _, log := range logs {
		time += log.Time
		pwr += log.Power

	}
	len := float64(len(logs))
	return &Log{
		Message: m,
		Power:   pwr / len,
		Time:    time / len,
	}

}
