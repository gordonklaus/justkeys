package main

import (
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

func main() {
	app.Main(func(a app.App) {
		var (
			glctx gl.Context
			sz    size.Event
		)

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
				if e.Type != touch.TypeMove {
					a.Send(paint.Event{})
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

	glctx.Enable(gl.BLEND)
	glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	keys.Draw()
}
