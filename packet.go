package goearth

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"unsafe"

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

// Parsable represents an object that can be read from a Packet.
type Parsable interface {
	Parse(pkt *Packet, pos *int)
}

// Composable represents an object that can be written to a Packet.
type Composable interface {
	Compose(pkt *Packet, pos *int)
}

// Represents a unique numeric identifier.
//
// Encoded as an int on Flash and Shockwave sessions, and a long on Unity sessions.
type Id int64

func (id *Id) Parse(p *Packet, pos *int) {
	switch p.Client {
	case Flash, Shockwave:
		*id = Id(p.ReadIntPtr(pos))
	case Unity:
		*id = Id(p.ReadLongPtr(pos))
	default:
		panic(fmt.Errorf("attempt to read Id on unknown client: %s", p.Client))
	}
}

func (id Id) Compose(p *Packet, pos *int) {
	switch p.Client {
	case Flash, Shockwave:
		p.WriteIntPtr(pos, int(id))
	case Unity:
		p.WriteLongPtr(pos, int64(id))
	default:
		panic("Cannot write ID: unknown client type.")
	}
}

// Repesents the length of an array or collection of items.
//
// Encoded as a short on Shockwave and Unity, otherwise as an int.
type Length int32

func (length *Length) Parse(p *Packet, pos *int) {
	switch p.Client {
	case Unity, Shockwave:
		*length = Length(p.ReadShortPtr(pos))
	default:
		*length = Length(p.ReadIntPtr(pos))
	}

}

func (length Length) Compose(p *Packet, pos *int) {
	switch p.Client {
	case Flash:
		p.WriteIntPtr(pos, int(length))
	case Unity, Shockwave:
		p.WriteShortPtr(pos, int16(length))
	default:
		panic("Cannot write Length: unknown client type.")
	}
}

// Represents a base-64 encoded integer, used in the Shockwave client.
type B64 int16

func (b64 *B64) Parse(p *Packet, pos *int) {
	p.assertCanRead(*pos, 2)
	*b64 = B64(encoding.B64Decode(p.Data[*pos : *pos+2]))
	*pos += 2
}

func (b64 B64) Compose(p *Packet, pos *int) {
	p.ensureLength(*pos, 2)
	encoding.B64Encode(p.Data[*pos:*pos+2], int(b64))
	*pos += 2
}

// Represents a variable-length base-64 encoded integer, used in the Shockwave client.
type VL64 int32

func (vl64 *VL64) Parse(p *Packet, pos *int) {
	p.assertCanRead(*pos, 1)
	n := encoding.VL64DecodeLen(p.Data[*pos])
	if n <= 0 || n > 6 {
		panic(fmt.Errorf("invalid byte length when decoding VL64: %d", n))
	}
	p.assertCanRead(*pos, n)
	*vl64 = VL64(encoding.VL64Decode(p.Data[*pos : *pos+n]))
	*pos += n
}

func (vl64 VL64) Compose(p *Packet, pos *int) {
	n := encoding.VL64EncodeLen(int(vl64))
	p.ensureLength(*pos, n)
	encoding.VL64Encode(p.Data[*pos:*pos+n], int(vl64))
	*pos += n
}

// Ensures the packet has enough capacity to store `n` bytes from the specified position.
func (p *Packet) ensureLength(pos int, n int) {
	if pos < 0 {
		panic("position cannot be < 0.")
	}
	if pos > len(p.Data) {
		panic("position cannot be > packet length.")
	}
	if length := (pos + n); length > len(p.Data) {
		if cap(p.Data) > len(p.Data) {
			p.Data = p.Data[:min(cap(p.Data), length)]
		}
		if length > len(p.Data) {
			p.Data = append(p.Data, make([]byte, length-len(p.Data))...)
		}
	}
}

// Ensures `n` bytes can be read from the specified position.
func (p *Packet) assertCanRead(pos int, n int) {
	if pos < 0 {
		panic(fmt.Errorf("packet position cannot be < 0"))
	}
	if (pos + n) > len(p.Data) {
		panic(fmt.Errorf("attempt to read past packet length"))
	}
}

// Gets the length of the packet's data.
func (p *Packet) Length() int {
	return len(p.Data)
}

/* Reading */

// Reads into the specified byte slice from the specified position and advances it.
func (p *Packet) ReadBufferPtr(pos *int, buf []byte) {
	length := len(buf)
	p.assertCanRead(*pos, length)
	copy(buf, p.Data[*pos:])
	*pos += length
}

