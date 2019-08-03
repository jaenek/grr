package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const W = 800
const H = 600

func cos(f float32) float32 { return float32(math.Cos(float64(f))) }
func sin(f float32) float32 { return float32(math.Sin(float64(f))) }

var Up mgl32.Vec3 = mgl32.Vec3{0, 1, 0}
var cameraPos mgl32.Vec3 = mgl32.Vec3{0, 1, 10}
var cameraFront mgl32.Vec3 = mgl32.Vec3{0, 0, -1}
var cameraRight mgl32.Vec3 = mgl32.Vec3{0, 0, 0}
var cameraUp mgl32.Vec3 = mgl32.Vec3{0, 1, 0}
var Yaw, Pitch float32 = -90, 0
var lastX, lastY float64 = W / 2.0, H / 2.0
var firstMouse bool = true

func main() {
	runtime.LockOSThread()

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(W, H, "grr", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	window.SetCursorPosCallback(mouse_callback)
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	// Configure global opengl state
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Build and compile the shader program
	program, err := newProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}

	gl.UseProgram(program)

	vertexData := loadObject("objects/cube.obj")

	// Configure the cube's VAO (and VBO)
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertexData)*4, gl.Ptr(vertexData), gl.STATIC_DRAW)

	gl.BindVertexArray(vao)

	// vertex attribute
	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("aPos\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(0))

	// coordinate attribute
	//coordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("aNormal\x00")))
	//gl.EnableVertexAttribArray(coordAttrib)
	//gl.VertexAttribPointer(coordAttrib, 2, gl.FLOAT, false, 8*4, gl.PtrOffset(3*4))

	// normal attribute
	normAttrib := uint32(gl.GetAttribLocation(program, gl.Str("aNormal\x00")))
	gl.EnableVertexAttribArray(normAttrib)
	gl.VertexAttribPointer(normAttrib, 3, gl.FLOAT, false, 8*4, gl.PtrOffset(5*4))

	// Set up matrices and link to shader program
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(W)/H, 0.1, 100.0)
	projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	camera := mgl32.LookAtV(cameraPos, cameraFront, cameraUp)
	cameraUniform := gl.GetUniformLocation(program, gl.Str("view\x00"))
	gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

	model := mgl32.Ident4()
	modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	objectColor := []float32{1, 0.5, 0.31}
	objectColorUniform := gl.GetUniformLocation(program, gl.Str("objectColor\x00"))
	gl.Uniform3fv(objectColorUniform, 1, &objectColor[0])

	lightColor := []float32{1, 1, 1}
	lightColorUniform := gl.GetUniformLocation(program, gl.Str("lightColor\x00"))
	gl.Uniform3fv(lightColorUniform, 1, &lightColor[0])

	lightPosUniform := gl.GetUniformLocation(program, gl.Str("lightPos\x00"))
	gl.Uniform3fv(lightPosUniform, 1, &cameraPos[0])

	cameraPosUniform := gl.GetUniformLocation(program, gl.Str("viewPos\x00"))
	gl.Uniform3fv(cameraPosUniform, 1, &cameraPos[0])

	gl.BindFragDataLocation(program, 0, gl.Str("FragColor\x00"))

	previousTime := glfw.GetTime()

	fmt.Print(len(vertexData))
	for !window.ShouldClose() {

		// Update
		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		// Render
		gl.ClearColor(0.1, 0.1, 0.1, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		cameraPos = processInputs(window, cameraPos, elapsed)
		camera = mgl32.LookAtV(cameraPos, cameraFront.Add(cameraPos), cameraUp)

		gl.UseProgram(program)
		gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])
		gl.Uniform3fv(lightPosUniform, 1, &cameraPos[0])
		gl.Uniform3fv(cameraPosUniform, 1, &cameraPos[0])

		gl.BindVertexArray(vao)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertexData)/8))

		// Maintenance
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

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

	cameraFront = mgl32.Vec3{cos(mgl32.DegToRad(Yaw)) * cos(mgl32.DegToRad(Pitch)), sin(mgl32.DegToRad(Pitch)), sin(mgl32.DegToRad(Yaw)) * cos(mgl32.DegToRad(Pitch))}.Normalize()
}

