// Copyright 2011-2014 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	mathRand "math/rand/v2"
)

const (
	// Standard width and height of a captcha image.
	StdWidth  = 240
	StdHeight = 80
	// Maximum absolute skew factor of a single digit.
	maxSkew = 0.7
	// Number of background circles.
	circleCount = 20
)

type Image struct {
	*image.Paletted
	numWidth  int
	numHeight int
	dotSize   int
	rng       *mathRand.Rand
}

// Int returns a pseudorandom int in the half-open interval [from, to].
func (m *Image) Int(from, to int) int {
	return from + m.rng.IntN(to-from+1)
}

// Float returns a pseudorandom float64 in the half-open interval [from, to).
func (m *Image) Float(from, to float64) float64 {
	return from + (to-from)*m.rng.Float64()
}

// NewImage returns a new captcha image of the given width and height with the
// given digits, where each digit must be in range 0-9.
func NewImage(id string, digits []byte, width, height int) *Image {
	m := new(Image)

	// Initialize PRNG.
	m.rng = mathRand.New(mathRand.NewChaCha8(deriveSeed(imageSeedPurpose, id, digits)))

	m.Paletted = image.NewPaletted(image.Rect(0, 0, width, height), m.getRandomPalette())
	m.calculateSizes(width, height, len(digits))
	// Randomly position captcha inside the image.
	maxx := width - (m.numWidth+m.dotSize)*len(digits) - m.dotSize
	maxy := height - m.numHeight - m.dotSize*2
	var border int
	if width > height {
		border = height / 5
	} else {
		border = width / 5
	}
	x := m.Int(border, maxx-border)
	y := m.Int(border, maxy-border)
	// Draw digits.
	for _, n := range digits {
		m.drawDigit(font[n], x, y)
		x += m.numWidth + m.dotSize
	}
	// Draw strike-through line.
	m.strikeThrough()
	// Apply wave distortion.
	m.distort(m.Float(5, 10), m.Float(100, 200))
	// Fill image with random circles.
	m.fillWithCircles(circleCount, m.dotSize)
	// Flip the image horizontally.
	m.flipX()
	return m
}

func (m *Image) getRandomPalette() color.Palette {
	p := make([]color.Color, circleCount+1)
	// Transparent color.
	p[0] = color.RGBA{0xFF, 0xFF, 0xFF, 0x00}
	// Primary color.
	prim := color.RGBA{
		uint8(m.rng.IntN(129)),
		uint8(m.rng.IntN(129)),
		uint8(m.rng.IntN(129)),
		0xFF,
	}
	p[1] = prim
	// Circle colors.
	for i := 2; i <= circleCount; i++ {
		p[i] = m.randomBrightness(prim, 255)
	}
	return p
}

// encodedPNG encodes an image to PNG and returns
// the result as a byte slice.
func (m *Image) encodedPNG() []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, m.Paletted); err != nil {
		panic(err.Error())
	}
	return buf.Bytes()
}

// WriteTo writes captcha image in PNG format into the given writer.
func (m *Image) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(m.encodedPNG())
	return int64(n), err
}

func (m *Image) calculateSizes(width, height, ncount int) {
	// Goal: fit all digits inside the image.
	var border int
	if width > height {
		border = height / 4
	} else {
		border = width / 4
	}
	// Convert everything to floats for calculations.
	w := float64(width - border*2)
	h := float64(height - border*2)
	// fw takes into account 1-dot spacing between digits.
	fw := float64(fontWidth + 1)
	fh := float64(fontHeight)
	nc := float64(ncount)
	// Calculate the width of a single digit taking into account only the
	// width of the image.
	nw := w / nc
	// Calculate the height of a digit from this width.
	nh := nw * fh / fw
	// Digit too high?
	if nh > h {
		// Fit digits based on height.
		nh = h
		nw = fw / fh * nh
	}
	// Calculate dot size.
	m.dotSize = max(int(nh/fh), 1)
	// Save everything, making the actual width smaller by 1 dot to account
	// for spacing between digits.
	m.numWidth = int(nw) - m.dotSize
	m.numHeight = int(nh)
}

