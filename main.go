package main

import (
	"log"
	"math"
	"time"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

const (
	tonicPitch            = 8
	harmonicAmplitudeBase = .88
	numHarmonics          = 12
)

type harmonic struct {
	ratio, pitch, amplitude float64
}

var harmonics []harmonic

func init() {
	for i := 1.0; i <= numHarmonics; i++ {
		harmonics = append(harmonics, harmonic{
			ratio:     i,
			pitch:     math.Log2(i),
			amplitude: math.Pow(harmonicAmplitudeBase, i) * (1 - harmonicAmplitudeBase),
		})
	}
}

func totalDissonance(pitch float64, playingPitches []float64) float64 {
	d := 0.0
	for _, playing := range playingPitches {
		for _, h1 := range harmonics {
			for _, h2 := range harmonics {
				d += beatAmplitude(h1.amplitude, h2.amplitude) * dissonance(h1.amplitude, h2.amplitude, playing+h1.pitch, pitch+h2.pitch)
			}
		}
	}
	return d
}

func main() {
	app.Main(func(a app.App) {
		var (
			glctx gl.Context
			sz    size.Event
		)

		repaint := make(chan struct{}, 1)
		go func() {
			for range repaint {
				a.Send(paint.Event{})
				time.Sleep(time.Second / 60)
			}
		}()

		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx, sz)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				sz = e
				if program != nil {
					program.Size(sz.Size())
				}
			case paint.Event:
				if glctx == nil {
					continue
				}

				onPaint(glctx, sz)
				a.Publish()
			case touch.Event:
				clipX := 2*e.X/float32(sz.WidthPx) - 1
				clipY := 1 - 2*e.Y/float32(sz.HeightPx)
				e.X, e.Y = program.Clip2World(clipX, clipY)

				keys.Touch(e)

				select {
				case repaint <- struct{}{}:
				default:
				}
			}
		}
	})
}

var (
	program *Program
	keys    *Keys
)

func onStart(glctx gl.Context, sz size.Event) {
	var err error
	program, err = NewProgram(glctx, sz.Size())
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	keys = NewKeys(glctx, program)

	startAudio()
}

func onStop(glctx gl.Context) {
	stopAudio()

	keys.Release()
	program.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(0, 0, 0, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)

	keys.Draw()
}