// Reads into the specified byte slice from the specified position.
func (p *Packet) ReadBufferAt(pos int, buf []byte) {
	p.ReadBufferPtr(&pos, buf)
}

// Reads into the specified byte slice from the current position.
func (p *Packet) ReadBuffer(buf []byte) {
	p.ReadBufferPtr(&p.Pos, buf)
}

// Reads a byte from the specified position and advances it.
func (p *Packet) ReadBytePtr(pos *int) (value byte) {
	p.assertCanRead(*pos, 1)
	value = p.Data[*pos]
	dbgPkt.Printf("%d", value)
	*pos++
	return
}

// Reads a byte from the specified position.
func (p *Packet) ReadByteAt(pos int) byte {
	return p.ReadBytePtr(&pos)
}

// Reads a byte from the current position.
func (p *Packet) ReadByte() byte {
	return p.ReadBytePtr(&p.Pos)
}

// Reads a bool from the specified position.
//
// Read as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) ReadBoolPtr(pos *int) (value bool) {
	p.assertCanRead(*pos, 1)
	var i int
	switch p.Client {
	case Shockwave:
		if encoding.VL64DecodeLen(p.Data[*pos]) != 1 {
			panic(fmt.Errorf("attempt to read boolean when VL64 length > 1"))
		}
		i = encoding.VL64Decode(p.Data[*pos : *pos+1])
		dbgPkt.Printf("vl64: %t", i == 1)
	default:
		i = int(p.Data[*pos])
		dbgPkt.Printf("%t", i == 1)
	}
	if i != 0 && i != 1 {
		panic(fmt.Errorf("attempt to read boolean from non-boolean value: %d", i))
	}
	value = i == 1
	*pos++
	return
}

// Reads a bool at the specified position.
//
// Read as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) ReadBoolAt(pos int) bool {
	return p.ReadBoolPtr(&pos)
}

// Reads a bool from the current position.
//
// Read as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) ReadBool() bool {
	return p.ReadBoolPtr(&p.Pos)
}

// Copies `n` bytes from the specified position and advances it.
func (p *Packet) ReadBytesPtr(pos *int, n int) (value []byte) {
	p.assertCanRead(*pos, n)
	value = make([]byte, n)
	dbgPkt.Printf("%v", value)
	*pos += copy(value, p.Data[*pos:])
	return
}

// Copies `n` bytes from the specified position.
func (p *Packet) ReadBytesAt(pos int, length int) []byte {
	return p.ReadBytesPtr(&pos, length)
}

// Copies `n` bytes from the current position.
func (p *Packet) ReadBytes(length int) []byte {
	return p.ReadBytesPtr(&p.Pos, length)
}

// Reads a short from the specified position and advances it.
//
// Read as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) ReadShortPtr(pos *int) (value int16) {
	if p.Client == Shockwave {
		switch p.Header.Dir {
		case In:
			var vl64 VL64
			vl64.Parse(p, pos)
			dbgPkt.Printf("vl64: %d", vl64)
			return int16(vl64)
		case Out:
			var b64 B64
			b64.Parse(p, pos)
			dbgPkt.Printf("b64: %d", b64)
			return int16(b64)
		default:
			panic(fmt.Errorf("%w: unknown packet direction when reading short on shockwave session",
				errors.ErrUnsupported))
		}
	}
	p.assertCanRead(*pos, 2)
	value = int16(binary.BigEndian.Uint16(p.Data[*pos:]))
	dbgPkt.Printf("%d", value)
	*pos += 2
	return
}

// Reads a short at the specified position.
//
// Read as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) ReadShortAt(pos int) int16 {
	return p.ReadShortPtr(&pos)
}

// Reads a short from the current position.
//
// Read as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) ReadShort() int16 {
	return p.ReadShortPtr(&p.Pos)
}

// Reads an int from the specified position.
//
// Read as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) ReadIntPtr(pos *int) (value int) {
	if p.Client == Shockwave {
		var vl64 VL64
		vl64.Parse(p, pos)
		dbgPkt.Printf("vl64: %d", vl64)
		return int(vl64)
	}
	p.assertCanRead(*pos, 4)
	value = int(int32(binary.BigEndian.Uint32(p.Data[*pos:])))
	dbgPkt.Printf("%d", value)
	*pos += 4
	return
}

// Reads an int at the specified position.
//
// Read as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) ReadIntAt(pos int) int {
	return p.ReadIntPtr(&pos)
}

