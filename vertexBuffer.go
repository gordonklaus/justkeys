package main

import (
	"encoding/binary"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

type VertexBuffer struct {
	glctx  gl.Context
	buffer gl.Buffer
	mode   gl.Enum
	length int
}

type Vertex struct {
	Position mgl32.Vec2
	Color    mgl32.Vec4
}

const coordsPerVertex = 6

func NewVertexBuffer(glctx gl.Context, mode gl.Enum, vs []Vertex) *VertexBuffer {
	buffer := glctx.CreateBuffer()

	data := make([]float32, coordsPerVertex*len(vs))
	for i, v := range vs {
		data[coordsPerVertex*i+0] = v.Position[0]
		data[coordsPerVertex*i+1] = v.Position[1]
		data[coordsPerVertex*i+2] = v.Color[0]
		data[coordsPerVertex*i+3] = v.Color[1]
		data[coordsPerVertex*i+4] = v.Color[2]
		data[coordsPerVertex*i+5] = v.Color[3]
	}
	glctx.BindBuffer(gl.ARRAY_BUFFER, buffer)
	glctx.BufferData(gl.ARRAY_BUFFER, f32.Bytes(binary.LittleEndian, data...), gl.STATIC_DRAW)

	return &VertexBuffer{glctx, buffer, mode, len(vs)}
}

func (b *VertexBuffer) Release() {
	b.glctx.DeleteBuffer(b.buffer)
}

func (b *VertexBuffer) Draw(pos, color gl.Attrib) {
	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buffer)
	b.glctx.VertexAttribPointer(pos, 2, gl.FLOAT, false, 4*coordsPerVertex, 0)
	b.glctx.EnableVertexAttribArray(pos)
	b.glctx.VertexAttribPointer(color, 4, gl.FLOAT, false, 4*coordsPerVertex, 4*2)
	b.glctx.EnableVertexAttribArray(color)

	b.glctx.DrawArrays(b.mode, 0, b.length)
}
