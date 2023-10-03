package handshake_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/edwces/gobt/wire/handshake"
)

func TestWriteHandshake(t *testing.T) {
	infoHash := [20]byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	peerId := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	hs := handshake.New(infoHash, peerId)

	want := []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116,
		111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 20, 19, 18, 17, 16, 15, 14, 13,
		12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20}

    var buf bytes.Buffer
    handshake.Write(&buf, hs)
    
    got := buf.Bytes()

	if !bytes.Equal(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestReadHandshake(t *testing.T) {
	tests := map[string]struct {
		input []byte
		want  *handshake.Handshake
		err   bool
	}{
        "handshake": {
            input: []byte{19, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116,
		            111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 20, 19, 18, 17, 16, 15, 14, 13,
		            12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		            13, 14, 15, 16, 17, 18, 19, 20},
            want: handshake.New(
                    [20]byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
                    [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
                ),
            err: false,
        },
        "zero bytes": {input: []byte{}, want: nil, err: true},
        "not enough bytes": {input: []byte{19, 66, 105, 116, 84, 111, 114}, want: nil, err: true},
        "pstrlen unexpected value": {
            input: []byte{16, 66, 105, 116, 84, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116,
		            111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 20, 19, 18, 17, 16, 15, 14, 13,
		            12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		            13, 14, 15, 16, 17, 18, 19, 20},
            want: nil,
            err: true,
        },
        "pstr unexpected value": {
            input: []byte{19, 56, 76, 16, 14, 111, 114, 114, 101, 110, 116, 32, 112, 114, 111, 116,
		            111, 99, 111, 108, 0, 0, 0, 0, 0, 0, 0, 0, 20, 19, 18, 17, 16, 15, 14, 13,
		            12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		            13, 14, 15, 16, 17, 18, 19, 20},
            want: nil,
            err: true,
        },
    }

    for name, test := range tests {
        t.Run(name, func(t *testing.T) {
            r := bytes.NewReader(test.input)
            hs, err := handshake.Read(r)
            if !test.err && err != nil {
                t.Fatalf("got error: %s, want nil", err.Error())
            }
            if test.err && err == nil {
                t.Fatalf("got nil, want err")
            }
            if !reflect.DeepEqual(hs, test.want) {
                t.Fatalf("got %#v, want %#v", hs, test.want)
            }
        })
    }
}
