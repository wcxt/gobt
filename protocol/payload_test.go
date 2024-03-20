package protocol_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/edwces/gobt/protocol"
)

func TestRequestMarshal(t *testing.T) {
	req := protocol.Request{Index: 1467, Offset: 64000, Length: 16000}

	got := req.Marshal()
	want := protocol.Payload{0x00, 0x00, 0x05, 0xBB,
		0x00, 0x00, 0xFA, 0x00,
		0x00, 0x00, 0x3E, 0x80}

	if !bytes.Equal(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestPayloadRequest(t *testing.T) {
	payload := protocol.Payload{0x00, 0x00, 0x05, 0xBB,
		0x00, 0x00, 0xFA, 0x00,
		0x00, 0x00, 0x3E, 0x80}

	got := payload.Request()
	want := protocol.Request{Index: 1467, Offset: 64000, Length: 16000}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestBlockMarshal(t *testing.T) {
	block := protocol.Block{Index: 1467, Offset: 64000, Block: []byte{0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}}

	got := block.Marshal()
	want := protocol.Payload{0x00, 0x00, 0x05, 0xBB,
		0x00, 0x00, 0xFA, 0x00,
		0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}

	if !bytes.Equal(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestPayloadBlock(t *testing.T) {
	payload := protocol.Payload{0x00, 0x00, 0x05, 0xBB,
		0x00, 0x00, 0xFA, 0x00,
		0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}

	got := payload.Block()
	want := protocol.Block{Index: 1467, Offset: 64000, Block: []byte{0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestHaveMarshal(t *testing.T) {
	have := protocol.Have(1467)

	got := have.Marshal()
	want := protocol.Payload{0x00, 0x00, 0x05, 0xBB}

	if !bytes.Equal(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestPayloadHave(t *testing.T) {
	payload := protocol.Payload{0x00, 0x00, 0x05, 0xBB}

	got := payload.Have()
	want := uint32(1467)

	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}
