package main

import (
	"image"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/gl"
)

type Program struct {
	glctx   gl.Context
	program gl.Program
	mvp     gl.Uniform
	pos     gl.Attrib
	color   gl.Attrib

	proj, view mgl32.Mat4
}

func NewProgram(glctx gl.Context, sz image.Point) (*Program, error) {
	const vertexShader = `#version 100
		uniform mat4 mvp;
		attribute vec2 pos;
		attribute vec4 color;
		varying vec4 vColor;

		void main() {
			gl_Position = mvp * vec4(pos, 0, 1);
			vColor = color;
		}`

	const fragmentShader = `#version 100
		precision mediump float;
		varying vec4 vColor;

		void main() {
			gl_FragColor = vColor;
		}`

	program, err := glutil.CreateProgram(glctx, vertexShader, fragmentShader)
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return nil, err
	}

	mvp := glctx.GetUniformLocation(program, "mvp")
	pos := glctx.GetAttribLocation(program, "pos")
	color := glctx.GetAttribLocation(program, "color")

	p := &Program{
		glctx:   glctx,
		program: program,
		mvp:     mvp,
		pos:     pos,
		color:   color,
	}
	p.Size(sz)
	p.LookAt(tonicPitch+.5, .5)

	return p, nil
}

func (p *Program) Release() {
	p.glctx.DeleteProgram(p.program)
}

func (p *Program) Size(sz image.Point) {
	width := float32(1.2)
	height := float32(1)
	p.proj = mgl32.Ortho2D(-width/2, width/2, -height/2, height/2)
}

func (p *Program) LookAt(x, y float32) {
	p.view = mgl32.LookAt(x, y, 1, x, y, 0, 0, 1, 0)
}

func (p *Program) Clip2World(x, y float32) (float32, float32) {
	v := p.proj.Mul4(p.view).Inv().Mul4x1(mgl32.Vec4{x, y, 0, 1})
	return v.X(), v.Y()
}

func (p *Program) Draw(buffer *VertexBuffer, model mgl32.Mat4) {
	p.glctx.UseProgram(p.program)

	mvp := p.proj.Mul4(p.view).Mul4(model)
	p.glctx.UniformMatrix4fv(p.mvp, mvp[:])

	buffer.Draw(p.pos, p.color)
}