func (m *Image) drawHorizLine(fromX, toX, y int, colorIdx uint8) {
	for x := fromX; x <= toX; x++ {
		m.SetColorIndex(x, y, colorIdx)
	}
}

func (m *Image) drawCircle(x, y, radius int, colorIdx uint8) {
	f := 1 - radius
	dfx := 1
	dfy := -2 * radius
	xo := 0
	yo := radius

	m.SetColorIndex(x, y+radius, colorIdx)
	m.SetColorIndex(x, y-radius, colorIdx)
	m.drawHorizLine(x-radius, x+radius, y, colorIdx)

	for xo < yo {
		if f >= 0 {
			yo--
			dfy += 2
			f += dfy
		}
		xo++
		dfx += 2
		f += dfx
		m.drawHorizLine(x-xo, x+xo, y+yo, colorIdx)
		m.drawHorizLine(x-xo, x+xo, y-yo, colorIdx)
		m.drawHorizLine(x-yo, x+yo, y+xo, colorIdx)
		m.drawHorizLine(x-yo, x+yo, y-xo, colorIdx)
	}
}

func (m *Image) fillWithCircles(n, maxradius int) {
	maxx := m.Bounds().Max.X
	maxy := m.Bounds().Max.Y
	for range n {
		colorIdx := uint8(m.Int(1, circleCount-1))
		r := m.Int(1, maxradius)
		m.drawCircle(m.Int(r, maxx-r), m.Int(r, maxy-r), r, colorIdx)
	}
}

func (m *Image) strikeThrough() {
	maxx := m.Bounds().Max.X
	maxy := m.Bounds().Max.Y
	y := m.Int(maxy/3, maxy-maxy/3)
	amplitude := m.Float(5, 20)
	period := m.Float(80, 180)
	dx := 2.0 * math.Pi / period
	for x := range maxx {
		xo := amplitude * math.Cos(float64(y)*dx)
		yo := amplitude * math.Sin(float64(x)*dx)
		for yn := 0; yn < m.dotSize; yn++ {
			r := m.rng.IntN(m.dotSize + 1)
			m.drawCircle(x+int(xo), y+int(yo)+(yn*m.dotSize), r/2, 1)
		}
	}
}

func (m *Image) drawDigit(digit []byte, x, y int) {
	skf := m.Float(-maxSkew, maxSkew)
	xs := float64(x)
	r := m.dotSize / 2
	y += m.Int(-r, r)
	for yo := range fontHeight {
		for xo := range fontWidth {
			if digit[yo*fontWidth+xo] != blackChar {
				continue
			}
			m.drawCircle(x+xo*m.dotSize, y+yo*m.dotSize, r, 1)
		}
		xs += skf
		x = int(xs)
	}
}

func (m *Image) distort(amplude float64, period float64) {
	w := m.Bounds().Max.X
	h := m.Bounds().Max.Y

	oldm := m.Paletted
	newm := image.NewPaletted(image.Rect(0, 0, w, h), oldm.Palette)

	dx := 2.0 * math.Pi / period
	for x := range w {
		for y := range h {
			xo := amplude * math.Sin(float64(y)*dx)
			yo := amplude * math.Cos(float64(x)*dx)
			newm.SetColorIndex(x, y, oldm.ColorIndexAt(x+int(xo), y+int(yo)))
		}
	}
	m.Paletted = newm
}

func (m *Image) flipX() {
	w := m.Bounds().Max.X
	h := m.Bounds().Max.Y

	src := m.Paletted
	dst := image.NewPaletted(image.Rect(0, 0, w, h), src.Palette)

	for x := range w {
		for y := range h {
			xo := w - x - 1
			dst.SetColorIndex(x, y, src.ColorIndexAt(xo, y))
		}
	}
	m.Paletted = dst
}

func (m *Image) randomBrightness(c color.RGBA, maxB uint8) color.RGBA {
	minc := min(min(c.R, c.G), c.B)
	maxc := max(max(c.R, c.G), c.B)
	if maxc > maxB {
		return c
	}
	n := m.rng.IntN(int(maxB-maxc)) - int(minc)
	return color.RGBA{
		uint8(int(c.R) + n),
		uint8(int(c.G) + n),
		uint8(int(c.B) + n),
		c.A,
	}
}
