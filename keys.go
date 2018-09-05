package main

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
	"gonum.org/v1/gonum/mathext"

	"github.com/gordonklaus/justkeys/voronoi"
)

type Keys struct {
	glctx   gl.Context
	program *Program
	// buffer  *VertexBuffer

	pressed      map[ratio]*Key
	recent       map[ratio]float64
	lastTime     time.Time
	pitchToRatio map[float64]ratio
	diagram      voronoi.Diagram
}

func NewKeys(glctx gl.Context, program *Program) *Keys {
	k := &Keys{
		glctx:   glctx,
		program: program,
		// buffer,
		pressed:  map[ratio]*Key{},
		recent:   map[ratio]float64{{1 << tonicPitch, 1}: 1},
		lastTime: time.Now(),
	}
	k.buildDiagram()
	return k
}

func (k *Keys) buildDiagram() {
	playing := []pitchAmplitude{}
	for p, a := range k.recent {
		playing = append(playing, pitchAmplitude{p, (1 - math.Cos(a*math.Pi)) / 2})
	}

	ratios := map[ratio]int{}
	for _, p := range playing {
		for _, r := range rats {
			ratios[p.pitch.mul(r)]++
		}
	}

	sites := []voronoi.Point{}
	k.pitchToRatio = map[float64]ratio{}
	for r, count := range ratios {
		if count < len(playing) {
			continue
		}

		pitch := math.Log2(r.float())
		if pitch < 4 || pitch > 14 {
			continue
		}
		diss := totalDissonance(pitch, playing)
		sites = append(sites, voronoi.Point{pitch, diss})
		k.pitchToRatio[pitch] = r
	}
	k.diagram = voronoi.ComputeDiagram(sites)
	closeCells(k.diagram)
}

func closeCells(diagram voronoi.Diagram) {
	for _, cell := range diagram.Cells {
		var out *voronoi.HalfEdge
		for edge := cell.Edges; ; {
			if edge.Type == voronoi.OutgoingRay {
				out = edge
				break
			}
			edge = edge.Next
			if edge == cell.Edges {
				break
			}
		}

		if out != nil {
			e := &voronoi.HalfEdge{
				Cell: cell,
				Type: voronoi.LineSegment,
				P1:   out.P2,
				P2:   out.Next.P1,
				Prev: out,
				Next: out.Next,
			}
			out.Pair.Type = voronoi.LineSegment
			out.Pair.Prev = e
			out.Type = voronoi.LineSegment
			out.Next = e
		}
	}
}

func (k *Keys) Release() {
	// k.buffer.Release()
}

func (k *Keys) Draw() {
	vs := []Vertex{}
	for _, cell := range k.diagram.Cells {
		color := mgl32.Vec4{.3, .3, .3, 1}
		if _, ok := k.pressed[k.pitchToRatio[cell.Site.X]]; ok {
			color = mgl32.Vec4{.6, .6, .6, 1}
		}
		site := voronoiVertexToVec2(cell.Site)
		for edge := cell.Edges; ; {
			if edge.Type == voronoi.LineSegment {
				va := voronoiVertexToVec2(edge.P1)
				vb := voronoiVertexToVec2(edge.P2)
				vs = append(vs,
					Vertex{Position: site, Color: color},
					Vertex{Position: va},
					Vertex{Position: vb},
				)
			}
			edge = edge.Next
			if edge == cell.Edges {
				break
			}
		}
	}
	buffer := NewVertexBuffer(k.glctx, gl.TRIANGLES, vs)

	k.glctx.LineWidth(1)
	k.program.Draw(buffer, mgl32.Ident4())

	buffer.Release()
}

func voronoiVertexToVec2(v voronoi.Point) mgl32.Vec2 {
	return mgl32.Vec2{float32(v.X), float32(v.Y)}
}

func (k *Keys) Update() {
	const attack, release = 4.0, 8.0
	dt := time.Now().Sub(k.lastTime).Seconds()
	k.lastTime = time.Now()

	pMax := ratio{}
	aMax := 0.0
	for p, a := range k.recent {
		if a > aMax {
			aMax = a
			pMax = p
		}
	}

	for p := range k.recent {
		d := dt / attack
		if k.pressed[p] == nil && !(len(k.pressed) == 0 && p == pMax) {
			d = -dt / release
		}
		k.recent[p] = math.Min(k.recent[p]+d, 1)
		if k.recent[p] < 0 {
			delete(k.recent, p)
		}
	}
	k.buildDiagram()
}

func (k *Keys) Touch(e touch.Event) {
	switch e.Type {
	case touch.TypeBegin:
		if pitch, ok := k.pitchForTouch(float64(e.X), float64(e.Y)); ok {
			if k.pressed[pitch] == nil {
				tone := NewTone(math.Log2(pitch.float()))
				tones.AddTone(tone)
				k.pressed[pitch] = &Key{map[touch.Sequence]struct{}{}, tone}
				if _, ok := k.recent[pitch]; !ok {
					k.recent[pitch] = 0
				}
			}
			k.pressed[pitch].seqs[e.Sequence] = struct{}{}
		}
	case touch.TypeEnd:
		for pitch, key := range k.pressed {
			delete(key.seqs, e.Sequence)
			if len(key.seqs) == 0 {
				key.tone.Release()
				delete(k.pressed, pitch)
			}
		}
	}
}

type Key struct {
	seqs map[touch.Sequence]struct{}
	tone *Tone
}

func (k *Keys) pitchForTouch(x, y float64) (ratio, bool) {
	if cell := k.diagram.Find(voronoi.Point{x, y}); cell != nil {
		return k.pitchToRatio[cell.Site.X], true
	}
	return ratio{}, false
}

func dissonance(a1, a2, p1, p2 float64) float64 {
	// f1 := math.Exp2(p1)
	// f2 := math.Exp2(p2)
	// // m := (a1*f1 + a2*f2) / (a1 + a2)
	// m := 1. //math.Pow(math.Pow(f1, a1)*math.Pow(f2, a2), 1/(a1+a2))
	// df := math.Abs(f2 - f1)
	// return m * df * math.Exp(-df/64)

	dp := math.Abs(p2 - p1)
	x := 20 * dp
	// x := dp * dp * 400
	return x * math.Exp(-x) * math.E
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

	m := math.Max(0, math.Min(1, 4*a1*a2/((a1+a2)*(a1+a2)))) // TODO: better math, not max/min
	squareMean := (a1 + a2) * mathext.CompleteE(m) / (math.Pi / 2)
	squareMean *= squareMean

	stddev := math.Sqrt(math.Max(0, meanSquare-squareMean))
	return stddev
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

var rats []ratio

func init() {
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
	for _, two := range []int{-3, -2, -1, 0, 1, 2, 3} {
		for _, three := range []int{-2, -1, 0, 1, 2} {
			for _, five := range []int{-1, 0, 1} {
				for _, seven := range []int{-1, 0, 1} {
					n, d := 1, 1
					n, d = mul(n, d, 2, two)
					n, d = mul(n, d, 3, three)
					n, d = mul(n, d, 5, five)
					n, d = mul(n, d, 7, seven)
					rats = append(rats, ratio{n, d})
				}
			}
		}
	}
}
