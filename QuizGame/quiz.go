package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	fileName := flag.String("filename", "problems.csv", "question answer file name")
	timeLimit := flag.Int("timeLimit", 5, "time limit for the quiz in seconds")

	lines := readFileContent(*fileName)
	scoreUser(lines, *timeLimit)
}

func readFileContent(fileName string) [][]string {
	// Default file is problem.csv in the current folder.
	// Users can specify a different location
	flag.Parse()

	// open the file
	csvFile, fileOpenErr := os.Open(fileName)
	if fileOpenErr != nil {
		exit(fmt.Sprintf("An error occured while opening the file: %v", fileOpenErr))
	}
	r := csv.NewReader(csvFile)
	lines, fileReadErr := r.ReadAll()
	if fileReadErr != nil {
		exit(fmt.Sprintf("An error occured while reading the file: %v", fileReadErr))
	}
	return lines
}

func scoreUser(lines [][]string, timelimit int) {
	correct := 0
	timer := time.NewTimer(time.Duration(timelimit) * time.Second)
loop:
	for idx, line := range lines {
		question, answer := line[0], line[1]
		answer = strings.TrimSpace(answer)
		fmt.Printf("Problem #%d: %s = ", idx+1, question)

		answerCh := make(chan string)
		go func() {
			var userAnswer string
			fmt.Scanf("%s\n", &userAnswer)
			userAnswer = strings.TrimSpace(userAnswer)
			answerCh <- userAnswer
		}()

		select {
		case <-timer.C:
			fmt.Printf("\nTime expired\n")
			break loop
		case ans := <-answerCh:
			if ans == answer {
				correct++
			}
		}
	}
	fmt.Printf("You scored %d out of %d.\n", correct, len(lines))
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
