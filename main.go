package main

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"math/cmplx"
	"sort"

	"github.com/ebitengine/oto/v3"

	"gonum.org/v1/gonum/dsp/fourier"
)

func main() {
	const (
		sampleRate      = 44100
		channelCount    = 1
		bitDepthInBytes = 2 // 16-bit PCM
		duration        = 10 * time.Second

		clockFreq = 697.0
		dataFreq  = 1209
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

	tone := []float64{}
	totalSamples := int(sampleRate * int(duration.Seconds()))
	buf := make([]byte, totalSamples*2)

	for i := range totalSamples {
		t := float64(i) / float64(sampleRate) // time in seconds
		data := amplitude * math.Sin(2*math.Pi*clockFreq*t)
		data2 := amplitude * math.Sin(2*math.Pi*dataFreq*t)

		tone = append(tone, data/2+data2/2)

		d := int16(data/2 + data2/2)
		buf[2*i] = byte(d)
		buf[2*i+1] = byte(d >> 8)
	}

	player := ctx.NewPlayer(bytes.NewReader(buf))
	player.Play()
	time.Sleep(4 * time.Second)

	samples := sampleRate
	fft := fourier.NewFFT(samples)
	tone = tone[:samples]
	coeff := fft.Coefficients(nil, tone)

	type freqMag struct {
		freq float64
		mag  float64
	}

	var mags []freqMag
	var total float64
	for i, c := range coeff[:samples/2] { // only look at positive freqs
		m := cmplx.Abs(c)
		total += m
		mags = append(mags, freqMag{
			freq: float64(i) * float64(sampleRate) / float64(samples),
			mag:  m,
		})
	}
	mean := total / float64(samples/2)

	// Filter to significant frequencies
	const thresholdFactor = 10.0
	var result []freqMag
	for _, fm := range mags {
		if fm.mag > mean*thresholdFactor {
			result = append(result, fm)
		}
	}

	// Sort by magnitude, descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].mag > result[j].mag
	})

	fmt.Println("Significant frequencies:")
	for _, r := range result {
		fmt.Printf("freq = %.1f Hz, mag = %.1f\n", r.freq, r.mag)
	}
	// 697 1209
}
