package main

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
)

func main() {
	const (
		sampleRate      = 44100
		channelCount    = 1
		bitDepthInBytes = 2 // 16-bit PCM
		duration        = 3 * time.Second

		clockFreq = 300.0
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
		buf := make([]byte, 2)

		var period int
		period = sampleRate / clockFreq
		halfPeriod := period / 2
		quarterPeriod := period / 4

		isFirstQuarter := true
		for i := range totalSamples {
			i = i + 1
			t := i % period
			ht := t % halfPeriod

			isFirstQuarter = !((t % halfPeriod) > quarterPeriod)
			var qt float64

			if isFirstQuarter {
				qt = float64(quarterPeriod - ht)
			} else {
				qt = float64(ht - quarterPeriod)
			}

			isFirstHalf := t < (period / 2)
			tScaled := amplitude / (float64(quarterPeriod) / float64(qt))

			var data int
			c := math.Pow(float64(amplitude), 2)
			a := math.Pow(float64(tScaled), 2)
			if isFirstHalf {
				data = int(math.Sqrt(c - a))
			} else {
				data = int(-math.Sqrt(c - a))
			}

			fmt.Printf("isFirstQuarter:%v, tmodquarrt:%v,  tmodhalf:%v, p:%v, quarterPeriod:%v, qt:%v,  t:%v data:%v c:%v, tScaled:%v\n", isFirstQuarter, (t % quarterPeriod), (t % halfPeriod), period, quarterPeriod, qt, t, data, c, tScaled)

			buf[0] = byte(data)
			buf[1] = byte(data >> 8)

			if _, err := w.Write(buf); err != nil {
				return // Pipe closed or error
			}
		}
	}()

	player := ctx.NewPlayer(r)
	player.Play()

	time.Sleep(duration + 100*time.Millisecond)
}