// Reads an int from the current position.
//
// Read as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) ReadInt() int {
	return p.ReadIntPtr(&p.Pos)
}

// Reads a float from the specified position.
//
// Read as a string and parsed to a float on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) ReadFloatPtr(pos *int) float32 {
	switch p.Client {
	case Flash, Shockwave:
		s := p.ReadStringPtr(pos)
		value, err := strconv.ParseFloat(s, 32)
		if err != nil {
			panic(fmt.Errorf("failed to parse float: %w", err))
		}
		dbgPkt.Printf("string: %f", value)
		return float32(value)
	default:
		p.assertCanRead(*pos, 4)
		bits := binary.BigEndian.Uint32(p.Data[*pos:])
		value := math.Float32frombits(bits)
		dbgPkt.Printf("%f", value)
		*pos += 4
		return value
	}
}

// Reads a float at the specified position.
//
// Read as a string and parsed to a float on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) ReadFloatAt(pos int) float32 {
	return p.ReadFloatPtr(&pos)
}

// Reads a float from the current position.
//
// Read as a string and parsed to a float on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) ReadFloat() float32 {
	return p.ReadFloatPtr(&p.Pos)
}

// Reads a long from the specified position and advances it.
//
// Only supported on Unity sessions.
func (p *Packet) ReadLongPtr(pos *int) (value int64) {
	if p.Client != Unity {
		panic(fmt.Errorf("%w: attempt to read long on client: %s", errors.ErrUnsupported, p.Client))
	}
	p.assertCanRead(*pos, 8)
	x := binary.BigEndian.Uint64(p.Data[*pos:])
	ptr := unsafe.Pointer(&x)
	value = *(*int64)(ptr)
	dbgPkt.Printf("%d", value)
	*pos += 8
	return
}

// Reads a long at the specified position.
//
// Only supported on Unity sessions.
func (p *Packet) ReadLongAt(pos int) int64 {
	return p.ReadLongPtr(&pos)
}

// Reads a long from the current position.
//
// Only supported on Unity sessions.
func (p *Packet) ReadLong() int64 {
	return p.ReadLongPtr(&p.Pos)
}

// Reads a string from the specified position and advances it.
//
// Read as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) ReadStringPtr(pos *int) (value string) {
	if p.Client == Shockwave && p.Header.Dir == In {
		i := *pos
		for i < len(p.Data) && p.Data[i] != 2 {
			i++
		}
		// some packets don't have a terminator byte for the final string..
		// if i >= len(p.Data) {
		// 	panic(fmt.Errorf("unterminated string"))
		// }
		value = string(p.Data[*pos:i])
		*pos = min(len(p.Data), i+1)
	} else {
		length := int(uint16(p.ReadShortAt(*pos)))
		p.assertCanRead(*pos, 2+length)
		value = string(p.Data[*pos+2 : *pos+2+length])
		*pos += 2 + length
	}
	dbgPkt.Printf("%q", value)
	return
}

// Reads a string at the specified position.
//
// Read as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) ReadStringAt(pos int) string {
	return p.ReadStringPtr(&pos)
}

// Reads a string from the current position.
//
// Read as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) ReadString() string {
	return p.ReadStringPtr(&p.Pos)
}

// Reads into the specified variables from the specified position and advances it.
// The provided variables must be a pointer type or implement [Parsable].
func (p *Packet) ReadPtr(pos *int, vars ...any) {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("packet read failed: %v", r))
		}
	}()
	for _, v := range vars {
		if !p.readInterfacePtr(pos, v) {
			p.readReflectPtr(pos, reflect.ValueOf(v))
		}
	}
}

// Reads into the specified variables at the specified position.
// The provided variables must be a pointer type or implement [Parsable].
func (p *Packet) ReadAt(pos int, vars ...any) {
	p.ReadPtr(&pos, vars...)
}

// Reads into the specified variables from the current position.
// The provided variables must be a pointer type or implement [Parsable].
func (p *Packet) Read(vars ...any) {
	p.ReadPtr(&p.Pos, vars...)
}

