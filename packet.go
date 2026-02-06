package goearth

import (
	"fmt"

	"xabbo.b7c.io/goearth/encoding"
	"xabbo.b7c.io/goearth/internal/debug"
)

var dbgPkt = debug.NewLoggerIf("[pkt]", debug.Ptrace)

type Packet struct {
	Client ClientType
	Header Header
	Data   []byte
	Pos    int
}

// [Parsable] represents an object that can be read from a [Packet].
type Parsable interface {
	Parse(r PacketReader)
}

// [Composable] represents an object that can be written to a [Packet].
type Composable interface {
	Compose(w PacketWriter)
}

// Represents a unique numeric identifier.
//
// Encoded as an int on [Flash] and [Shockwave] sessions, and a long on [Unity] sessions.
type Id int64

func (id *Id) Parse(r PacketReader) {
	switch r.Client() {
	case Flash, Shockwave:
		*id = Id(r.ReadInt())
	case Unity:
		*id = Id(r.ReadLong())
	default:
		panic(fmt.Errorf("attempt to read Id on unknown client: %s", r.Client()))
	}
}

func (id Id) Compose(w PacketWriter) {
	switch w.Client() {
	case Flash, Shockwave:
		w.WriteInt(int(id))
	case Unity:
		w.WriteLong(int64(id))
	default:
		panic("Cannot write ID: unknown client type.")
	}
}

// Repesents the length of an array or collection of items.
//
// Encoded as a short on [Shockwave] and [Unity], otherwise as an int.
type Length int32

func (length *Length) Parse(r PacketReader) {
	switch r.Client() {
	case Unity, Shockwave:
		*length = Length(r.ReadShort())
	default:
		*length = Length(r.ReadInt())
	}
}

func (length Length) Compose(w PacketWriter) {
	switch w.Client() {
	case Flash:
		w.WriteInt(int(length))
	case Unity, Shockwave:
		w.WriteShort(int16(length))
	default:
		panic("Cannot write Length: unknown client type.")
	}
}

// Represents a base-64 encoded integer, used in the [Shockwave] client.
type B64 int16

func (b64 *B64) Parse(r PacketReader) {
	assertCanRead(r, 2)
	if ir, ok := r.(packetReader); ok {
		*b64 = B64(encoding.B64Decode(ir.p.Data[*ir.pos : *ir.pos+2]))
		*ir.pos += 2
	} else {
		*b64 = B64(encoding.B64Decode(r.ReadBytes(2)))
	}
}

func (b64 B64) Compose(w PacketWriter) {
	encoding.B64Encode(w.Allocate(2), int(b64))
}

// Represents a variable-length base-64 encoded integer, used in the [Shockwave] client.
type VL64 int32

func (vl64 *VL64) Parse(r PacketReader) {
	n := encoding.VL64DecodeLen(r.Copy().ReadByte())
	if n <= 0 || n > 6 {
		panic(fmt.Errorf("invalid byte length when decoding VL64: %d", n))
	}
	*vl64 = VL64(encoding.VL64Decode(r.ReadBytes(n)))
}

func (vl64 VL64) Compose(w PacketWriter) {
	encoding.VL64Encode(w.Allocate(encoding.VL64EncodeLen(int(vl64))), int(vl64))
}

// Gets the length of the packet's data.
func (p *Packet) Length() int {
	return len(p.Data)
}

/* Reading */

// Returns a [PacketReader] referencing this packet and its position.
func (p *Packet) Reader() PacketReader {
	return packetReader{p, &p.Pos}
}

// Returns a [PacketReader] referencing this packet at the specified position.
func (p *Packet) ReaderAt(pos int) PacketReader {
	return packetReader{p, &pos}
}

// Reads into the specified byte slice from the current position and advances it.
func (p *Packet) ReadBuffer(buf []byte) {
	p.Reader().ReadBuffer(buf)
}

// Reads into the specified byte slice from the specified position and advances it.
func (p *Packet) ReadBufferAt(pos int, buf []byte) {
	p.ReaderAt(pos).ReadBuffer(buf)
}

// Reads a byte from the current position and advances it.
func (p *Packet) ReadByte() byte {
	return p.Reader().ReadByte()
}

// Reads a byte from the specified position.
func (p *Packet) ReadByteAt(pos int) byte {
	return p.ReaderAt(pos).ReadByte()
}

// Reads a bool from the current position and advances it.
//
// Read as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) ReadBool() bool {
	return p.Reader().ReadBool()
}

// Reads a bool from the specified position.
//
// Read as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) ReadBoolAt(pos int) bool {
	return p.ReaderAt(pos).ReadBool()
}

// Copies `n` bytes from the current position and advances it.
func (p *Packet) ReadBytes(n int) []byte {
	return p.Reader().ReadBytes(n)
}

// Copies `n` bytes from the specified position.
func (p *Packet) ReadBytesAt(pos int, n int) []byte {
	return p.ReaderAt(pos).ReadBytes(n)
}

// Reads a short from the current position and advances it.
//
// Read as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) ReadShort() int16 {
	return p.Reader().ReadShort()
}

// Reads a short from the specified position.
//
// Read as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) ReadShortAt(pos int) int16 {
	return p.ReaderAt(pos).ReadShort()
}

// Reads an int from the current position and advances it.
//
// Read as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) ReadInt() int {
	return p.Reader().ReadInt()
}

// Reads an int from the specified position.
//
// Read as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) ReadIntAt(pos int) int {
	return p.ReaderAt(pos).ReadInt()
}

