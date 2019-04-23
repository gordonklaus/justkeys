package main

import (
	"math"
	"sort"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

const tonicPitch = 7

type Keys struct {
	glctx   gl.Context
	program *Program

	keys    []*Key
	pressed map[ratio]*Key
}

func NewKeys(glctx gl.Context, program *Program) *Keys {
	k := &Keys{
		glctx:   glctx,
		program: program,
		pressed: map[ratio]*Key{},
		keys: []*Key{{
			pitch: ratio{1 << tonicPitch, 1},
			seqs:  map[touch.Sequence]struct{}{},
		}},
	}
	return k
}

func (k *Keys) Release() {
	// k.buffer.Release()
}

func (k *Keys) Draw() {
	vs := []Vertex{}
	for _, key := range k.keys {
		p := float32(math.Log2(key.pitch.float()))
		color := mgl32.Vec4{.3, .3, .3, 1}
		if key.tone != nil {
			color = mgl32.Vec4{.6, .6, .6, 1}
		}
		vs = append(vs,
			Vertex{Position: mgl32.Vec2{p, .5}, Color: color},
			Vertex{Position: mgl32.Vec2{p + .02, .5}},
			Vertex{Position: mgl32.Vec2{p, 1}},

			Vertex{Position: mgl32.Vec2{p, .5}, Color: color},
			Vertex{Position: mgl32.Vec2{p, 1}},
			Vertex{Position: mgl32.Vec2{p - .02, .5}},

			Vertex{Position: mgl32.Vec2{p, .5}, Color: color},
			Vertex{Position: mgl32.Vec2{p - .02, .5}},
			Vertex{Position: mgl32.Vec2{p, 0}},

			Vertex{Position: mgl32.Vec2{p, .5}, Color: color},
			Vertex{Position: mgl32.Vec2{p, 0}},
			Vertex{Position: mgl32.Vec2{p + .02, .5}},
		)
	}
	buffer := NewVertexBuffer(k.glctx, gl.TRIANGLES, vs)

	k.glctx.LineWidth(1)
	k.program.Draw(buffer, mgl32.Ident4())

	buffer.Release()
}

func (k *Keys) Touch(e touch.Event) {
	if e.Y > .85 {
		tones.mu.Lock()
		pmIndex = math.Max(0, math.Min(.97, float64(e.X-tonicPitch)))
		for _, t := range tones.MultiVoice.Voices {
			t.(*Tone).mu.Lock()
			for _, o := range t.(*Tone).Osc {
				o.Index(pmIndex)
			}
			t.(*Tone).mu.Unlock()
		}
		tones.mu.Unlock()
		return
	}

	switch e.Type {
	case touch.TypeBegin:
		key := k.keyAt(float64(e.X))
		key.seqs[e.Sequence] = struct{}{}
		if key.tone == nil {
			k.pressed[key.pitch] = key
			tone := NewTone(math.Log2(key.pitch.float()))
			tones.AddTone(tone)
			key.tone = tone
			k.update()
			// } else {
			// 	key.tone.Release()
			// 	key.tone = nil
			// 	for seq := range key.seqs {
			// 		delete(key.seqs, seq)
			// 		delete(k.pressed, key.pitch)
			// 	}
			// 	k.update()
		}
	case touch.TypeEnd:
		var key *Key
		for _, k := range k.pressed {
			if _, ok := k.seqs[e.Sequence]; ok {
				key = k
				break
			}
		}
		delete(key.seqs, e.Sequence)
		if len(key.seqs) == 0 {
			delete(k.pressed, key.pitch)
			key.tone.Release()
			key.tone = nil
			k.update()
		}
	}
}

func (k *Keys) keyAt(x float64) *Key {
	freq := math.Exp2(x)
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

func (k *Keys) update() {
	if len(k.pressed) == 0 {
		k.keys = []*Key{{
			pitch: ratio{1 << tonicPitch, 1},
			seqs:  map[touch.Sequence]struct{}{},
		}}
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
				seqs:  map[touch.Sequence]struct{}{},
			})
		}
	}
}

type Key struct {
	pitch ratio
	seqs  map[touch.Sequence]struct{}
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
