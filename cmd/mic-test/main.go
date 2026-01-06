package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

// AudioRecorder provides simple microphone testing
type AudioRecorder struct {
	sampleRate    int
	channels      int
	bitsPerSample int
	device        string
}

func NewRecorder() *AudioRecorder {
	return &AudioRecorder{
		sampleRate:    16000,
		channels:      1,
		bitsPerSample: 16,
		device:        "default",
	}
}

func (r *AudioRecorder) Record(duration time.Duration, outputFile string) error {
	log.Printf("🎤 Starting microphone test recording...")
	log.Printf("   Duration: %v", duration)
	log.Printf("   Sample rate: %d Hz", r.sampleRate)
	log.Printf("   Channels: %d", r.channels)
	log.Printf("   Bits per sample: %d", r.bitsPerSample)
	log.Printf("   Output: %s", outputFile)

	// Calculate buffer size
	bytesPerSample := r.bitsPerSample / 8
	bytesPerSecond := r.sampleRate * r.channels * bytesPerSample
	totalBytes := int(float64(bytesPerSecond) * duration.Seconds())

	log.Printf("   Expected size: %d bytes (~%.2f seconds)", totalBytes, float64(totalBytes)/float64(bytesPerSecond))

	// Check for available recording tools
	log.Println("\n📋 Checking recording tools...")

	// Try arecord (ALSA)
	if r.isToolAvailable("arecord") {
		log.Println("   ✓ Using arecord (ALSA)")
		return r.recordWithArecord(duration, outputFile)
	}

	// Try rec (sox)
	if r.isToolAvailable("rec") {
		log.Println("   ✓ Using rec (sox)")
		return r.recordWithRec(duration, outputFile)
	}

	// Try pactl (PulseAudio)
	if os.Getenv("PULSE_SERVER") != "" {
		log.Println("   ✓ Using pactl (PulseAudio)")
		return r.recordWithPactl(duration, outputFile)
	}

	log.Println("   ✗ No recording tool found!")
	log.Println("\n💡 Install recording tools:")
	log.Println("   sudo apt install alsa-utils    # for arecord")
	log.Println("   sudo apt install sox           # for rec")
	log.Println("   sudo apt install libpulse-dev  # for PulseAudio")

	return fmt.Errorf("no recording tool available")
}

func (r *AudioRecorder) recordWithArecord(duration time.Duration, outputFile string) error {
	log.Println("\n🔴 Recording with arecord...")

	args := []string{
		"-D", r.device,
		"-f", "S16_LE",
		"-r", fmt.Sprintf("%d", r.sampleRate),
		"-c", fmt.Sprintf("%d", r.channels),
		"-d", fmt.Sprintf("%d", int(duration.Seconds())),
		"-q", // quiet mode
		outputFile,
	}

	cmd := exec.Command("arecord", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("   arecord output: %s", string(output))
		return fmt.Errorf("arecord failed: %w", err)
	}

	// Check file size
	info, err := os.Stat(outputFile)
	if err != nil {
		return fmt.Errorf("failed to check output file: %w", err)
	}

	log.Printf("   ✓ Recording complete!")
	log.Printf("   📁 File size: %d bytes (%.2f KB)", info.Size(), float64(info.Size())/1024)

	return nil
}

func (r *AudioRecorder) recordWithRec(duration time.Duration, outputFile string) error {
	log.Println("\n🔴 Recording with rec (sox)...")

	args := []string{
		"-r", fmt.Sprintf("%d", r.sampleRate),
		"-c", fmt.Sprintf("%d", r.channels),
		"-b", fmt.Sprintf("%d", r.bitsPerSample),
		"-e", "signed-integer",
		"-d", duration.String(),
		outputFile,
		"silence",
		"1", "0.1", "1%", // remove silence
	}

	cmd := exec.Command("rec", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("   rec output: %s", string(output))
		return fmt.Errorf("rec failed: %w", err)
	}

	info, err := os.Stat(outputFile)
	if err != nil {
		return fmt.Errorf("failed to check output file: %w", err)
	}

	log.Printf("   ✓ Recording complete!")
	log.Printf("   📁 File size: %d bytes (%.2f KB)", info.Size(), float64(info.Size())/1024)

	return nil
}

