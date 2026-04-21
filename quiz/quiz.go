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

func readCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
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
	quiz, err := readCSVFile(*filePath)
	if err != nil {
		log.Fatalf("Unable to read CSV file: %v", err)
	}

	if strings.ToLower(*shuffle) == "true" {
		rand.Shuffle(len(quiz), func(i, j int) { quiz[i], quiz[j] = quiz[j], quiz[i] })
	}

	ctx := context.Background()
	var cancel context.CancelFunc

	if *timerDuration != "" {
		fmt.Printf("Ready to start? You have %v...", *timerDuration)
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
	for i, entry := range quiz {
		answerCh := make(chan string, 1)

		go func() {
			text, _ := reader.ReadString('\n')
			answerCh <- strings.TrimSpace(text)
		}()

		fmt.Printf("Problem #%v: %v = ", i+1, entry[0])

		select {
		case <-ctx.Done():
			fmt.Printf("\nTimer ran out!\n")
			goto results
		case <-sigCh:
			fmt.Println()
			goto results
		case userAnswer := <-answerCh:
			if strings.EqualFold(userAnswer, entry[1]) {
				score++
			}
		}
	}

results:
	fmt.Printf("You scored %v/%v\n", score, len(quiz))
	if score == len(quiz) {
		fmt.Printf("Congratulations, you got a perfect score!")
	}
}
