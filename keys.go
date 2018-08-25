package main

import (
	"math"

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

	pressed map[float64]*Key
	diagram voronoi.Diagram
}

func NewKeys(glctx gl.Context, program *Program) *Keys {
	k := &Keys{
		glctx:   glctx,
		program: program,
		// buffer,
		pressed: map[float64]*Key{},
	}
	k.buildDiagram()
	return k
}

func (k *Keys) buildDiagram() {
	playingPitches := []float64{}
	for p := range k.pressed {
		playingPitches = append(playingPitches, p)
	}
	if len(playingPitches) == 0 {
		playingPitches = []float64{tonicPitch}
	}

	// sites := []voronoi.Point{
	// 	{8, 0},
	// 	{8.5, 0.05},
	// 	{8.25, 0.25},
	// 	{8.2, .5},
	// 	{8.45, .4},
	// }

	sites := []voronoi.Point{}
	for _, r := range rats {
		pitch := tonicPitch + math.Log2(float64(r.a)/float64(r.b))
		if pitch < 4 || pitch > 14 {
			continue
		}
		diss := totalDissonance(pitch, playingPitches)
		sites = append(sites, voronoi.Point{pitch, diss})
	}
	// bbox := voronoi.NewBBox(4, 14, -.5, 1.5)
	k.diagram = voronoi.ComputeDiagram(sites)
}

func (k *Keys) Release() {
	// k.buffer.Release()
}

func (k *Keys) Draw() {
	vs := []Vertex{}
	ps := []Vertex{}
	for _, cell := range k.diagram.Cells {
		color := mgl32.Vec4{.9, .9, .9, 1}
		if _, ok := k.pressed[cell.Site.X]; ok {
			color = mgl32.Vec4{.6, .6, .6, 1}
		}
		site := voronoiVertexToVec2(cell.Site)
		ps = append(ps,
			Vertex{Position: mgl32.Vec2{site[0] - .005, site[1]}, Color: color},
			Vertex{Position: mgl32.Vec2{site[0] + .005, site[1]}, Color: color},
			Vertex{Position: mgl32.Vec2{site[0], site[1] - .005}, Color: color},
			Vertex{Position: mgl32.Vec2{site[0], site[1] + .005}, Color: color},
		)
		for edge := cell.Edges; ; {
			if edge.Type == voronoi.LineSegment {
				va := voronoiVertexToVec2(edge.P1)
				vb := voronoiVertexToVec2(edge.P2)
				vs = append(vs,
					Vertex{Position: site.Add(va.Sub(site).Mul(.95)), Color: color},
					Vertex{Position: site.Add(vb.Sub(site).Mul(.95)), Color: color},
				)
			}
			edge = edge.Next
			if edge == cell.Edges {
				break
			}
		}
	}
	buffer := NewVertexBuffer(k.glctx, gl.LINES, vs)
	pbuf := NewVertexBuffer(k.glctx, gl.LINES, ps)

	k.glctx.LineWidth(1)
	k.program.Draw(buffer, mgl32.Ident4())
	k.program.Draw(pbuf, mgl32.Ident4())

	pbuf.Release()
	buffer.Release()
}

func voronoiVertexToVec2(v voronoi.Point) mgl32.Vec2 {
	return mgl32.Vec2{float32(v.X), float32(v.Y)}
}

func (k *Keys) Touch(e touch.Event) {
	switch e.Type {
	case touch.TypeBegin:
		if pitch, ok := k.pitchForTouch(float64(e.X), float64(e.Y)); ok && k.pressed[pitch] == nil {
			tone := NewTone(pitch)
			tones.AddTone(tone)
			k.pressed[pitch] = &Key{e.Sequence, tone}
			k.buildDiagram()
		}
	case touch.TypeEnd:
		for pitch, key := range k.pressed {
			if key.seq == e.Sequence {
				key.tone.Release()
				delete(k.pressed, pitch)
				k.buildDiagram()
			}
		}
	}
}

type Key struct {
	seq  touch.Sequence
	tone *Tone
}

func (k *Keys) pitchForTouch(x, y float64) (float64, bool) {
	for _, cell := range k.diagram.Cells {
		if cell.Contains(voronoi.Point{x, y}) {
			return cell.Site.X, true
		}
	}
	return 0, false
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

	m := 4 * a1 * a2 / ((a1 + a2) * (a1 + a2))
	squareMean := (a1 + a2) * mathext.CompleteE(m) / (math.Pi / 2)
	squareMean *= squareMean

	stddev := math.Sqrt(meanSquare - squareMean)
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