func (p *Packet) readReflectPtr(pos *int, v reflect.Value) {
	if v.CanAddr() {
		if parsable, ok := v.Addr().Interface().(Parsable); ok {
			dbgPkt.Printf("Parsable: %s", reflect.TypeOf(v).Name())
			parsable.Parse(p, pos)
			return
		}
	}
	switch v.Kind() {
	case reflect.Pointer:
		p.readReflectPtr(pos, v.Elem())
	case reflect.Array:
		n := v.Len()
		dbgPkt.Printf("array[%d]", n)
		for i := 0; i < n; i++ {
			p.readReflectPtr(&p.Pos, v.Index(i))
		}
	case reflect.Slice:
		t := v.Type()
		var len Length
		len.Parse(p, pos)
		dbgPkt.Printf("slice[%d]", len)
		slc := reflect.MakeSlice(t, int(len), int(len))
		for i := 0; i < int(len); i++ {
			p.readReflectPtr(pos, slc.Index(i))
		}
		v.Set(slc)
	case reflect.Struct:
		n := v.NumField()
		dbgPkt.Printf("struct: %s", v.Type().Name())
		for i := 0; i < n; i++ {
			if v := v.Field(i); v.CanSet() {
				p.readReflectPtr(pos, v)
			}
		}
	case reflect.Interface:
		if p.readInterfacePtr(pos, v.Interface()) {
			return
		}
	default:
		if v.CanAddr() && v.CanInterface() {
			if p.readInterfacePtr(pos, v.Addr().Interface()) {
				return
			}
		}
		panic(fmt.Errorf("cannot read unsupported type: %+v", v.Type()))
	}
}

func (p *Packet) readInterfacePtr(pos *int, v any) bool {
	switch v := v.(type) {
	case Parsable:
		dbgPkt.Printf("Parsable: %s", reflect.TypeOf(v).Name())
		v.Parse(p, pos)
	case *bool:
		*v = p.ReadBoolPtr(pos)
	case *byte:
		*v = p.ReadBytePtr(pos)
	case *int16:
		*v = p.ReadShortPtr(pos)
	case *uint16:
		*v = uint16(p.ReadShortPtr(pos))
	case *int:
		*v = p.ReadIntPtr(pos)
	case *uint:
		*v = uint(p.ReadIntPtr(pos))
	case *int32:
		*v = int32(p.ReadIntPtr(pos))
	case *uint32:
		*v = uint32(p.ReadIntPtr(pos))
	case *float32:
		*v = p.ReadFloatPtr(pos)
	case *float64:
		*v = float64(p.ReadFloatPtr(pos))
	case *int64:
		*v = p.ReadLongPtr(pos)
	case *uint64:
		*v = uint64(p.ReadLongPtr(pos))
	case *string:
		*v = p.ReadStringPtr(pos)
	default:
		return false
	}
	return true
}

/* Writing */

// Writes a bool at the specified position and advances it.
//
// Written as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) WriteBoolPtr(pos *int, value bool) *Packet {
	p.ensureLength(*pos, 1)
	b := uint8(0)
	if value {
		b = 1
	}
	if p.Client == Shockwave {
		VL64(b).Compose(p, pos)
	} else {
		p.Data[*pos] = b
	}
	*pos++
	return p
}

// Writes a bool at the specified position.
//
// Written as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) WriteBoolAt(pos int, value bool) *Packet {
	return p.WriteBoolPtr(&pos, value)
}

// Writes a bool at the current position.
//
// Written as a VL64 on Shockwave, otherwise as a byte.
func (p *Packet) WriteBool(value bool) *Packet {
	return p.WriteBoolPtr(&p.Pos, value)
}

// Writes a byte at the specified position and advances it.
func (p *Packet) WriteBytePtr(pos *int, value byte) *Packet {
	p.ensureLength(*pos, 1)
	p.Data[*pos] = value
	*pos++
	return p
}

// Writes a byte at the specified position.
func (p *Packet) WriteByteAt(pos int, value byte) *Packet {
	return p.WriteBytePtr(&pos, value)
}

// Writes a byte at the current position.
func (p *Packet) WriteByte(value byte) *Packet {
	return p.WriteBytePtr(&p.Pos, value)
}

// Writes a slice of bytes at the specified position and advances it.
func (p *Packet) WriteBytesPtr(pos *int, value []byte) *Packet {
	length := len(value)
	p.ensureLength(*pos, length)
	copy(p.Data[*pos:], value)
	*pos += length
	return p
}

// Writes a slice of bytes at the specified position.
func (p *Packet) WriteBytesAt(pos int, value []byte) *Packet {
	return p.WriteBytesPtr(&pos, value)
}

// Writes a slice of bytes at the current position.
func (p *Packet) WriteBytes(value []byte) *Packet {
	return p.WriteBytesPtr(&p.Pos, value)
}

