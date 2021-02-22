package internal

import "bytes"

const encoding = "0123456789abcdefghijklmnopqrstuv"

func encodeXIDToBuf(id []byte, dst *bytes.Buffer) {
	dst.WriteByte(encoding[id[0]>>3])
	dst.WriteByte(encoding[(id[1]>>6)&0x1F|(id[0]<<2)&0x1F])
	dst.WriteByte(encoding[(id[1]>>1)&0x1F])
	dst.WriteByte(encoding[(id[2]>>4)&0x1F|(id[1]<<4)&0x1F])
	dst.WriteByte(encoding[id[3]>>7|(id[2]<<1)&0x1F])
	dst.WriteByte(encoding[(id[3]>>2)&0x1F])
	dst.WriteByte(encoding[id[4]>>5|(id[3]<<3)&0x1F])
	dst.WriteByte(encoding[id[4]&0x1F])
	dst.WriteByte(encoding[id[5]>>3])
	dst.WriteByte(encoding[(id[6]>>6)&0x1F|(id[5]<<2)&0x1F])
	dst.WriteByte(encoding[(id[6]>>1)&0x1F])
	dst.WriteByte(encoding[(id[7]>>4)&0x1F|(id[6]<<4)&0x1F])
	dst.WriteByte(encoding[id[8]>>7|(id[7]<<1)&0x1F])
	dst.WriteByte(encoding[(id[8]>>2)&0x1F])
	dst.WriteByte(encoding[(id[9]>>5)|(id[8]<<3)&0x1F])
	dst.WriteByte(encoding[id[9]&0x1F])
	dst.WriteByte(encoding[id[10]>>3])
	dst.WriteByte(encoding[(id[11]>>6)&0x1F|(id[10]<<2)&0x1F])
	dst.WriteByte(encoding[(id[11]>>1)&0x1F])
	dst.WriteByte(encoding[(id[11]<<4)&0x1F])
}
