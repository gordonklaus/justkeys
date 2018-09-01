package voronoi

import (
	"container/heap"
	"math"
)

type Diagram struct {
	Cells map[Point]*Cell
}

func (d *Diagram) Find(p Point) *Cell {
	var c *Cell
	for _, c = range d.Cells {
		break
	}
	for {
		c2 := c.Find(p)
		if c2 == c || c2 == nil {
			return c2
		}
		c = c2
	}
}

type Point struct {
	X, Y float64
}

func (p Point) Add(q Point) Point     { return Point{p.X + q.X, p.Y + q.Y} }
func (p Point) Sub(q Point) Point     { return Point{p.X - q.X, p.Y - q.Y} }
func (p Point) Mul(a float64) Point   { return Point{p.X * a, p.Y * a} }
func (p Point) Len() float64          { return math.Hypot(p.X, p.Y) }
func (p Point) Dot(q Point) float64   { return p.X*q.X + p.Y*q.Y }
func (p Point) Cross(q Point) float64 { return p.X*q.Y - p.Y*q.X }

func (p Point) LessThan(q Point) bool {
	if p.Y == q.Y {
		return p.X < q.X
	}
	return p.Y < q.Y
}

type Cell struct {
	Site  Point
	Edges *HalfEdge
}

func (c *Cell) Contains(p Point) bool {
	return c.Find(p) == c
}

func (c *Cell) Find(p Point) *Cell {
	for edge := c.Edges; ; {
		if v1, v2 := edge.P2.Sub(edge.P1), p.Sub(edge.P2); v1.Cross(v2) < 0 {
			if edge.Pair != nil {
				return edge.Pair.Cell
			}
		}

		edge = edge.Next
		if edge == c.Edges {
			return c
		}
	}
}

type HalfEdge struct {
	Cell             *Cell
	Type             EdgeType
	P1, P2           Point
	Prev, Next, Pair *HalfEdge
}

type EdgeType int8

const (
	Line EdgeType = iota
	IncomingRay
	OutgoingRay
	LineSegment
)

func newEdge(c1, c2 *Cell) (*HalfEdge, *HalfEdge) {
	p1 := c1.Site
	p2 := c2.Site
	d := p2.Sub(p1)
	p := p1.Add(d.Mul(.5))
	d.X, d.Y = -d.Y, d.X
	e1 := &HalfEdge{
		Cell: c1,
		P1:   p,
		P2:   p.Add(d),
	}
	e2 := &HalfEdge{
		Cell: c2,
		P1:   p,
		P2:   p.Sub(d),
		Pair: e1,
	}
	e1.Pair = e2

	c1.Edges = e1
	c2.Edges = e2

	return e1, e2
}

func (e *HalfEdge) setP2(p Point) {
	if e.Type&OutgoingRay == 0 {
		e.P1 = p.Add(e.P1.Sub(e.P2))
	}
	e.P2 = p
	e.Type |= IncomingRay

	if e := e.Pair; e.Type&IncomingRay == 0 {
		e.P2 = p.Add(e.P2.Sub(e.P1))
	}
	e.Pair.P1 = p
	e.Pair.Type |= OutgoingRay
}

func (e *HalfEdge) setNext(e2 *HalfEdge) {
	if e.Cell != e2.Cell {
		panic("mismatch")
	}
	e.Next = e2
	e2.Prev = e
}