// Reads a float from the current position and advances it.
//
// Read as a string and parsed to a float on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) ReadFloat() float32 {
	return p.Reader().ReadFloat()
}

// Reads a float from the specified position.
//
// Read as a string and parsed to a float on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) ReadFloatAt(pos int) float32 {
	return p.ReaderAt(pos).ReadFloat()
}

// Reads a long from the current position and advances it.
//
// Only supported on Unity sessions.
func (p *Packet) ReadLong() int64 {
	return p.Reader().ReadLong()
}

// Reads a long from the specified position.
//
// Only supported on Unity sessions.
func (p *Packet) ReadLongAt(pos int) int64 {
	return p.ReaderAt(pos).ReadLong()
}

// Reads a string from the current position and advances it.
//
// Read as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) ReadString() string {
	return p.Reader().ReadString()
}

// Reads a string from the specified position.
//
// Read as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) ReadStringAt(pos int) string {
	return p.ReaderAt(pos).ReadString()
}

// Reads into the specified variables from the current position and advances it.
// The provided variables must be a pointer type or implement [Parsable].
func (p *Packet) Read(vars ...any) {
	p.Reader().Read(vars...)
}

// Reads into the specified variables from the specified position.
// The provided variables must be a pointer type or implement [Parsable].
func (p *Packet) ReadAt(pos int, vars ...any) {
	p.ReaderAt(pos).Read(vars...)
}

/* Writing */

// Returns a [PacketWriter] referencing this packet and its position.
func (p *Packet) Writer() PacketWriter {
	return packetWriter{p, &p.Pos}
}

// Returns a [PacketWriter] referencing this packet at the specified position.
func (p *Packet) WriterAt(pos int) PacketWriter {
	return packetWriter{p, &pos}
}

// Writes a bool at the current position and advances it.
//
// Written as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) WriteBool(value bool) {
	p.Writer().WriteBool(value)
}

// Writes a bool at the specified position.
//
// Written as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) WriteBoolAt(pos int, value bool) {
	p.WriterAt(pos).WriteBool(value)
}

// Writes a byte at the current position and advances it.
func (p *Packet) WriteByte(value byte) {
	p.Writer().WriteByte(value)
}

// Writes a byte at the specified position.
func (p *Packet) WriteByteAt(pos int, value byte) {
	p.WriterAt(pos).WriteByte(value)
}

// Writes a slice of bytes at the current position and advances it.
func (p *Packet) WriteBytes(value []byte) {
	p.Writer().WriteBytes(value)
}

// Writes a slice of bytes at the specified position.
func (p *Packet) WriteBytesAt(pos int, value []byte) {
	p.WriterAt(pos).WriteBytes(value)
}

// Writes a short at the current position and advances it.
//
// Written as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) WriteShort(value int16) {
	p.Writer().WriteShort(value)
}

// Writes a short at the specified position.
//
// Written as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) WriteShortAt(pos int, value int16) {
	p.WriterAt(pos).WriteShort(value)
}

// Writes an int at the current position and advances it.
//
// Written as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) WriteInt(value int) {
	p.Writer().WriteInt(value)
}

// Writes an int at the specified position.
//
// Written as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) WriteIntAt(pos int, value int) {
	p.WriterAt(pos).WriteInt(value)
}

// Writes a float at the current position and advances it.
//
// Written as a string on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) WriteFloat(value float32) {
	p.Writer().WriteFloat(value)
}

// Writes a float at the specified position.
//
// Written as a string on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) WriteFloatAt(pos int, value float32) {
	p.WriterAt(pos).WriteFloat(value)
}

// Writes a long at the current position and advances it.
//
// Only supported on Unity sessions.
func (p *Packet) WriteLong(value int64) {
	p.Writer().WriteLong(value)
}

// Writes a long at the specified position.
//
// Only supported on Unity sessions.
func (p *Packet) WriteLongAt(pos int, value int64) {
	p.WriterAt(pos).WriteLong(value)
}

// Writes a string at the current position and advances it.
//
// Written as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) WriteString(value string) {
	p.Writer().WriteString(value)
}

// Writes a string at the specified position.
//
// Written as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) WriteStringAt(pos int, value string) {
	p.WriterAt(pos).WriteString(value)
}

// Writes the specified values at the current position and advances it.
func (p *Packet) Write(values ...any) {
	p.Writer().Write(values...)
}

// Writes the specified values at the specified position.
func (p *Packet) WriteAt(pos int, values ...any) {
	p.WriterAt(pos).Write(values...)
}

/* Modification */

// Modifies a string at the current position and advances it.
func (p *Packet) ModifyString(transform func(string) string) {
	p.Writer().ModifyString(transform)
}

// Modifies a string at the specified position.
func (p *Packet) ModifyStringAt(pos int, transform func(string) string) {
	p.WriterAt(pos).ModifyString(transform)
}

// Replaces a string at the current position and advances it.
func (p *Packet) ReplaceString(value string) {
	p.Writer().ReplaceString(value)
}

// Replaces a string at the specified position.
func (p *Packet) ReplaceStringAt(pos int, value string) {
	p.WriterAt(pos).ReplaceString(value)
}

/* Skipping */

// Skips the types specified by the provided values, advancing the current position.
func (p *Packet) Skip(values ...any) {
	p.Reader().Skip(values...)
}

/* Cloning */

// Makes a copy of the packet.
func (p *Packet) Copy() *Packet {
	data := make([]byte, len(p.Data))
	copy(data, p.Data)
	return &Packet{
		Client: p.Client,
		Header: p.Header,
		Data:   data,
	}
}