// Writes a short at the specified position and advances it.
//
// Written as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) WriteShortPtr(pos *int, value int16) *Packet {
	if p.Client == Shockwave {
		switch p.Header.Dir {
		case In:
			VL64(value).Compose(p, pos)
		case Out:
			B64(value).Compose(p, pos)
		default:
			panic(fmt.Errorf("%w: unknown packet direction when writing short on shockwave session",
				errors.ErrUnsupported))
		}
		return p
	}
	p.ensureLength(*pos, 2)
	binary.BigEndian.PutUint16(p.Data[*pos:], uint16(value))
	*pos += 2
	return p
}

// Writes a short at the specified position.
//
// Written as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) WriteShortAt(pos int, value int16) *Packet {
	return p.WriteShortPtr(&pos, value)
}

// Writes a short at the current position.
//
// Written as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (p *Packet) WriteShort(value int16) *Packet {
	return p.WriteShortPtr(&p.Pos, value)
}

// Writes an int at the specified position and advances it.
//
// Written as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) WriteIntPtr(pos *int, value int) *Packet {
	if p.Client == Shockwave {
		VL64(value).Compose(p, pos)
		return p
	}
	p.ensureLength(*pos, 4)
	binary.BigEndian.PutUint32(p.Data[*pos:], uint32(value))
	*pos += 4
	return p
}

// Writes an int at the specified position.
//
// Written as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) WriteIntAt(pos, value int) *Packet {
	return p.WriteIntPtr(&pos, value)
}

// Writes an int at the current position.
//
// Written as a VL64 on Shockwave, otherwise as an int32.
func (p *Packet) WriteInt(value int) *Packet {
	return p.WriteIntPtr(&p.Pos, value)
}

// Writes a float at the specified position and advances it.
//
// Written as a string on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) WriteFloatPtr(pos *int, value float32) *Packet {
	switch p.Client {
	case Flash, Shockwave:
		p.WriteStringPtr(pos, strconv.FormatFloat(float64(value), 'f', -1, 32))
	case Unity:
		p.ensureLength(*pos, 4)
		bits := math.Float32bits(value)
		binary.BigEndian.PutUint32(p.Data[*pos:], bits)
		*pos += 4
	default:
		panic(fmt.Errorf("attempt to write float on unknown client: %s", p.Client))
	}
	return p
}

// Writes a float at the specified position.
//
// Written as a string on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) WriteFloatAt(pos int, value float32) *Packet {
	return p.WriteFloatPtr(&pos, value)
}

// Writes a float at the current position.
//
// Written as a string on Flash and Shockwave sessions, otherwise as a float32.
func (p *Packet) WriteFloat(value float32) *Packet {
	return p.WriteFloatPtr(&p.Pos, value)
}

// Writes a long at the specified position and advances it.
//
// Only supported on Unity sessions.
func (p *Packet) WriteLongPtr(pos *int, value int64) *Packet {
	if p.Client == Shockwave {
		panic(fmt.Errorf("%w: attempt to write long on client: %s", errors.ErrUnsupported, p.Client))
	}
	p.ensureLength(*pos, 8)
	binary.BigEndian.PutUint64(p.Data[*pos:], uint64(value))
	*pos += 8
	return p
}

// Writes a long at the specified position.
//
// Only supported on Unity sessions.
func (p *Packet) WriteLongAt(pos int, value int64) *Packet {
	return p.WriteLongPtr(&pos, value)
}

// Writes a long at the current position.
//
// Only supported on Unity sessions.
func (p *Packet) WriteLong(value int64) *Packet {
	return p.WriteLongPtr(&p.Pos, value)
}

// Writes a string at the specified position and advances it.
//
// Written as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) WriteStringPtr(pos *int, value string) *Packet {
	b := []byte(value)
	size := len(b)

	if p.Client == Shockwave && p.Header.Dir == In {
		p.ensureLength(*pos, 1+size)
		copy(p.Data[*pos:], b)
		p.Data[*pos+size] = 2
		*pos += size + 1
		return p
	}

	if size >= (1 << 16) {
		panic(fmt.Errorf("string length cannot fit into a uint16"))
	}

	p.ensureLength(*pos, 2+size)
	p.WriteShortAt(*pos, int16(size))
	copy(p.Data[*pos+2:], b)
	*pos += (2 + size)
	return p
}

// Writes a string at the specified position.
//
// Written as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) WriteStringAt(pos int, value string) *Packet {
	return p.WriteStringPtr(&pos, value)
}