func processInputs(window *glfw.Window, cameraPos mgl32.Vec3, deltaTime float64) mgl32.Vec3 {
	cameraSpeed := float32(1.6 * deltaTime)

	if window.GetKey(glfw.KeyLeftShift) == glfw.Press {
		cameraSpeed *= 2
	}
	if window.GetKey(glfw.KeyW) == glfw.Press {
		cameraPos = cameraPos.Add(cameraFront.Mul(cameraSpeed))
	}
	if window.GetKey(glfw.KeyS) == glfw.Press {
		cameraPos = cameraPos.Sub(cameraFront.Mul(cameraSpeed))
	}
	if window.GetKey(glfw.KeyA) == glfw.Press {
		cameraPos = cameraPos.Sub(cameraFront.Cross(cameraUp).Normalize().Mul(cameraSpeed))
	}
	if window.GetKey(glfw.KeyD) == glfw.Press {
		cameraPos = cameraPos.Add(cameraFront.Cross(cameraUp).Normalize().Mul(cameraSpeed))
	}
	return cameraPos
}

func newProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
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

var vertexShader = `
#version 330 core
in vec3 aPos;
in vec3 aNormal;

out vec3 FragPos;
out vec3 Normal;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main()
{
    FragPos = vec3(model * vec4(aPos, 1.0));
    Normal = aNormal;

    gl_Position = projection * view * vec4(FragPos, 1.0);
}
` + "\x00"

var fragmentShader = `
#version 330 core
out vec4 FragColor;

in vec3 Normal;
in vec3 FragPos;

uniform vec3 lightPos;
uniform vec3 viewPos;
uniform vec3 lightColor;
uniform vec3 objectColor;

float luma(vec3 color) {
  return dot(color, vec3(0.299, 0.587, 0.114));
}

float dither8x8(vec2 position, float brightness) {
  int x = int(mod(position.x, 8.0));
  int y = int(mod(position.y, 8.0));
  int index = x + y * 8;
  float limit = 0.0;

  if (x < 8) {
    if (index == 0) limit = 0.015625;
    if (index == 1) limit = 0.515625;
    if (index == 2) limit = 0.140625;
    if (index == 3) limit = 0.640625;
    if (index == 4) limit = 0.046875;
    if (index == 5) limit = 0.546875;
    if (index == 6) limit = 0.171875;
    if (index == 7) limit = 0.671875;
    if (index == 8) limit = 0.765625;
    if (index == 9) limit = 0.265625;
    if (index == 10) limit = 0.890625;
    if (index == 11) limit = 0.390625;
    if (index == 12) limit = 0.796875;
    if (index == 13) limit = 0.296875;
    if (index == 14) limit = 0.921875;
    if (index == 15) limit = 0.421875;
    if (index == 16) limit = 0.203125;
    if (index == 17) limit = 0.703125;
    if (index == 18) limit = 0.078125;
    if (index == 19) limit = 0.578125;
    if (index == 20) limit = 0.234375;
    if (index == 21) limit = 0.734375;
    if (index == 22) limit = 0.109375;
    if (index == 23) limit = 0.609375;
    if (index == 24) limit = 0.953125;
    if (index == 25) limit = 0.453125;
    if (index == 26) limit = 0.828125;
    if (index == 27) limit = 0.328125;
    if (index == 28) limit = 0.984375;
    if (index == 29) limit = 0.484375;
    if (index == 30) limit = 0.859375;
    if (index == 31) limit = 0.359375;
    if (index == 32) limit = 0.0625;
    if (index == 33) limit = 0.5625;
    if (index == 34) limit = 0.1875;
    if (index == 35) limit = 0.6875;
    if (index == 36) limit = 0.03125;
    if (index == 37) limit = 0.53125;
    if (index == 38) limit = 0.15625;
    if (index == 39) limit = 0.65625;
    if (index == 40) limit = 0.8125;
    if (index == 41) limit = 0.3125;
    if (index == 42) limit = 0.9375;
    if (index == 43) limit = 0.4375;
    if (index == 44) limit = 0.78125;
    if (index == 45) limit = 0.28125;
    if (index == 46) limit = 0.90625;
    if (index == 47) limit = 0.40625;
    if (index == 48) limit = 0.25;
    if (index == 49) limit = 0.75;
    if (index == 50) limit = 0.125;
    if (index == 51) limit = 0.625;
    if (index == 52) limit = 0.21875;
    if (index == 53) limit = 0.71875;
    if (index == 54) limit = 0.09375;
    if (index == 55) limit = 0.59375;
    if (index == 56) limit = 1.0;
    if (index == 57) limit = 0.5;
    if (index == 58) limit = 0.875;
    if (index == 59) limit = 0.375;
    if (index == 60) limit = 0.96875;
    if (index == 61) limit = 0.46875;
    if (index == 62) limit = 0.84375;
    if (index == 63) limit = 0.34375;
  }

  return brightness < limit ? 0.0 : 1.0;
}

vec3 dither8x8(vec2 position, vec3 color) {
  return color * dither8x8(position, luma(color));
}

void main()
{
    // ambient
    float ambientStrength = 0.1;
    vec3 ambient = ambientStrength * lightColor;

    // diffuse
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(lightPos - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * lightColor;

    // specular
    float specularStrength = 0.5;
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 32);
    vec3 specular = specularStrength * spec * lightColor;

    vec3 result = (ambient + diffuse + specular) * objectColor;
    FragColor = vec4(dither8x8(gl_FragCoord.xy, result), 1.0);
}
` + "\x00"

