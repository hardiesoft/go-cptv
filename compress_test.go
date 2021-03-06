// Copyright 2018 The Cacophony Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cptv

import (
	"bytes"
	"testing"

	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestCamera struct {
}

func (cam *TestCamera) ResX() int {
	return 160
}
func (cam *TestCamera) ResY() int {
	return 120
}
func (cam *TestCamera) FPS() int {
	return 21
}
func TestCompressDecompress(t *testing.T) {
	camera := new(TestCamera)
	frame0 := makeTestFrame(camera)
	frame1 := makeOffsetFrame(camera, frame0)

	// Compress the frames.
	compressor := NewCompressor(camera)
	bitWidth0, frameComp := compressor.Next(frame0)
	// first frame has no compression
	assert.Equal(t, uint8(14), bitWidth0)
	assert.Equal(t, 33603, len(frameComp))
	frame0Comp := make([]byte, len(frameComp))
	copy(frame0Comp, frameComp)

	bitWidth1, frame1Comp := compressor.Next(frame1)
	assert.Equal(t, uint8(2), bitWidth1)
	assert.Equal(t, 4804, len(frame1Comp))

	// Decompress the frames and confirm the output is the same as the original.
	decompressor := NewDecompressor(camera)

	frame0d := cptvframe.NewFrame(camera)
	err := decompressor.Next(bitWidth0, bytes.NewReader(frame0Comp), frame0d)
	require.NoError(t, err)
	assert.Equal(t, frame0, frame0d)

	frame1d := cptvframe.NewFrame(camera)
	err = decompressor.Next(bitWidth1, bytes.NewReader(frame1Comp), frame1d)
	require.NoError(t, err)
	assert.Equal(t, frame1, frame1d)
}

func TestTwosComp(t *testing.T) {
	tests := []struct {
		input    int32
		width    uint8
		expected uint32
	}{

		{-1, 4, 15},

		// Width 8
		{0, 8, 0},
		{1, 8, 1},
		{-1, 8, 255},
		{15, 8, 15},
		{-15, 8, 241},
		{127, 8, 127},
		{-127, 8, 129},
		{-128, 8, 128},

		{-12, 9, 500},

		// Width 5
		{0, 5, 0},
		{1, 5, 1},
		{-1, 5, 31},
		{15, 5, 15},
		{-15, 5, 17},
		{-16, 5, 16},

		// Width 14
		{0, 14, 0},
		{1, 14, 1},
		{-1, 14, 16383},
		{15, 14, 15},
		{-15, 14, 16369},
		{8191, 14, 8191},
		{-8192, 14, 8192},
	}

	for _, x := range tests {
		twos := twosComp(x.input, x.width)
		assert.Equal(t, x.expected, twos, "twosComp(%d, %d)", x.input, x.width)

		untwos := twosUncomp(twos, x.width)
		assert.Equal(t, x.input, untwos, "twosUncomp(%d, %d)", twos, x.width)
	}
}

func makeTestFrame(c cptvframe.CameraSpec) *cptvframe.Frame {
	// Generate a frame with values between 1024 and 8196
	out := cptvframe.NewFrame(c)
	const minVal = 1024
	const maxVal = 8196
	for y, row := range out.Pix {
		for x, _ := range row {
			out.Pix[y][x] = uint16(((y * x) % (maxVal - minVal)) + minVal)
		}
	}

	return out
}

func makeOffsetFrame(c cptvframe.CameraSpec, in *cptvframe.Frame) *cptvframe.Frame {
	out := cptvframe.NewFrame(c)

	for y, row := range out.Pix {
		for x, _ := range row {
			out.Pix[y][x] = in.Pix[y][x] + uint16(x)
		}
	}
	return out
}
