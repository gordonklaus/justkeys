package main

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gordonklaus/ui"
	"gonum.org/v1/gonum/mathext"
)

const tonicPitch = 8

type Keyboard struct {
	ui.View

	buf *ui.TriangleBuffer

	keys      []*Key
	autoTuner *AutoTuner
}

type Key struct {
	pitch    float64
	pointers map[ui.PointerID]struct{}
	tone     *Tone
	autoTone *AutoTone
}

func NewKeyboard() *Keyboard {
	k := &Keyboard{}
	k.View = ui.NewView(k)
	k.autoTuner = NewAutoTuner()
	k.autoTuner.NewTone(tonicPitch)
	go k.updateKeys()
	go k.updateTones()
	return k
}

func (k *Keyboard) Release() {
	// k.buffer.Release()
}

func (k *Keyboard) updateTones() {
	for range time.Tick(time.Millisecond) {
		tones.mu.Lock()
		for _, t := range tones.MultiVoice.Voices {
			t := t.(*Tone)
			t.SetPitch(t.autoTone.GetPitch())
		}
		tones.mu.Unlock()
	}
}

func (k *Keyboard) updateKeys() {
	for range time.Tick(time.Second / 60) {
		k.Redraw()
	}
}

func (k *Keyboard) Draw(gfx *ui.Graphics) {
	if len(k.keys) == 0 {
		return
	}

	black := ui.Color{}

	ts := []ui.Triangle{}
	for _, key := range k.keys {
		p := key.autoTone.GetPitch()
		color := ui.Color{.6, .6, .6, 1}
		ts = append(ts, ui.Triangle{
			k.vertex(p-.01, 0, color),
			k.vertex(p+.01, 0, color),
			k.vertex(p, 1, black),
		})
	}
	k.buf = ui.NewTriangleBuffer(gfx, ts)

	gfx.Draw(k.buf, mgl32.Ident4())

	k.buf.Release()
}

const octaveWidth = 165

func (k *Keyboard) pitchToX(pitch float64) float64 {
	octaves := k.Width() / octaveWidth
	minPitch := tonicPitch - octaves/2
	return k.Width() * (pitch - minPitch) / octaves
}

func (k *Keyboard) xToPitch(x float64) float64 {
	octaves := k.Width() / octaveWidth
	minPitch := tonicPitch - octaves/2
	return minPitch + octaves*x/k.Width()
}

func (k *Keyboard) vertex(pitch, y float64, color ui.Color) ui.Vertex {
	x := k.pitchToX(pitch)
	y *= k.Height()
	return ui.Vertex{Position: ui.Position{x, y}, Color: color}
}

func (k *Keyboard) PointerDown(p ui.Pointer) {
	if p.Type.Mouse() {
		return
	}

	if k.pmIndex(p) {
		return
	}

	key := k.keyAt(p.Position)

	if key != nil {
		key.pointers[p.ID] = struct{}{}
		if p.Type.Mouse() {
			key.tone.Release()
			for i := range k.keys {
				if k.keys[i] == key {
					k.keys = append(k.keys[:i], k.keys[i+1:]...)
					break
				}
			}
			k.update()
		}
		return
	}

	pitch := k.xToPitch(p.X)
	autoTone := k.autoTuner.NewTone(pitch)

	tone := NewTone(pitch, autoTone)
	tones.AddTone(tone)

	key = &Key{
		pitch:    pitch,
		pointers: map[ui.PointerID]struct{}{p.ID: struct{}{}},
		tone:     tone,
		autoTone: autoTone,
	}
	k.keys = append(k.keys, key)
	k.update()
}

func (k *Keyboard) PointerMove(p ui.Pointer) {
	if p.Type.Mouse() {
		return
	}

	if k.pmIndex(p) {
		return
	}

	// for _, key := range k.keys {
	// 	if _, ok := key.pointers[p.ID]; ok {
	// 		key.autoTone.SetTargetPitch(k.xToPitch(p.X))
	// 		k.update()
	// 		break
	// 	}
	// }
}

func (k *Keyboard) PointerUp(p ui.Pointer) {
	if p.Type.Mouse() {
		return
	}

	if k.pmIndex(p) {
		return
	}

	for i, key := range k.keys {
		if _, ok := key.pointers[p.ID]; ok {
			delete(key.pointers, p.ID)
			if len(key.pointers) == 0 {
				k.keys = append(k.keys[:i], k.keys[i+1:]...)
				key.tone.Release()
				k.autoTuner.RemoveTone(key.autoTone)
				k.update()
			}
			break
		}
	}
}

func (k *Keyboard) pmIndex(p ui.Pointer) bool {
	if p.Y/k.Height() > .85 {
		tones.mu.Lock()
		pmIndex = math.Max(0, math.Min(.97, p.X/k.Width()))
		for _, t := range tones.MultiVoice.Voices {
			t.(*Tone).mu.Lock()
			for _, o := range t.(*Tone).Osc {
				o.Index(pmIndex)
			}
			t.(*Tone).mu.Unlock()
		}
		tones.mu.Unlock()
		return true
	}
	return false
}

func (k *Keyboard) keyAt(p ui.Position) *Key {
	var nearest *Key
	mindp := math.MaxFloat64
	for _, key := range k.keys {
		dp := math.Abs(key.pitch - k.xToPitch(p.X))
		if dp < mindp {
			mindp = dp
			nearest = key
		}
	}
	if mindp < .02 {
		return nearest
	}

	return nil
}

func (k *Keyboard) update() {
	k.Redraw()
}

func beatAmplitude(a1, a2 float64) float64 {
	// return math.Min(a1, a2)
	// return a1 * a2 * math.Hypot(a1, a2)
	// return math.Hypot(a1, a2)

	// Avoid floating point rounding errors.
	if math.Abs(math.Log10(a1/a2)) > 10 {
		return math.Min(a1, a2)
	}

	meanSquare := a1*a1 + a2*a2

	m := 4 * a1 * a2 / ((a1 + a2) * (a1 + a2))
	squareMean := (a1 + a2) * mathext.CompleteE(m) / (math.Pi / 2)
	squareMean *= squareMean

	stddev := math.Sqrt(meanSquare - squareMean)
	return stddev
}
