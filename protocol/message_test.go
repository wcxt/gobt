package protocol_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/edwces/gobt/protocol"
)

func TestMarshalMessage(t *testing.T) {
	tests := map[string]struct {
		input *protocol.Message
		want  []byte
		err   bool
	}{
		"keep alive": {
			input: &protocol.Message{KeepAlive: true},
			want:  []byte{0, 0, 0, 0},
			err:   false,
		},
		"with id": {
			input: &protocol.Message{ID: protocol.IDUnchoke},
			want:  []byte{0, 0, 0, 1, 1},
			err:   false,
		},
		"with payload": {
			input: &protocol.Message{ID: protocol.IDBitfield, Payload: []byte{10, 15, 5}},
			want:  []byte{0, 0, 0, 4, 5, 10, 15, 5},
			err:   false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			_, err := buf.Write(test.input.Marshal())
			got := buf.Bytes()

			if !test.err && err != nil {
				t.Fatalf("got error: %s, want nil", err.Error())
			}
			if test.err && err == nil {
				t.Fatalf("got nil, want err")
			}
			if !bytes.Equal(got, test.want) {
				t.Fatalf("got %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestUnmarshalMessage(t *testing.T) {
	tests := map[string]struct {
		input []byte
		want  *protocol.Message
		err   bool
	}{
		"keep alive": {
			input: []byte{0, 0, 0, 0},
			want:  &protocol.Message{KeepAlive: true},
			err:   false,
		},
		"with id": {
			input: []byte{0, 0, 0, 1, 1},
			want:  &protocol.Message{ID: protocol.IDUnchoke, Payload: protocol.Payload([]byte{})},
			err:   false,
		},
		"with payload": {
			input: []byte{0, 0, 0, 4, 5, 10, 15, 5},
			want:  &protocol.Message{ID: protocol.IDBitfield, Payload: []byte{10, 15, 5}},
			err:   false,
		},
		"zero bytes":        {input: []byte{}, want: nil, err: true},
		"not enough bytes":  {input: []byte{0, 0}, want: nil, err: true},
		"not enough length": {input: []byte{0, 0, 0, 20, 1, 0}, want: nil, err: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := bytes.NewReader(test.input)
			msg, err := protocol.UnmarshalMessage(r)

			if !test.err && err != nil {
				t.Fatalf("got error: %s, want nil", err.Error())
			}
			if test.err && err == nil {
				t.Fatalf("got nil, want err")
			}
			if !reflect.DeepEqual(msg, test.want) {
				t.Fatalf("got %#v, want %#v", msg, test.want)
			}
		})
	}

}
