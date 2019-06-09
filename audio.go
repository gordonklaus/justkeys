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
	Reverb     [2]audio.Reverb
}

func (t *Tones) AddTone(v *Tone) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.MultiVoice.Add(v)
}

func (t *Tones) Sing() (float64, float64) {
	x := t.MultiVoice.Sing() / 8
	l := audio.Crossfade(x, .2, t.Reverb[0].Filter(x))
	r := audio.Crossfade(x, .2, t.Reverb[1].Filter(x))
	return l, r
}

func (t *Tones) Done() bool {
	return false
}

type Tone struct {
	mu  sync.Mutex
	Osc []*audio.SineSelfPM
	Amp audio.Control
}

var pmIndex = .8

func NewTone(pitch float64) *Tone {
	t := &Tone{}
	for i := 0; i < 12; i++ {
		t.Osc = append(t.Osc, new(audio.SineSelfPM).Index(pmIndex).Freq(math.Exp2(pitch+rand.NormFloat64()/256)))
	}
	t.Amp.SetPoints([]*audio.ControlPoint{{0, -12}, {.03, 0}, {.18, -.5}, {99999, -1}})
	return t
}

func (t *Tone) Release() {
	t.mu.Lock()
	a := t.Amp.Sing()
	t.Amp.SetPoints([]*audio.ControlPoint{{0, a}, {4, -12}})
	t.mu.Unlock()
}

func (t *Tone) Sing() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	x := 0.0
	for _, o := range t.Osc {
		x += o.Sing()
	}
	return math.Exp2(t.Amp.Sing()) * x / float64(len(t.Osc))
}

func (t *Tone) Done() bool {
	return t.Amp.Done()
}
