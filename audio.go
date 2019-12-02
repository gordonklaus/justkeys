package main

import (
	"math"
	"math/rand"
	"sync"

	"github.com/gordonklaus/audio"
)

var (
	tones       Tones
	playControl audio.PlayControl
)

func startAudio() {
	playControl = audio.PlayAsync(&tones)
}

func stopAudio() {
	playControl.Stop()
}

type Tones struct {
	mu         sync.Mutex
	MultiVoice audio.MultiVoice

	Reverb [2]audio.Reverb
}

func (t *Tones) AddTone(v *Tone) {
	t.mu.Lock()
	t.MultiVoice.Add(v)
	t.mu.Unlock()
}

func (t *Tones) Sing() (float64, float64) {
	t.mu.Lock()
	x := t.MultiVoice.Sing() / 8
	t.mu.Unlock()

	l := audio.Crossfade(x, .2, t.Reverb[0].Filter(x))
	r := audio.Crossfade(x, .2, t.Reverb[1].Filter(x))
	return l, r
}

func (t *Tones) Done() bool {
	return false
}

type Tone struct {
	mu  sync.Mutex
	Amp audio.Control
	Osc []*audio.SineSelfPM

	autoTone     *AutoTone
	pitchOffsets []float64
}

var pmIndex = .8

func NewTone(pitch float64, autoTone *AutoTone) *Tone {
	t := &Tone{
		autoTone: autoTone,
	}
	t.Amp.SetPoints([]*audio.ControlPoint{{0, -12}, {.03, 0}, {.18, -.5}, {99999, -1}})
	for i := 0; i < 12; i++ {
		offset := rand.NormFloat64() / 256
		t.pitchOffsets = append(t.pitchOffsets, offset)
		t.Osc = append(t.Osc, new(audio.SineSelfPM).Index(pmIndex).Freq(math.Exp2(pitch+offset)))
	}
	return t
}

func (t *Tone) SetPitch(p float64) {
	t.mu.Lock()
	for i, o := range t.Osc {
		o.Freq(math.Exp2(p + t.pitchOffsets[i]))
	}
	t.mu.Unlock()
}

func (t *Tone) Release() {
	t.mu.Lock()
	a := t.Amp.Sing()
	t.Amp.SetPoints([]*audio.ControlPoint{{0, a}, {4, -12}})
	t.mu.Unlock()
}

func (t *Tone) Sing() float64 {
	t.mu.Lock()
	a := t.Amp.Sing()
	t.mu.Unlock()

	x := 0.0
	for _, o := range t.Osc {
		x += o.Sing()
	}
	return math.Exp2(a) * x / float64(len(t.Osc))
}

func (t *Tone) Done() bool {
	t.mu.Lock()
	done := t.Amp.Done()
	t.mu.Unlock()

	return done
}
