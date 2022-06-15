package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

type Log struct {
	Message   string  `csv:"Message"`
	Power     float64 `csv:"Average power(Watts)"`
	Time      float64 `csv:"time"`
	Threshold int     `csv:"threshold"`
	Frecuenzy int     `csv:"Frecuenzy"`
	Energy    float64 `csv:"Energy"`
}

var ocupacion int

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Println("Input file is missing.")
		os.Exit(1)
	}
	filePath := args[1]
	var err error
	strings.Split(filePath, "-")
	ocupacion, err = strconv.Atoi(strings.Split(filePath, "-")[0])
	if err != nil {
		panic(err)
	}
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = ' '
		return gocsv.NewSafeCSVWriter(writer)
	})

	gocsv.SetCSVReader(func(out io.Reader) gocsv.CSVReader {
		reader := csv.NewReader(out)
		reader.Comma = ';'
		return reader

	})

	logs := []*Log{}

	if err := gocsv.UnmarshalFile(file, &logs); err != nil { // Load clients from file
		panic(err)
	}
	logsRealPower := []*Log{}
	for _, log := range logs {

		if log.Message == "idle" {
			continue
		}
		logsRealPower = append(logsRealPower, log)
		/*if log.Message != "idle" {

			newRealLog := &Log{Message: log.Message, Power: log.Power - logs[i-1].Power, Time: log.Time}

			logsRealPower = append(logsRealPower, newRealLog)
		}*/
	}
	m := make(map[string][]Log)
	for _, log := range logsRealPower {
		m[log.Message] = append(m[log.Message], Log{Power: log.Power, Time: log.Time})

	}
	logsAverageRealPower := []*Log{}
	for k, v := range m {

		log := calcAverage(v, k)
		msg := log.Message
		frz, _ := strconv.Atoi(strings.Split(msg, "-")[0])
		th, _ := strconv.Atoi(strings.Split(msg, "-")[2])
		log.Threshold = th
		log.Frecuenzy = frz
		log.Energy = log.Time * log.Power
		logsAverageRealPower = append(logsAverageRealPower, log)
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
	writeFile(csvContent, "fr-"+filePath)
	sort.Slice(logsAverageRealPower, func(i, j int) bool {
		msgi := logsAverageRealPower[i].Message
		msgj := logsAverageRealPower[j].Message
		frzi, _ := strconv.Atoi(strings.Split(msgi, "-")[0])
		frzj, _ := strconv.Atoi(strings.Split(msgj, "-")[0])
		erri, _ := strconv.Atoi(strings.Split(msgi, "-")[2])
		errj, _ := strconv.Atoi(strings.Split(msgj, "-")[2])

		return (erri == errj && frzi < frzj) || erri < errj

	})

	csvContent, err = gocsv.MarshalString(&logsAverageRealPower)

	writeFile(csvContent, "th-"+filePath)

	mtx := createMatrix(logsAverageRealPower)
	matrixContentfile := matrixToString(*mtx)
	writeFile(matrixContentfile, "plot-power-"+filePath)

	mtxP := createMatrixPower(logsAverageRealPower)
	matrixPContentfile := matrixToString(*mtxP)
	writeFile(matrixPContentfile, "plot-power-"+filePath)

	mtxT := createMatrixThrogput(logsAverageRealPower)
	matrixTContentfile := matrixToString(*mtxT)
	writeFile(matrixTContentfile, "plot-throgput-"+filePath)
}

func matrixToString(mtx [][]float64) string {

	content := ""
	for i := 0; i < len(mtx); i++ {
		fr := 0.8 + float64(i)*0.1

		content += strconv.FormatFloat(fr, 'f', 4, 64) + " " + " "

		for j := 0; j < len(mtx[0]); j++ {
			content += strconv.FormatFloat(mtx[i][j], 'f', 4, 64) + " "
		}
		content += "\n"
	}
	return content
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

func writeFile(data, outputFile string) error {
	f, err := os.Create("means-" + outputFile)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err2 := f.WriteString(data)

	if err2 != nil {
		return err2
	}
	return nil

}

func createMatrix(data []*Log) *[][]float64 {

	freqInterval := getFreqInterval(data)
	numFrequencies := len(freqInterval)
	thresholdInterval := getThresholdInterval(data)
	numThreshold := len(thresholdInterval)

	matrix := make([][]float64, numFrequencies)
	for i := 0; i < numFrequencies; i++ {
		matrix[i] = make([]float64, numThreshold)
	}

	for _, v := range data {
		fr := freqInterval[v.Frecuenzy]
		th := thresholdInterval[v.Threshold]
		matrix[fr][th] = v.Energy
	}

	return &matrix

}

func createMatrixPower(data []*Log) *[][]float64 {

	freqInterval := getFreqInterval(data)
	numFrequencies := len(freqInterval)
	thresholdInterval := getThresholdInterval(data)
	numThreshold := len(thresholdInterval)

	matrix := make([][]float64, numFrequencies)
	for i := 0; i < numFrequencies; i++ {
		matrix[i] = make([]float64, numThreshold)
	}

	for _, v := range data {
		fr := freqInterval[v.Frecuenzy]
		th := thresholdInterval[v.Threshold]
		matrix[fr][th] = v.Power
	}

	return &matrix

}

func createMatrixThrogput(data []*Log) *[][]float64 {

	freqInterval := getFreqInterval(data)
	numFrequencies := len(freqInterval)
	thresholdInterval := getThresholdInterval(data)
	numThreshold := len(thresholdInterval)

	matrix := make([][]float64, numFrequencies)
	for i := 0; i < numFrequencies; i++ {
		matrix[i] = make([]float64, numThreshold)
	}

	for _, v := range data {
		fr := freqInterval[v.Frecuenzy]
		th := thresholdInterval[v.Threshold]
		matrix[fr][th] = float64(ocupacion) / v.Time
	}

	return &matrix

}

func valueOnSlice(v int, s []int) bool {
	for _, e := range s {
		if v == e {
			return true
		}
	}
	return false
}

func generatePositionMap(values []int) map[int]int {

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	res := make(map[int]int)
	for i, v := range values {
		res[v] = i
	}

	return res

}

func getFreqInterval(data []*Log) map[int]int {
	values := []int{}

	for _, l := range data {
		if !valueOnSlice(l.Frecuenzy, values) {
			values = append(values, l.Frecuenzy)

		}

	}
	return generatePositionMap(values)

}

func getThresholdInterval(data []*Log) map[int]int {
	values := []int{}
	for _, l := range data {
		if !valueOnSlice(l.Threshold, values) {
			values = append(values, l.Threshold)

		}

	}

	return generatePositionMap(values)
}
