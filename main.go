package main

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

type Log struct {
	Message string  `csv:"Message"`
	Power   float64 `csv:"Average power(Watts)"`
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
	for _, client := range logs {
		fmt.Println("line", client.Power)
	}

	if _, err := file.Seek(0, 0); err != nil { // Go to the start of the file
		panic(err)
	}

	fmt.Println(logs[0]) // Display all clients as CSV string

}
