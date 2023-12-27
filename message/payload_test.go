package message_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/edwces/gobt/message"
)

func TestNewBitfieldPayload(t *testing.T) {
    bitfield := message.Bitfield{136, 133}

    got := message.NewBitfieldPayload(bitfield)
    want := message.Payload{136, 133}
    
    if !bytes.Equal(got, want) {
        t.Fatalf("got %#v, want %#v", got, want)
    }
}

func TestNewRequestPayload(t *testing.T) {
    req := message.Request{Index: 1467, Offset: 64000, Length: 16000}

    got := message.NewRequestPayload(req)
    want := message.Payload{0x00, 0x00, 0x05, 0xBB,
                            0x00, 0x00, 0xFA, 0x00,
                            0x00, 0x00, 0x3E, 0x80}

    if !bytes.Equal(got, want) {
        t.Fatalf("got %#v, want %#v", got, want)
    }
}

func TestPayloadRequest(t *testing.T) {
    payload := message.Payload{0x00, 0x00, 0x05, 0xBB,
                               0x00, 0x00, 0xFA, 0x00,
                               0x00, 0x00, 0x3E, 0x80}
    
    got := payload.Request()
    want := message.Request{Index: 1467, Offset: 64000, Length: 16000}

    if !reflect.DeepEqual(got, want) {
        t.Fatalf("got %#v, want %#v", got, want)
    }
}

func TestNewBlockPayload(t *testing.T) {
    block := message.Block{Index: 1467, Offset: 64000, Block: []byte{0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}}

    got := message.NewBlockPayload(block)
    want := message.Payload{0x00, 0x00, 0x05, 0xBB,
                            0x00, 0x00, 0xFA, 0x00,
                            0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}

    if !bytes.Equal(got, want) {
        t.Fatalf("got %#v, want %#v", got, want)
    }
}

func TestPayloadBlock(t *testing.T) {
    payload := message.Payload{0x00, 0x00, 0x05, 0xBB,
                               0x00, 0x00, 0xFA, 0x00,
                               0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}
    
    got := payload.Block()
    want := message.Block{Index: 1467, Offset: 64000, Block: []byte{0x00, 0xE2, 0x06, 0xAB, 0x00, 0x31}}

    if !reflect.DeepEqual(got, want) {
        t.Fatalf("got %#v, want %#v", got, want)
    }
}

func TestNewHavePayload(t *testing.T) {
    have := uint32(1467)

    got := message.NewHavePayload(have)
    want := message.Payload{0x00, 0x00, 0x05, 0xBB}

    if !bytes.Equal(got, want) {
        t.Fatalf("got %#v, want %#v", got, want)
    }
}

func TestPayloadHave(t *testing.T) {
    payload := message.Payload{0x00, 0x00, 0x05, 0xBB}

    got := payload.Have()
    want := uint32(1467)

    if got != want {
        t.Fatalf("got %d, want %d", got, want)
    }
}
