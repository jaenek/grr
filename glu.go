package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func createWindow(w, h int, title string) (*glfw.Window, error) {
	// Initialize glfw
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	// Create window
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(W, H, title, nil, nil)
	if err != nil {
		return nil, err
	}

	// Create context
	window.MakeContextCurrent()

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// Configure global opengl state
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	return window, nil
}

type shader struct {
	fragmentSource string
	vertexSource   string
	vertexData     []float32
	vao            uint32
	vbo            uint32
}

func createShader(fragmentFilename, vertexFilename string) (*shader, error) {
	s := &shader{}

	b, err := ioutil.ReadFile(fragmentFilename)
	if err != nil {
		return &shader{}, err
	}
	s.fragmentSource = string(b) + "\x00"

	b, err = ioutil.ReadFile(vertexFilename)
	if err != nil {
		return &shader{}, err
	}
	s.vertexSource = string(b) + "\x00"

	return s, nil
}

func (s shader) createProgram() (uint32, error) {
	vertexShader, err := compileShader(s.vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(s.fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

type shaderProgram uint32

func (p shaderProgram) loadVertexData(s *shader) {
	gl.GenVertexArrays(1, &s.vao)
	gl.GenBuffers(1, &s.vbo)

	gl.BindBuffer(gl.ARRAY_BUFFER, s.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(s.vertexData)*4, gl.Ptr(s.vertexData), gl.STATIC_DRAW)

	gl.BindVertexArray(s.vao)

	// Vertex attribute
	vertAttrib := uint32(gl.GetAttribLocation(uint32(p), gl.Str("aPos\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))

	/* Texture coordinate attribute
	texcoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("aNormal\x00")))
	gl.EnableVertexAttribArray(texcoordAttrib)
	gl.VertexAttribPointer(texcoordAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))
	*/

	// Normal attribute
	normAttrib := uint32(gl.GetAttribLocation(uint32(p), gl.Str("aNormal\x00")))
	gl.EnableVertexAttribArray(normAttrib)
	gl.VertexAttribPointer(normAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(5*4))
}

func (p shaderProgram) UniformMatrix4fv(name string, vec mgl32.Mat4) int32 {
	uniform := gl.GetUniformLocation(uint32(p), gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(uniform, 1, false, &vec[0])

	return uniform
}

func (p shaderProgram) Uniform3fv(name string, vec mgl32.Vec3) int32 {
	uniform := gl.GetUniformLocation(uint32(p), gl.Str(name+"\x00"))
	gl.Uniform3fv(uniform, 1, &vec[0])

	return uniform
}

func (p shaderProgram) bindFragDataLocation() {
	gl.BindFragDataLocation(uint32(p), 0, gl.Str("FragColor\x00"))
}
