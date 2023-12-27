package gobt

type Bitfield []byte

func (bf Bitfield) Set(i int, val bool) {
	byteI := i / 8
	bitI := i % 8

    if val {
	    bf[byteI] |= 0b10000000 >> bitI
    } else {
        bf[byteI] &= 0b01111111 >> bitI
    }
}

func (bf Bitfield) Get(i int) bool {
	byteI := i / 8
	bitI := i % 8
	val := int(bf[byteI] & (0b10000000 >> bitI))
	return val != 0
}
