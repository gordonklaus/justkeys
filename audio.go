package main

import (
	"math"
	"sync"

	"github.com/gordonklaus/audio"
)

var (
	tones       Tones
	playControl audio.PlayControl
)

func startAudio() {
	tones.Reverb = audio.NewReverb()
	playControl = audio.PlayAsync(&tones)
}

func stopAudio() {
	playControl.Stop()
}

type Tones struct {
	mu         sync.Mutex
	MultiVoice audio.MultiVoice
	Reverb     *audio.Reverb
}

func (t *Tones) AddTone(v *Tone) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.MultiVoice.Add(v)
}

func (t *Tones) Sing() float64 {
	x := t.MultiVoice.Sing() / 8
	return (3*x + t.Reverb.Filter(x)) / 4
}

func (t *Tones) Done() bool {
	return false
}

type Tone struct {
	mu        sync.Mutex
	Harmonics []*ToneHarmonic
	Amp       audio.Control
}

func NewTone(pitch float64) *Tone {
	ToneHarmonics := make([]*ToneHarmonic, len(harmonics))
	for i, h := range harmonics {
		ToneHarmonics[i] = NewToneHarmonic(h)
	}
	t := &Tone{Harmonics: ToneHarmonics}
	t.SetPitch(pitch)
	t.Amp.SetPoints([]*audio.ControlPoint{{0, -12}, {.05, 0}, {99999, 0}})
	return t
}

func (t *Tone) Release() {
	t.mu.Lock()
	a := t.Amp.Sing()
	t.Amp.SetPoints([]*audio.ControlPoint{{0, a}, {4, -12}})
	t.mu.Unlock()
}

func (t *Tone) SetPitch(p float64) {
	freq := math.Exp2(p)
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, h := range t.Harmonics {
		h.SetFreq(freq)
	}
}

func (t *Tone) Sing() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	x := 0.0
	for i := range t.Harmonics {
		x += t.Harmonics[i].Sing()
	}
	return math.Exp2(t.Amp.Sing()) * x
}

func (t *Tone) Done() bool {
	return t.Amp.Done()
}

type ToneHarmonic struct {
	harmonic harmonic
	Sine     audio.SineOsc
}

func NewToneHarmonic(h harmonic) *ToneHarmonic {
	return &ToneHarmonic{harmonic: h}
}

func (h *ToneHarmonic) SetFreq(freq float64) {
	h.Sine.Freq(freq * h.harmonic.ratio)
}

func (h *ToneHarmonic) Sing() float64 {
	return h.harmonic.amplitude * h.Sine.Sing()
}
