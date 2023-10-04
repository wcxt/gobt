package message_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/edwces/gobt/wire/message"
)

func TestWriteMessage(t *testing.T) {
    tests := map[string]struct{
        input *message.Message
        want  []byte
        err     bool 
    }{
        "keep alive": {
            input: &message.Message{KeepAlive: true},
            want: []byte{0, 0, 0, 0},
            err: false,
        },
        "with id": {
            input: &message.Message{ID: message.IDUnchoke},
            want: []byte{0, 0, 0, 1, 1},
            err: false,
        },
        "with payload": {
            input: &message.Message{ID: message.IDBitfield, Payload: message.NewBitfieldPayload(message.Bitfield([]byte{10, 15, 5}))},
            want: []byte{0, 0, 0, 4, 5, 10, 15, 5},
            err: false,
        },
    }

     for name, test := range tests {
        t.Run(name, func(t *testing.T) {
            var buf bytes.Buffer
            _, err := message.Write(&buf, test.input)
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


func TestReadMessage(t *testing.T) {
    tests := map[string]struct{
        input []byte
        want  *message.Message
        err     bool 
    }{
        "keep alive": {
            input: []byte{0, 0, 0, 0},
            want: &message.Message{KeepAlive: true},
            err: false,
        },
        "with id": {
            input: []byte{0, 0, 0, 1, 1},
            want: &message.Message{ID: message.IDUnchoke, Payload: message.Payload([]byte{})},
            err: false,
        },
        "with payload": {
            input: []byte{0, 0, 0, 4, 5, 10, 15, 5},
            want: &message.Message{ID: message.IDBitfield, Payload: message.NewBitfieldPayload(message.Bitfield([]byte{10, 15, 5}))},
            err: false,
        },
        "zero bytes": {input: []byte{}, want: nil, err: true},
        "not enough bytes": {input: []byte{0, 0}, want: nil, err: true},
        "not enough length": {input: []byte{0, 0, 0, 20, 1, 0}, want: nil, err: true},
    }

    for name, test := range tests {
        t.Run(name, func(t *testing.T) {
            r := bytes.NewReader(test.input)
            msg, err := message.Read(r)

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
