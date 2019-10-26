package main

import (
	"math"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/gordonklaus/ui"
)

const tonicPitch = 7
const maxRelativeRoughness = 2

type Keyboard struct {
	ui.View

	buf *ui.TriangleBuffer

	keys         []*Key
	pressed      map[ratio]*Key
	lastReleased ratio
}

func NewKeyboard() *Keyboard {
	k := &Keyboard{
		pressed:      map[ratio]*Key{},
		lastReleased: ratio{1 << tonicPitch, 1},
	}
	k.View = ui.NewView(k)
	k.update()
	return k
}

func (k *Keyboard) Release() {
	// k.buffer.Release()
}

func (k *Keyboard) Draw(gfx *ui.Graphics) {
	black := ui.Color{}

	ts := []ui.Triangle{}
	for _, key := range k.keys {
		p := math.Log2(key.freq.float())
		y := key.y / maxRelativeRoughness
		color := ui.Color{.3, .3, .3, 1}
		if key.tone != nil {
			color = ui.Color{.6, .6, .6, 1}
		}
		ts = append(ts, ui.Triangle{
			k.vertex(p-.01, y, color),
			k.vertex(p+.01, y, color),
			k.vertex(p, 1, black),
		})
	}
	k.buf = ui.NewTriangleBuffer(ts)

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
	key.pointers[p.ID] = struct{}{}
	if key.tone == nil {
		k.pressed[key.freq] = key
		tone := NewTone(math.Log2(key.freq.float()))
		tones.AddTone(tone)
		key.tone = tone
		k.update()
	} else if p.Type.Mouse() {
		key.tone.Release()
		key.tone = nil
		k.lastReleased = key.freq
		for ptr := range key.pointers {
			delete(key.pointers, ptr)
			delete(k.pressed, key.freq)
		}
		k.update()
	}
}

func (k *Keyboard) PointerMove(p ui.Pointer) {
	if p.Type.Mouse() {
		return
	}

	if k.pmIndex(p) {
		return
	}
}

func (k *Keyboard) PointerUp(p ui.Pointer) {
	if p.Type.Mouse() {
		return
	}

	if k.pmIndex(p) {
		return
	}

	for _, key := range k.pressed {
		if _, ok := key.pointers[p.ID]; !ok {
			continue
		}
		delete(key.pointers, p.ID)
		if len(key.pointers) == 0 {
			delete(k.pressed, key.freq)
			key.tone.Release()
			key.tone = nil
			k.lastReleased = key.freq
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

func (k *Keyboard) keyAt(p ui.Position) *Key {
	freq := math.Exp2(k.xToPitch(p.X))
	i := sort.Search(len(k.keys), func(i int) bool { return k.keys[i].freq.float() >= freq })
	if i == len(k.keys) {
		i = len(k.keys) - 1
	}
	if i > 0 && freq/k.keys[i-1].freq.float() < k.keys[i].freq.float()/freq {
		i--
	}

	iLeft := i
	for iLeft > 0 && maxRelativeRoughness*p.Y/k.Height() < k.keys[iLeft].y {
		iLeft--
	}
	iRight := i
	for iRight < len(k.keys) && maxRelativeRoughness*p.Y/k.Height() < k.keys[iRight].y {
		iRight++
	}
	if iLeft < 0 {
		if iRight >= len(k.keys) {
			return nil
		}
		return k.keys[iRight]
	}
	if iRight >= len(k.keys) {
		return k.keys[iLeft]
	}
	if freq/k.keys[iLeft].freq.float() < k.keys[iRight].freq.float()/freq {
		return k.keys[iLeft]
	}
	return k.keys[iRight]
}

func (k *Keyboard) update() {
	anchors := []ratio{}
	isAnchor := func(r ratio) bool {
		for _, s := range anchors {
			if s == r {
				return true
			}
		}
		return false
	}
	for f := range k.pressed {
		anchors = append(anchors, f)
	}
	if len(anchors) == 0 {
		anchors = []ratio{k.lastReleased}
	}
	// fmt.Println(anchors)

	type keyFreq struct {
		freq ratio
		y    float64
	}

	freqMap := map[ratio]keyFreq{}
	A := make([]float64, 10)
	for i := range A {
		A[i] = 1 / float64(i+1)
	}
	anchorRoughness := roughness(A, ratioN(anchors)...)
	add := func(f ratio) {
		if x := k.pitchToX(math.Log2(f.float())); x < 0 || x > k.Width() {
			return
		}
		if _, ok := freqMap[f]; ok {
			return
		}
		y := anchorRoughness
		if !isAnchor(f) {
			y = roughness(A, ratioN(append(anchors, f))...)
		}
		// fmt.Println(a, b, f, N, y)
		if (y-anchorRoughness)/maxRelativeRoughness < 1 {
			freqMap[f] = keyFreq{
				freq: f,
				y:    y,
			}
		}
	}

	for _, anchor := range anchors {
		for b := 1; b <= 8; b++ {
			for a := b; a <= 4*b; a++ {
				if gcd(a, b) != 1 {
					continue
				}
				add(anchor.mul(ratio{a, b}))
				if !(a == 1 && b == 1) {
					add(anchor.mul(ratio{b, a}))
				}
			}
		}
	}

	freqs := make([]keyFreq, 0, len(freqMap))
	for _, f := range freqMap {
		freqs = append(freqs, f)
	}
	sort.Slice(freqs, func(i, j int) bool { return freqs[i].freq.less(freqs[j].freq) })

	yMin := freqs[0].y
	for i := range freqs {
		if freqs[i].y < yMin {
			yMin = freqs[i].y
		}
	}
	for i := range freqs {
		freqs[i].y -= yMin
	}

	k.keys = nil
	for _, f := range freqs {
		if key, ok := k.pressed[f.freq]; ok {
			key.y = f.y
			k.keys = append(k.keys, key)
		} else {
			k.keys = append(k.keys, &Key{
				freq:     f.freq,
				y:        f.y,
				pointers: map[ui.PointerID]struct{}{},
			})
		}
	}

	k.Redraw()
}

func ratioN(R []ratio) []int {
	m := R[0].b
	for _, r := range R[1:] {
		m = lcm(m, r.b)
	}

	N := make([]int, len(R))
	for i, r := range R {
		N[i] = r.mul(ratio{m, 1}).a
	}
	return N
}

type Key struct {
	freq     ratio
	y        float64
	pointers map[ui.PointerID]struct{}
	tone     *Tone
}

func factorize(r ratio) (threes, fives, sevens int) {
	n := r.a * r.b
	for n%3 == 0 {
		n /= 3
		threes++
	}
	for n%5 == 0 {
		n /= 5
		fives++
	}
	for n%7 == 0 {
		n /= 7
		sevens++
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

func lcm(a, b int) int {
	_, _, gcd, a, b := gcd2(a, b)
	return a * b * gcd
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
