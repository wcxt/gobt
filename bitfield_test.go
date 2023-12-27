package gobt_test

import (
	"testing"

	"github.com/edwces/gobt"
)

func TestBitfieldSet(t *testing.T) {
	tests := map[string]struct {
		input []byte
		index int
		want  bool
	}{
		"first bit": {
			input: []byte{0, 0},
			index: 0,
			want:  true,
		},
		"last bit": {
			input: []byte{0, 0},
			index: 15,
			want:  true,
		},
		"false value": {
			input: []byte{255, 255},
			index: 13,
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bitfield := gobt.Bitfield(test.input)
			bitfield.Set(test.index, test.want)

			got := bitfield.Get(test.index)

			if got != test.want {
				t.Fatalf("got %t, want %t", got, test.want)
			}
		})
	}
}

func TestBitfieldGet(t *testing.T) {
	tests := map[string]struct {
		input []byte
		index int
		want  bool
	}{
		"first bit": {
			input: []byte{128, 0},
			index: 0,
			want:  true,
		},
		"last bit": {
			input: []byte{0, 1},
			index: 15,
			want:  true,
		},
		"false value": {
			input: []byte{251, 255},
			index: 5,
			want:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bitfield := gobt.Bitfield(test.input)
			got := bitfield.Get(test.index)

			if got != test.want {
				t.Fatalf("got %t, want %t", got, test.want)
			}
		})
	}
}
