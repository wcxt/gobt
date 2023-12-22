package message_test

import (
	"bytes"
	"testing"

	"github.com/edwces/gobt/message"
)

func TestNewBitfieldPayload(t *testing.T) {
    bitfield := make(message.Bitfield, 2)

    bitfield.Set(0)
    bitfield.Set(15)

    bitfield.Set(4)
    bitfield.Set(13)
    bitfield.Set(8)

    got := message.NewBitfieldPayload(bitfield)
    want := []byte{136, 133}
    
    if !bytes.Equal(got, want) {
        t.Fatalf("got %#v, want %#v", got, want)
    }
}

func TestBitfieldGet(t *testing.T) {
    tests := map[string]struct{
        input message.Bitfield
        index int
        want bool
    }{
        "first bit": {
            input: []byte{128, 0},
            index: 0,
            want: true,
        },

        "last bit": {
            input: []byte{0, 1},
            index: 15,
            want: true,
        },

        "correct indexing": {
            input: []byte{0, 32},
            index: 10,
            want: true,
        },
        
        "false value": {
            input: []byte{251, 255},
            index: 5,
            want: false,
        },
    }

    for name, test := range tests {
        t.Run(name, func(t *testing.T) {
            bitfield := message.Bitfield(test.input)
            got := bitfield.Get(test.index)

            if got != test.want {
                t.Fatalf("got %t, want %t", got, test.want)
            }
        })
    }


}