func ComputeDiagram(sites []Point) Diagram {
	diagram := Diagram{
		Cells: map[Point]*Cell{},
	}

	if len(sites) == 0 {
		return diagram
	}

	events := make(eventQueue, len(sites))
	for i, s := range sites {
		events[i] = &siteEvent{s}
		diagram.Cells[s] = &Cell{
			Site: s,
		}
	}
	heap.Init(&events)

	beach := beachTreeNode(&beachTreeArc{site: heap.Pop(&events).(*siteEvent).site})

	for events.Len() > 0 {
		switch e := heap.Pop(&events).(type) {
		case *siteEvent:
			old, new := insertArc(&beach, e.site)
			if old.vertexEvent != nil {
				old.vertexEvent.arc = nil
			}
			if ve, ok := newVertexEvent(new.pred.pred); ok {
				heap.Push(&events, ve)
			}
			if ve, ok := newVertexEvent(new.succ.succ); ok {
				heap.Push(&events, ve)
			}

			he1, he2 := newEdge(diagram.Cells[old.site], diagram.Cells[new.site])
			new.pred.edge = he1
			new.succ.edge = he2
		case *vertexEvent:
			if e.arc == nil {
				continue
			}

			removeArc(e.arc)
			left := e.arc.pred.pred
			right := e.arc.succ.succ
			if left.vertexEvent != nil {
				left.vertexEvent.arc = nil
			}
			if right.vertexEvent != nil {
				right.vertexEvent.arc = nil
			}
			if ve, ok := newVertexEvent(left); ok {
				heap.Push(&events, ve)
			}
			if ve, ok := newVertexEvent(right); ok {
				heap.Push(&events, ve)
			}

			leftEdge := e.arc.pred.edge
			rightEdge := e.arc.succ.edge
			he1, he2 := newEdge(diagram.Cells[left.site], diagram.Cells[right.site])
			leftEdge.setP2(e.vertex)
			rightEdge.setP2(e.vertex)
			he2.setP2(e.vertex)
			rightEdge.setNext(leftEdge.Pair)
			leftEdge.setNext(he1)
			he2.setNext(rightEdge.Pair)

			left.succ.edge = he1
		}
	}

	if b, ok := beach.(*beachTreeEdge); ok {
		for {
			e, ok := b.left.(*beachTreeEdge)
			if !ok {
				break
			}
			b = e
		}

		start := b.edge
		prev := start.Pair
		b = b.succ.succ
		for ; b != nil; b = b.succ.succ {
			if b.edge.Type == IncomingRay {
				panic("INCOMING")
			}

			b.edge.setNext(prev)
			prev = b.edge.Pair
		}
		start.setNext(prev)
	}

	for _, cell := range diagram.Cells {
		for edge := cell.Edges; ; {
			if edge.P2.Sub(edge.P1).Len() < 1e-9 {
				edge.Prev.setNext(edge.Next)
				edge.Pair.Prev.setNext(edge.Pair.Next)
				p := edge.P1.Add(edge.P2).Mul(.5)
				edge.Prev.setP2(p)
				edge.Pair.Prev.setP2(p)
				edge.Cell.Edges = edge.Prev
				edge.Pair.Cell.Edges = edge.Pair.Prev
			}
			edge = edge.Next
			if edge == cell.Edges {
				break
			}
		}
	}

	return diagram
}

type eventQueue []event

func (q eventQueue) Len() int            { return len(q) }
func (q eventQueue) Less(i, j int) bool  { return q[i].priority().LessThan(q[j].priority()) }
func (q eventQueue) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *eventQueue) Push(x interface{}) { *q = append(*q, x.(event)) }
func (q *eventQueue) Pop() interface{} {
	x := (*q)[len(*q)-1]
	*q = (*q)[:len(*q)-1]
	return x
}

type event interface {
	priority() Point
}

type siteEvent struct {
	site Point
}

func (e *siteEvent) priority() Point { return e.site }

type vertexEvent struct {
	arc    *beachTreeArc
	vertex Point
	y      float64
}

func newVertexEvent(a *beachTreeArc) (*vertexEvent, bool) {
	if a.pred == nil || a.succ == nil {
		return nil, false
	}

	left := a.pred.pred
	right := a.succ.succ
	v, ok := circumcenter(left.site, a.site, right.site)
	if !ok {
		return nil, false
	}

	// if !onEdge(v.X, left.site, a.site) || !onEdge(v.X, a.site, right.site) {
	// 	return nil, false
	// }

	r := a.site.Sub(v).Len()
	e := &vertexEvent{
		arc:    a,
		vertex: v,
		y:      v.Y + r,
	}
	a.vertexEvent = e

	return e, true
}

func onEdge(x float64, left, right Point) bool {
	if left.Y > right.Y {
		if x >= left.X {
			return true
		}
	} else {
		if x < right.X {
			return true
		}
	}
	return false
}

