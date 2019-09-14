package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type camera struct {
	window  *glfw.Window
	program shaderProgram
	Up      mgl32.Vec3
	Pos     mgl32.Vec3
	Front   mgl32.Vec3
	Right   mgl32.Vec3
}

var c camera = camera{
	Up:    mgl32.Vec3{0, 1, 0},
	Pos:   mgl32.Vec3{0, 1, 10},
	Front: mgl32.Vec3{0, 0, -1},
	Right: mgl32.Vec3{0, 0, 0},
}

func createCamera(w *glfw.Window, p shaderProgram) *camera {
	c.window = w
	c.program = p

	w.SetCursorPosCallback(mouse_callback)
	w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	// Set up matrices and link to shader program
	p.UniformMatrix4fv("projection", mgl32.Perspective(mgl32.DegToRad(45.0), float32(W)/H, 0.1, 100.0))
	p.UniformMatrix4fv("view", mgl32.LookAtV(c.Pos, c.Front, c.Up))
	p.UniformMatrix4fv("model", mgl32.Ident4())

	p.Uniform3fv("lightColor", mgl32.Vec3{1, 1, 1})
	p.Uniform3fv("lightPos", mgl32.Vec3{0})
	p.Uniform3fv("viewPos", mgl32.Vec3{0})

	return &c
}

func (c camera) update(elapsed float64) {
	c.Pos = processInputs(c.window, c.Pos, elapsed)
	view := mgl32.LookAtV(c.Pos, c.Front.Add(c.Pos), c.Up)

	p := uint32(c.program)
	gl.UseProgram(p)
	viewUniform := gl.GetUniformLocation(p, gl.Str("view\x00"))
	viewPosUniform := gl.GetUniformLocation(p, gl.Str("viewPos\x00"))
	lightPosUniform := gl.GetUniformLocation(p, gl.Str("lightPos\x00"))

	gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])
	gl.Uniform3fv(lightPosUniform, 1, &c.Pos[0])
	gl.Uniform3fv(viewPosUniform, 1, &c.Pos[0])
}

var (
	lastX      float64 = W / 2.0
	lastY      float64 = H / 2.0
	Yaw        float32 = -90
	Pitch      float32 = 0
	firstMouse bool    = true
)

func mouse_callback(window *glfw.Window, xpos, ypos float64) {
	if firstMouse {
		lastX = xpos
		lastY = ypos
		firstMouse = false
	}

	xoffset := xpos - lastX
	yoffset := lastY - ypos

	lastX = xpos
	lastY = ypos

	xoffset *= 0.1
	yoffset *= 0.1

	Yaw += float32(xoffset)
	Pitch += float32(yoffset)

	if Pitch > 89 {
		Pitch = 89
	}
	if Pitch < -89 {
		Pitch = -89
	}

	c.Front = mgl32.Vec3{
		cos(mgl32.DegToRad(Yaw)) * cos(mgl32.DegToRad(Pitch)),
		sin(mgl32.DegToRad(Pitch)),
		sin(mgl32.DegToRad(Yaw)) * cos(mgl32.DegToRad(Pitch)),
	}.Normalize()
}

func processInputs(window *glfw.Window, cameraPos mgl32.Vec3, deltaTime float64) mgl32.Vec3 {
	cameraSpeed := float32(1.6 * deltaTime)

	if window.GetKey(glfw.KeyLeftShift) == glfw.Press {
		cameraSpeed *= 2
	}
	if window.GetKey(glfw.KeyW) == glfw.Press {
		c.Pos = c.Pos.Add(c.Front.Mul(cameraSpeed))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		c.Pos = c.Pos.Sub(c.Front.Mul(cameraSpeed))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		c.Pos = c.Pos.Sub(
			c.Front.Cross(c.Up).Normalize().Mul(cameraSpeed))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		c.Pos = c.Pos.Add(
			c.Front.Cross(c.Up).Normalize().Mul(cameraSpeed))
	}
	return cameraPos
}
