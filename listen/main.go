package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/CyCoreSystems/goertzel"
	"github.com/gordonklaus/portaudio"
)

func main() {
	const sampleRate = 44100
	const seconds = 3
	const channels = 1
	const chunkSize = 1000 // number of samples per read

	portaudio.Initialize()
	defer portaudio.Terminate()

	chunk := make([]int16, chunkSize)
	stream, err := portaudio.OpenDefaultStream(channels, 0, sampleRate, len(chunk), chunk)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	devs, _ := portaudio.Devices()
	for _, dev := range devs {
		println("DEV: ", dev.Name)
	}
	fmt.Println("Recording...")
	if err := stream.Start(); err != nil {
		panic(err)
	}

	audioChan := make(chan []int16, 100000)
	go processAudio(audioChan, sampleRate)

	// var fullBuffer []int16
	// start := time.Now()
	for { // time.Since(start) < time.Second*time.Duration(seconds) {
		if err := stream.Read(); err != nil {
			panic(err)
		}
		audioChan <- chunk
		// fullBuffer = append(fullBuffer, chunk...) // collect samples
	}

	fmt.Println("stopping...")
	if err := stream.Stop(); err != nil {
		panic(err)
	}
	fmt.Println("Done recording.")

	// Write to WAV file
	// outFile, err := os.Create("output.wav")
	// if err != nil {
	// 	panic(err)
	// }
	// defer outFile.Close()

	// encoder := wav.NewEncoder(outFile, sampleRate, 16, channels, 1)
	// intBuf := &audio.IntBuffer{
	// 	Data:           make([]int, len(fullBuffer)),
	// 	Format:         &audio.Format{SampleRate: sampleRate, NumChannels: channels},
	// 	SourceBitDepth: 16,
	// }
	// for i, v := range fullBuffer {
	// 	intBuf.Data[i] = int(v)
	// }
	// if err := encoder.Write(intBuf); err != nil {
	// 	panic(err)
	// }
	// encoder.Close()
	// fmt.Println("Saved to output.wav")
}

func processBlocks(in <-chan *goertzel.BlockSummary) {
	println("waiting for blocks")
	for {
		sum := <-in

		if sum != nil && sum.Present {
			println(time.Now().Nanosecond(), "I HEAR IT!")
		}
	}
}

func writeInt16Slice(w io.Writer, data []int16) error {
	return binary.Write(w, binary.LittleEndian, data)
}

func processAudio(in <-chan []int16, sampleRate int) {
	targetFreq := 697.0
	minDur := 20 * time.Millisecond

	target := goertzel.NewTarget(targetFreq, float64(sampleRate), minDur)
	// target.UseOptimized = true

	go processBlocks(target.Blocks())

	r, w := io.Pipe()

	go func() {
		for {
			bytes := <-in

			if err := writeInt16Slice(w, bytes); err != nil {
				panic("PANIC writting slice to pipe")
			}
		}
	}()
	println("READ START")
	err := target.Read(r)
	if err != nil {
		panic("READ PANIC")
	}
	println("READ END")

	time.Sleep(10 * time.Second)
}
