package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	W = 800
	H = 600
)

func main() {
	runtime.LockOSThread()

	window, err := createWindow(W, H, "grr")
	if err != nil {
		log.Fatalln(err)
	}
	defer glfw.Terminate()

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	// Load and compile the shader program
	s, err := createShader("shader.frag", "shader.vert")
	if err != nil {
		log.Fatalln(err)
	}

	program, err := s.createProgram()
	if err != nil {
		log.Fatalln(err)
	}

	gl.UseProgram(program)

	// Load objects
	//loadObject(s, "objects/cube.obj")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	err = loadObject(s, "objects/monkey.obj")
	if err != nil {
		log.Fatalln(err)
	}

	p := shaderProgram(program)
	p.loadVertexData(s)

	// Define Uniforms
	c := createCamera(window, p)
	p.Uniform3fv("objectColor", mgl32.Vec3{0.91, 0.58, 0.49})

	p.bindFragDataLocation()

	// Main loop
	previousTime := glfw.GetTime()

	for !window.ShouldClose() {
		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		// Render
		gl.ClearColor(0.49, 0.83, 0.91, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		c.update(elapsed)

		gl.BindVertexArray(s.vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(s.vertexData)/8))

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
