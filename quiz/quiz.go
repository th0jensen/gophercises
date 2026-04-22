package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"
)

type question struct {
	prompt string
	answer string
}

func readCSVFile(filePath string) ([]question, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	entries, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	records := make([]question, len(entries))
	for i, entry := range entries {
		records[i] = question{entry[0], entry[1]}
	}

	return records, nil
}

func main() {
	var (
		filePath      = flag.String("i", "./quiz/problems.csv", "Input file path (ex ./problems.csv)")
		timerDuration = flag.String("t", "", "Quiz Timer (ex. 30s)")
		shuffle       = flag.String("s", "", "Shuffle Questions? (ex. false)")
	)

	flag.Parse()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	reader := bufio.NewReader(os.Stdin)
	questions, err := readCSVFile(*filePath)
	if err != nil {
		log.Fatalf("Unable to read CSV file: %s", err)
	}

	if strings.ToLower(*shuffle) == "true" {
		rand.Shuffle(len(questions), func(i, j int) { questions[i], questions[j] = questions[j], questions[i] })
	}

	ctx := context.Background()
	var cancel context.CancelFunc

	if *timerDuration != "" {
		fmt.Printf("Ready to start? You have %s...", *timerDuration)
		reader.ReadString('\n')
		duration, err := time.ParseDuration(*timerDuration)
		if err != nil {
			log.Fatal("Failed to parse timer duration!")
		}
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	} else {
		fmt.Printf("Ready to start? You have no time limit...")
		reader.ReadString('\n')
	}

	var score int
	for i, question := range questions {
		answerCh := make(chan string, 1)

		go func() {
			text, _ := reader.ReadString('\n')
			answerCh <- strings.TrimSpace(text)
		}()

		fmt.Printf("Problem #%d: %s = ", i+1, question.prompt)

		select {
		case <-ctx.Done():
			fmt.Printf("\nTimer ran out!\n")
			goto results
		case <-sigCh:
			fmt.Println()
			goto results
		case userAnswer := <-answerCh:
			if strings.EqualFold(userAnswer, question.answer) {
				score++
			}
		}
	}

results:
	fmt.Printf("You scored %d/%d\n", score, len(questions))
	if score == len(questions) {
		fmt.Println("Congratulations, you got a perfect score!")
	}
}
