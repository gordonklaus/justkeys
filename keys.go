package main

import (
	"math"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gordonklaus/ui"
)

const tonicPitch = 7

type Keyboard struct {
	ui.View

	buf *ui.TriangleBuffer

	keys    []*Key
	pressed map[ratio]*Key
}

func NewKeyboard() *Keyboard {
	k := &Keyboard{
		pressed: map[ratio]*Key{},
		keys: []*Key{{
			pitch: ratio{1 << tonicPitch, 1},
			seqs:  map[ui.PointerID]struct{}{},
		}},
	}
	k.View = ui.NewView(k)
	return k
}

func (k *Keyboard) Release() {
	// k.buffer.Release()
}

func (k *Keyboard) Draw(gfx *ui.Graphics) {
	black := ui.Color{}

	ts := []ui.Triangle{}
	for _, key := range k.keys {
		p := math.Log2(key.pitch.float())
		color := ui.Color{.3, .3, .3, 1}
		if key.tone != nil {
			color = ui.Color{.6, .6, .6, 1}
		}
		ts = append(ts, []ui.Triangle{{
			k.vertex(p, .5, color),
			k.vertex(p+.02, .5, black),
			k.vertex(p, 1, black),
		}, {
			k.vertex(p, .5, color),
			k.vertex(p, 1, black),
			k.vertex(p-.02, .5, black),
		}, {
			k.vertex(p, .5, color),
			k.vertex(p-.02, .5, black),
			k.vertex(p, 0, black),
		}, {
			k.vertex(p, .5, color),
			k.vertex(p, 0, black),
			k.vertex(p+.02, .5, black),
		}}...)
	}
	k.buf = ui.NewTriangleBuffer(ts)

	gfx.Draw(k.buf, mgl32.Ident4())

	k.buf.Release()
}

const octaveWidth = 165

func (k *Keyboard) freqToX(freq float64) float64 {
	octaves := k.Width() / octaveWidth
	minFreq := tonicPitch - octaves/2
	return k.Width() * (freq - minFreq) / octaves
}

func (k *Keyboard) xToFreq(x float64) float64 {
	octaves := k.Width() / octaveWidth
	minFreq := tonicPitch - octaves/2
	return minFreq + octaves*x/k.Width()
}

func (k *Keyboard) vertex(freq, y float64, color ui.Color) ui.Vertex {
	x := k.freqToX(freq)
	y *= k.Height()
	return ui.Vertex{Position: ui.Position{x, y}, Color: color}
}

func (k *Keyboard) PointerDown(p ui.Pointer) {
	if k.pmIndex(p) {
		return
	}

	key := k.keyAt(float64(p.X))
	key.seqs[p.ID] = struct{}{}
	if key.tone == nil {
		k.pressed[key.pitch] = key
		tone := NewTone(math.Log2(key.pitch.float()))
		tones.AddTone(tone)
		key.tone = tone
		k.update()
	} else if p.Type.Mouse() {
		key.tone.Release()
		key.tone = nil
		for seq := range key.seqs {
			delete(key.seqs, seq)
			delete(k.pressed, key.pitch)
		}
		k.update()
	}
}

func (k *Keyboard) PointerMove(p ui.Pointer) {
	if k.pmIndex(p) {
		return
	}
}

func (k *Keyboard) PointerUp(p ui.Pointer) {
	if k.pmIndex(p) {
		return
	}

	if p.Type.Mouse() {
		return
	}

	for _, key := range k.pressed {
		if _, ok := key.seqs[p.ID]; !ok {
			continue
		}
		delete(key.seqs, p.ID)
		if len(key.seqs) == 0 {
			delete(k.pressed, key.pitch)
			key.tone.Release()
			key.tone = nil
			k.update()
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

func (k *Keyboard) keyAt(x float64) *Key {
	freq := math.Exp2(k.xToFreq(x))
	i := sort.Search(len(k.keys), func(i int) bool { return k.keys[i].pitch.float() >= freq })
	if i == len(k.keys) {
		return k.keys[len(k.keys)-1]
	}
	if i == 0 {
		return k.keys[0]
	}
	if freq/k.keys[i-1].pitch.float() < k.keys[i].pitch.float()/freq {
		return k.keys[i-1]
	}
	return k.keys[i]
}

func (k *Keyboard) update() {
	if len(k.pressed) == 0 {
		k.keys = []*Key{{
			pitch: ratio{1 << tonicPitch, 1},
			seqs:  map[ui.PointerID]struct{}{},
		}}
		k.Redraw()
		return
	}

	pitches := []ratio{}
	pitch := ratio{}
	for _, key := range k.pressed {
		pitch = key.pitch
		break
	}

	pow := func(a, x int) int {
		y := 1
		for x > 0 {
			y *= a
			x--
		}
		return y
	}
	mul := func(n, d, a, x int) (int, int) {
		if x > 0 {
			return n * pow(a, x), d
		}
		return n, d * pow(a, -x)
	}
threes:
	for _, three := range []int{-2, -1, 0, 1, 2} {
	fives:
		for _, five := range []int{-1, 0, 1} {
			n, d := 1, 1
			n, d = mul(n, d, 3, three)
			n, d = mul(n, d, 5, five)
			p2 := pitch.mul(ratio{n, d})
			for p := range k.pressed {
				three, five := factorize(p2.div(p))
				if three > 2 {
					continue threes
				}
				if five > 1 {
					continue fives
				}
			}
			pitches = append(pitches, p2)
		}
	}

	for _, p := range pitches {
		for p := p.mul(ratio{2, 1}); p.less(ratio{1 << 10, 1}); p = p.mul(ratio{2, 1}) {
			pitches = append(pitches, p)
		}
		for p := p.mul(ratio{1, 2}); !p.less(ratio{1, 1 << 7}); p = p.mul(ratio{1, 2}) {
			pitches = append(pitches, p)
		}
	}

	k.keys = nil
	sort.Slice(pitches, func(i, j int) bool { return pitches[i].less(pitches[j]) })
	for _, p := range pitches {
		if key, ok := k.pressed[p]; ok {
			k.keys = append(k.keys, key)
		} else {
			k.keys = append(k.keys, &Key{
				pitch: p,
				seqs:  map[ui.PointerID]struct{}{},
			})
		}
	}

	k.Redraw()
}

type Key struct {
	pitch ratio
	seqs  map[ui.PointerID]struct{}
	tone  *Tone
}

func factorize(r ratio) (threes, fives int) {
	n := r.a * r.b
	for n%3 == 0 {
		n /= 3
		threes++
	}
	for n%5 == 0 {
		n /= 5
		fives++
	}
	return
}

func gcd(a, b int) int {
	if a > b {
		a, b = b, a
	}
	for a > 0 {
		a, b = b%a, a
	}
	return b
}

type ratio struct {
	a, b int
}

func (r ratio) mul(s ratio) ratio {
	r.a *= s.a
	r.b *= s.b
	d := gcd(r.a, r.b)
	r.a /= d
	r.b /= d
	return r
}

func (r ratio) div(s ratio) ratio { return r.mul(ratio{s.b, s.a}) }

func (r ratio) less(s ratio) bool { return r.a*s.b < s.a*r.b }
func (r ratio) float() float64    { return float64(r.a) / float64(r.b) }
