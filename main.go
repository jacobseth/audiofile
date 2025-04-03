package main

import (
	"io"
	"time"

	"github.com/ebitengine/oto/v3"
)

func main() {
	const (
		sampleRate      = 44100
		channelCount    = 1
		bitDepthInBytes = 2 // 16-bit PCM
		duration        = 3 * time.Second

		clockFreq = 10.0
		dataFreq  = 20
		amplitude = 16000
	)

	op := &oto.NewContextOptions{}
	op.SampleRate = sampleRate
	op.ChannelCount = 1
	op.Format = oto.FormatSignedInt16LE

	ctx, ready, err := oto.NewContext(op)
	if err != nil {
		panic(err)
	}
	<-ready

	r, w := io.Pipe()

	go func() {
		defer w.Close()

		totalSamples := int(sampleRate * int(duration.Seconds()))
		clockHalfPeriod := int(float64(sampleRate) / clockFreq / 2)
		dataHalfPeriod := sampleRate / dataFreq / 2

		clockVal := int16(amplitude)
		dataVal := int16(amplitude)

		buf := make([]byte, 2)

		for i := 0; i < totalSamples; i++ {
			if i%clockHalfPeriod == 0 {
				clockVal = -clockVal
			}
			if i%dataHalfPeriod == 0 {
				dataVal = -dataVal
			}

			sample := clockVal/2 + dataVal/2
			buf[0] = byte(sample)
			buf[1] = byte(sample >> 8)

			if _, err := w.Write(buf); err != nil {
				return // Pipe closed or error
			}
		}
	}()

	player := ctx.NewPlayer(r)
	player.Play()

	time.Sleep(duration + 100*time.Millisecond)
}
