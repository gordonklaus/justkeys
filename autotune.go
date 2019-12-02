package main

import (
	"math"
	"sync"
	"time"
)

type AutoTuner struct {
	mu    sync.Mutex
	tones []*AutoTone
}

type AutoTone struct {
	mu                 sync.Mutex
	targetPitch, pitch float64

	dp float64
}

func (t *AutoTone) SetTargetPitch(p float64) {
	t.mu.Lock()
	t.targetPitch = p
	t.mu.Unlock()
}

func (t *AutoTone) GetTargetPitch() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.targetPitch
}

func (t *AutoTone) GetPitch() float64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.pitch
}

func (t *AutoTone) SetPitch(p float64) {
	t.mu.Lock()
	t.pitch = p
	t.mu.Unlock()
}

func NewAutoTuner() *AutoTuner {
	at := &AutoTuner{}
	go at.tune()
	return at
}

func (at *AutoTuner) NewTone(pitch float64) *AutoTone {
	t := &AutoTone{
		targetPitch: pitch,
		pitch:       pitch,
	}

	at.mu.Lock()
	at.tones = append(at.tones, t)
	at.mu.Unlock()

	return t
}

func (at *AutoTuner) RemoveTone(t *AutoTone) {
	at.mu.Lock()
	for i, t2 := range at.tones {
		if t2 == t {
			at.tones = append(at.tones[:i], at.tones[i+1:]...)
			break
		}
	}
	at.mu.Unlock()
}

func (at *AutoTuner) tune() {
	for range time.Tick(time.Millisecond) {
		at.mu.Lock()
		ddp := make([]float64, len(at.tones))
		for i2, k2 := range at.tones {
			for i1, k1 := range at.tones[:i2] {
				dd := totalDissonanceDerivative(k2.pitch - k1.pitch)
				ddp[i1] += dd
				ddp[i2] -= dd
			}
		}
		dp0 := 0.
		for i, t := range at.tones {
			t.dp += ddp[i] / 1000
			t.dp *= .9
			if i == 0 {
				dp0 = t.dp
			}
			t.dp -= dp0

			// target := t.GetTargetPitch()
			// t.dp += (target - t.pitch) / 1
			// t.SetTargetPitch(target + (t.pitch-target)/2)
		}
		for _, t := range at.tones {
			t.SetPitch(t.pitch + t.dp/1000)
		}
		at.mu.Unlock()
	}
}

const (
	harmonicAmplitudeBase = .88
	numHarmonics          = 15
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

func totalDissonanceDerivative(dp float64) float64 {
	d := 0.0
	for _, h1 := range harmonics {
		for _, h2 := range harmonics {
			d += beatAmplitude(h1.amplitude, h2.amplitude) * dissonanceDerivative(dp+h2.pitch-h1.pitch)
		}
	}
	return d
}

func dissonanceDerivative(dp float64) float64 {
	x := math.Abs(20 * dp)
	return math.Copysign(20*(1-x)*math.Exp(1-x), dp)
}