// Writes a string at the current position.
//
// Written as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (p *Packet) WriteString(value string) *Packet {
	return p.WriteStringPtr(&p.Pos, value)
}

// Writes the specified values at the specified position and advances it.
func (p *Packet) WritePtr(pos *int, values ...any) *Packet {
	for _, value := range values {
		switch v := value.(type) {
		case Composable:
			v.Compose(p, pos)
		case bool:
			p.WriteBoolPtr(pos, v)
		case int8:
			p.WriteBytePtr(pos, byte(v))
		case uint8:
			p.WriteBytePtr(pos, v)
		case int16:
			p.WriteShortPtr(pos, v)
		case uint16:
			p.WriteShortPtr(pos, int16(v))
		case int:
			p.WriteIntPtr(pos, v)
		case int32:
			p.WriteIntPtr(pos, int(v))
		case uint32:
			p.WriteIntPtr(pos, int(v))
		case float32:
			p.WriteFloatPtr(pos, v)
		case float64:
			p.WriteFloatPtr(pos, float32(v))
		case int64:
			p.WriteLongPtr(pos, v)
		case uint64:
			p.WriteLongPtr(pos, int64(v))
		case string:
			p.WriteStringPtr(pos, v)
		case []byte:
			p.WriteBytesPtr(pos, v)
		default:
			r := reflect.ValueOf(v)
			if r.Kind() == reflect.Pointer {
				r = r.Elem()
			}
			switch r.Kind() {
			case reflect.Struct:
				n := r.NumField()
				for i := 0; i < n; i++ {
					p.WritePtr(pos, r.Field(i).Interface())
				}
			default:
				panic(fmt.Errorf("cannot write type %T to packet: (%+v)", v, v))
			}
		}
	}
	return p
}

// Writes the specified values at the specified position.
func (p *Packet) WriteAt(pos int, values ...any) *Packet {
	return p.WritePtr(&pos, values...)
}

// Writes the specified values at the current position.
func (p *Packet) Write(values ...any) *Packet {
	return p.WritePtr(&p.Pos, values...)
}

/* Replacement */

// Modifies a string at the specified position and advances it.
func (p *Packet) ModifyStringPtr(pos *int, transform func(string) string) *Packet {
	// read original string
	start := *pos
	value := p.ReadStringPtr(pos)
	end := *pos

	// save tail
	tail := make([]byte, len(p.Data)-end)
	copy(tail, p.Data[end:])

	// write modified string & calculate offsets
	*pos = start
	p.WriteStringPtr(pos, transform(value))

	diff := *pos - end
	if diff < 0 {
		p.Data = p.Data[:len(p.Data)+diff]
	} else if diff > 0 {
		p.ensureLength(end+len(tail), diff)
	}

	// restore tail
	copy(p.Data[*pos:], tail)
	return p
}

// Modifies a string at the specified position.
func (p *Packet) ModifyStringAt(pos int, transform func(string) string) *Packet {
	return p.ModifyStringPtr(&pos, transform)
}

// Modifies a string at the current position.
func (p *Packet) ModifyString(transform func(string) string) *Packet {
	return p.ModifyStringPtr(&p.Pos, transform)
}

// Replaces a string at the specified position and advances it.
func (p *Packet) ReplaceStringPtr(pos *int, value string) *Packet {
	return p.ModifyStringPtr(pos, func(_ string) string { return value })
}

// Replaces a string at the specified position.
func (p *Packet) ReplaceStringAt(pos int, value string) *Packet {
	return p.ReplaceStringPtr(&pos, value)
}

// Replaces a string at the current position.
func (p *Packet) ReplaceString(value string) *Packet {
	return p.ReplaceStringPtr(&p.Pos, value)
}

/* Skipping */

// Skips the types indicated specified by the provided values from the current position.
func (p *Packet) Skip(values ...any) {
	for _, v := range values {
		switch v := v.(type) {
		case Parsable:
			v.Parse(p, &p.Pos)
		case bool:
			p.ReadBool()
		case int8, uint8:
			p.ReadByte()
		case int16, uint16:
			p.ReadShort()
		case int, int32, uint32:
			p.ReadInt()
		case int64, uint64:
			p.ReadLong()
		case float32, float64:
			p.ReadFloat()
		case string:
			p.ReadString()
		default:
			p.readReflectPtr(&p.Pos, reflect.ValueOf(v))
		}
	}
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
