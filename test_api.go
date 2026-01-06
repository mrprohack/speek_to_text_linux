package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"VoiceType/internal/api"
	"VoiceType/pkg/errors"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_api.go <audio_file>")
		os.Exit(1)
	}

	audioFile := os.Args[1]
	data, err := os.ReadFile(audioFile)
	if err != nil {
		log.Fatalf("Failed to read audio file: %v", err)
	}

	fmt.Printf("Read %d bytes from %s\n", len(data), audioFile)

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: GROQ_API_KEY not set")
	}

	errHandler := errors.NewHandler()
	client := api.NewClient(apiKey, errHandler)

	fmt.Printf("Using model: %s\n", client.GetModel())
	fmt.Println("Transcribing...")

	ctx := context.Background()
	text, err := client.Transcribe(ctx, data)
	if err != nil {
		log.Fatalf("Transcription failed: %v", err)
	}

	fmt.Printf("\n✓ Transcription successful!\n")
	fmt.Printf("Result: %s\n", text)
}