func loadObject(filename string) []float32 {
	var verts []float32
	var coords []float32
	var norms []float32

	parseFloat := func(words []string) {
		for _, w := range words[1:] {
			f, err := strconv.ParseFloat(w, 32)
			if err != nil {
				log.Fatalf("Error: Couldn't parse float values from a object file %s,\n", filename)
			}

			switch words[0] {
			case "v":
				verts = append(verts, float32(f))
			case "vt":
				coords = append(coords, float32(f))
			case "vn":
				norms = append(norms, float32(f))
			}
		}
	}

	parseInt := func(words []string) []float32 {
		f := func(c rune) bool {
			if c == '/' {
				return true
			}
			return false
		}

		var vao []float32
		for _, w := range words {
			fields := strings.FieldsFunc(w, f)
			for n, field := range fields {
				i, err := strconv.ParseInt(field, 10, 64)
				if err != nil {
					log.Fatalf("Error: Couldn't parse face elements from a object file %s,\n", filename)
				}

				switch n {
				case 0:
					i = 3 * (i - 1)
					if len(fields) == 1 {
						vao = append(append(vao, verts[i:i+3]...), []float32{0, 0, 0, 0, 0}...)
					} else {
						vao = append(vao, verts[i:i+3]...)
					}
				case 1:
					if len(fields) == 2 {
						i = 3 * (i - 1)
						vao = append(append(vao, []float32{0.0, 0.0}...), norms[i:i+3]...)
					} else {
						i = 2 * (i - 1)
						vao = append(vao, coords[i:i+2]...)
					}
				case 2:
					i = 3 * (i - 1)
					vao = append(vao, norms[i:i+3]...)
				}
			}
		}
		return vao
	}

	if filepath.Ext(filename) != ".obj" {
		log.Fatalln("Error: Couldn't load a file %s,\n%s extension is not supported.\n", filename, filepath.Ext(filename))
	}

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var vao []float32
	s := bufio.NewScanner(file)
	for s.Scan() {
		words := strings.Fields(s.Text())

		// add vp handling
		switch words[0] {
		case "v":
			parseFloat(words)
		case "vt":
			parseFloat(words)
		case "vn":
			parseFloat(words)
		case "f":
			vao = append(vao, parseInt(words[1:])...)
		}
	}

	if err := s.Err(); err != nil {

		panic(err)
	}

	return vao
}