func sort3(a, b, c Point) [3]Point {
	if c.LessThan(b) {
		b, c = c, b
	}
	if b.LessThan(a) {
		a, b = b, a
	}
	if c.LessThan(b) {
		b, c = c, b
	}
	return [3]Point{a, b, c}
}

func (e *vertexEvent) priority() Point { return Point{e.vertex.X, e.y} }

func circumcenter(a, b, c Point) (Point, bool) {
	ab := b.Sub(a)
	bc := c.Sub(b)
	ca := a.Sub(c)
	d := ab.Cross(bc)
	if d <= 0 {
		return Point{}, false
	}
	t := ab.Dot(bc) / d / 2
	ca_ := Point{-ca.Y, ca.X}
	return a.Add(c).Mul(.5).Add(ca_.Mul(-t)), true
}

type beachTreeNode interface{}

type beachTreeEdge struct {
	parent      *beachTreeEdge
	left, right beachTreeNode
	pred, succ  *beachTreeArc
	edge        *HalfEdge
}

func (e *beachTreeEdge) x(y float64) float64 {
	p1 := e.pred.site
	p2 := e.succ.site

	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	if dy == 0 {
		return p1.X + dx/2
	}

	dy1 := y - p1.Y
	dy2 := y - p2.Y
	b := (p2.X*dy1 - p1.X*dy2) / dy
	b2_c := math.Sqrt(dy1 * dy2 * (dx*dx/(dy*dy) + 1))
	if dy < 0 {
		return b + b2_c
	}
	return b - b2_c
}

type beachTreeArc struct {
	site        Point
	parent      *beachTreeEdge
	pred, succ  *beachTreeEdge
	vertexEvent *vertexEvent
}

func insertArc(beach *beachTreeNode, site Point) (*beachTreeArc, *beachTreeArc) {
	switch b := (*beach).(type) {
	case *beachTreeEdge:
		if leftOfEdge(site, b) {
			// if site.X < b.x(site.Y) {
			return insertArc(&b.left, site)
		} else {
			return insertArc(&b.right, site)
		}
	case *beachTreeArc:
		a := &beachTreeArc{
			site: site,
		}
		left := &beachTreeArc{
			site: b.site,
			pred: b.pred,
		}
		right := &beachTreeArc{
			site: b.site,
			succ: b.succ,
		}
		rightEdge := &beachTreeEdge{
			left:  a,
			right: right,
			pred:  a,
			succ:  right,
		}
		leftEdge := &beachTreeEdge{
			parent: b.parent,
			left:   left,
			right:  rightEdge,
			pred:   left,
			succ:   a,
		}
		rightEdge.parent = leftEdge
		right.parent = rightEdge
		right.pred = rightEdge
		left.parent = leftEdge
		left.succ = leftEdge
		a.parent = rightEdge
		a.pred = leftEdge
		a.succ = rightEdge
		if left.pred != nil {
			left.pred.succ = left
		}
		if right.succ != nil {
			right.succ.pred = right
		}

		*beach = leftEdge
		return b, a
	}

	panic("unreachable")
}

func leftOfEdge(site Point, e *beachTreeEdge) bool {
	p1 := e.pred.site
	p2 := e.succ.site

	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	if dy == 0 {
		return site.X < p1.X+dx/2
	}

	// TODO?: dx == 0

	if dy > 0 && site.X > p2.X {
		return false
	}
	if dy < 0 && site.X < p1.X {
		return true
	}
	return parabola(p1, site) > parabola(p2, site)
}

func parabola(p, site Point) float64 {
	dx := site.X - p.X
	return (dx*dx/(p.Y-site.Y) + p.Y + site.Y) / 2
}

func removeArc(a *beachTreeArc) {
	p := a.parent
	var n beachTreeNode
	if a == p.left {
		n = p.right
		a.pred.succ = p.succ
		p.succ.pred = a.pred
	} else {
		n = p.left
		p.pred.succ = a.succ
		a.succ.pred = p.pred
	}
	switch n := n.(type) {
	case *beachTreeEdge:
		n.parent = p.parent
	case *beachTreeArc:
		n.parent = p.parent
	}
	if p == p.parent.left {
		p.parent.left = n
	} else {
		p.parent.right = n
	}
}
