package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func loadObject(s *shader, filename string) error {
	var verts, coords, norms []float32

	parseFloat := func(words []string) error {
		for _, w := range words[1:] {
			f, err := strconv.ParseFloat(w, 32)
			if err != nil {
				return fmt.Errorf("Error: Couldn't parse float values from a object file %s,", filename)
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
		return nil
	}

	parseInt := func(words []string) ([]float32, error) {
		if len(norms) == 0 {
			return []float32{0}, fmt.Errorf("Error: No normals in a object file %s,\n", filename)
		}

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
					return []float32{0}, fmt.Errorf("Error: Couldn't parse face elements from a object file %s,\n", filename)
				}

				switch n {
				case 0:
					i = 3 * (i - 1)
					vao = append(vao, verts[i:i+3]...)
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
		return vao, nil
	}

	if filepath.Ext(filename) != ".obj" {
		return fmt.Errorf("Error: Couldn't load a file %s,\n%s extension is not supported.\n", filename, filepath.Ext(filename))
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())

		// add vp handling
		switch words[0] {
		case "v":
			fallthrough
		case "vt":
			fallthrough
		case "vn":
			err := parseFloat(words)
			if err != nil {
				return err
			}
		case "f":
			ints, err := parseInt(words[1:])
			if err != nil {
				return err
			}
			s.vertexData = append(s.vertexData, ints...)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