func (r *AudioRecorder) recordWithPactl(duration time.Duration, outputFile string) error {
	log.Println("\n🔴 Recording with pactl (PulseAudio)...")

	log.Println("   Note: PulseAudio recording requires additional setup")
	log.Println("   Using arecord as fallback...")

	return r.recordWithArecord(duration, outputFile)
}

func (r *AudioRecorder) Play(file string) error {
	log.Printf("\n🔊 Playing audio file: %s", file)

	if !r.isToolAvailable("aplay") {
		log.Println("   ✗ aplay not found. Install: sudo apt install alsa-utils")
		return fmt.Errorf("aplay not available")
	}

	cmd := exec.Command("aplay", "-D", r.device, file)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("   aplay output: %s", string(output))
		return fmt.Errorf("playback failed: %w", err)
	}

	log.Println("   ✓ Playback complete!")
	return nil
}

func (r *AudioRecorder) Info(file string) error {
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("file not found: %s", file)
	}

	log.Printf("\n📊 Audio file info:")
	log.Printf("   Path: %s", file)
	log.Printf("   Size: %d bytes (%.2f KB)", info.Size(), float64(info.Size())/1024)

	if r.isToolAvailable("file") {
		cmd := exec.Command("file", file)
		output, _ := cmd.Output()
		log.Printf("   Type: %s", string(output))
	}

	if r.isToolAvailable("sox") {
		cmd := exec.Command("sox", "--i", file)
		output, _ := cmd.Output()
		log.Printf("   Details:\n%s", string(output))
	}

	return nil
}

func (r *AudioRecorder) isToolAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func (r *AudioRecorder) ListDevices() error {
	log.Println("\n🎛️  Available audio input devices:")

	if r.isToolAvailable("arecord") {
		cmd := exec.Command("arecord", "-l")
		output, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Println(string(output))
			return nil
		}
	}

	if r.isToolAvailable("aplay") {
		cmd := exec.Command("aplay", "-l")
		output, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Println(string(output))
			return nil
		}
	}

	log.Println("   No devices found")
	log.Println("   Install: sudo apt install alsa-utils")
	return nil
}

func main() {
	// Parse flags
	durationFlag := flag.Duration("duration", 3*time.Second, "Recording duration (e.g., 3s, 5s, 1m)")
	outputFlag := flag.String("output", "test_recording.wav", "Output file name")
	playFlag := flag.Bool("play", false, "Play the recording after recording")
	infoFlag := flag.Bool("info", false, "Show info about the output file")
	listFlag := flag.Bool("list", false, "List available audio devices")
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          🎤 VoiceType Microphone Test Tool              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	recorder := NewRecorder()

	// List devices if requested
	if *listFlag {
		recorder.ListDevices()
		os.Exit(0)
	}

	// Show info if requested
	if *infoFlag {
		recorder.Info(*outputFlag)
		os.Exit(0)
	}

	// Record
	ctx, cancel := context.WithTimeout(context.Background(), *durationFlag+5*time.Second)
	defer cancel()

	done := make(chan bool, 1)

	go func() {
		err := recorder.Record(*durationFlag, *outputFlag)
		if err != nil {
			log.Printf("❌ Recording failed: %v", err)
			done <- false
			return
		}
		done <- true
	}()

	// Show progress
	log.Println("\n⏳ Recording in progress...")
	log.Println("   (Speak into your microphone now)")

	select {
	case <-ctx.Done():
		log.Println("❌ Recording timed out")
	case success := <-done:
		if !success {
			os.Exit(1)
		}
	}

	// Show info
	recorder.Info(*outputFlag)

	// Play if requested
	if *playFlag {
		recorder.Play(*outputFlag)
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ✅ Test Complete!                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Play the recording to verify your microphone works")
	fmt.Println("  2. If you hear your voice, microphone is working!")
	fmt.Println("  3. Run VoiceType: ./VoiceType")
	fmt.Println()
}
