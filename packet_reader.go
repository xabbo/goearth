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
)

// PacketReader provides methods to read values from an underlying packet.
type PacketReader interface {
	// Returns a new [PacketReader] referencing a copy of this reader's position.
	Copy() PacketReader
	// Gets the [ClientType] of the underlying packet.
	Client() ClientType
	// Gets the [Header] of the underlying packet.
	Header() Header
	// Gets the position of the reader.
	Pos() int
	// Sets the position of the reader.
	Seek(pos int)
	// Gets the size of the underlying packet.
	Length() int
	// Gets the number of available bytes that can be read from the current position.
	Available() int
	// Reads into the specified byte slice from the current position and advances it.
	ReadBuffer(buf []byte)
	// Reads a byte from the current position and advances it.
	ReadByte() (value byte)
	// Reads a bool from the current position and advances it.
	//
	// Read as a VL64 on Shockwave, otherwise as a byte.
	ReadBool() (value bool)
	// Copies `n` bytes from the current position and advances it.
	ReadBytes(n int) (value []byte)
	// Reads a short from the current position and advances it.
	//
	// Read as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
	ReadShort() (value int16)
	// Reads an int from the current position and advances it.
	//
	// Read as a VL64 on Shockwave, otherwise as an int32.
	ReadInt() (value int)
	// Reads a float from the current position and advances it.
	//
	// Read as a string and parsed to a float on Flash and Shockwave sessions, otherwise as a float32.
	ReadFloat() float32
	// Reads a long from the current position and advances it.
	//
	// Only supported on Unity sessions.
	ReadLong() (value int64)
	// Reads a string from the current position and advances it.
	//
	// Read as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
	// otherwise as a short length-prefixed UTF-8 string.
	ReadString() (value string)
	// Reads into the specified variables from the current position and advances it.
	// The provided variables must be a pointer type or implement [Parsable].
	Read(vars ...any)
	// Skips the types specified by the provided values, advancing the current position.
	Skip(values ...any)
}

type packetReader struct {
	p   *Packet
	pos *int
}

// Ensures `n` bytes can be read from the current position.
func assertCanRead(r PacketReader, n int) {
	if r.Pos() < 0 {
		panic(fmt.Errorf("packet position cannot be < 0"))
	}
	if (r.Pos() + n) > r.Length() {
		panic(fmt.Errorf("attempt to read past packet length"))
	}
}

func (r packetReader) Data() []byte {
	return r.p.Data
}

func (r packetReader) Copy() PacketReader {
	pos := *r.pos
	return packetReader{r.p, &pos}
}

func (r packetReader) Client() ClientType {
	return r.p.Client
}

func (r packetReader) Header() Header {
	return r.p.Header
}

func (r packetReader) Pos() int {
	return *r.pos
}

func (r packetReader) Seek(pos int) {
	if pos < 0 {
		panic(fmt.Errorf("attempt to seek to a negative value"))
	}
	if pos > r.p.Length() {
		panic(fmt.Errorf("attempt to seek past the packet length"))
	}
	*r.pos = pos
}

// Gets the size of the underlying packet.
func (r packetReader) Length() int {
	return r.p.Length()
}

// Gets the number of available bytes that can be read from the current position.
func (r packetReader) Available() int {
	return r.p.Length() - *r.pos
}

// Reads into the specified byte slice from the current position and advances it.
func (r packetReader) ReadBuffer(buf []byte) {
	length := len(buf)
	assertCanRead(r, length)
	copy(buf, r.p.Data[*r.pos:])
	*r.pos += length
}

// Reads a byte from the current position and advances it.
func (r packetReader) ReadByte() (value byte) {
	assertCanRead(r, 1)
	value = r.p.Data[*r.pos]
	dbgPkt.Printf("%d", value)
	*r.pos++
	return
}

// Reads a bool from the current position and advances it.
//
// Read as a VL64 on Shockwave, otherwise as a byte.
func (r packetReader) ReadBool() (value bool) {
	assertCanRead(r, 1)
	var i int
	switch r.p.Client {
	case Shockwave:
		if encoding.VL64DecodeLen(r.p.Data[*r.pos]) != 1 {
			panic(fmt.Errorf("attempt to read boolean when VL64 length > 1"))
		}
		i = encoding.VL64Decode(r.p.Data[*r.pos : *r.pos+1])
		dbgPkt.Printf("vl64: %t", i == 1)
	default:
		i = int(r.p.Data[*r.pos])
		dbgPkt.Printf("%t", i == 1)
	}
	if i != 0 && i != 1 {
		panic(fmt.Errorf("attempt to read boolean from non-boolean value: %d", i))
	}
	value = i == 1
	*r.pos++
	return
}

// Copies `n` bytes from the current position and advances it.
func (r packetReader) ReadBytes(n int) (value []byte) {
	assertCanRead(r, n)
	value = make([]byte, n)
	dbgPkt.Printf("%v", value)
	*r.pos += copy(value, r.p.Data[*r.pos:])
	return
}

// Reads a short from the current position and advances it.
//
// Read as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (r packetReader) ReadShort() (value int16) {
	if r.p.Client == Shockwave {
		switch r.p.Header.Dir {
		case In:
			var vl64 VL64
			vl64.Parse(r)
			dbgPkt.Printf("vl64: %d", vl64)
			return int16(vl64)
		case Out:
			var b64 B64
			b64.Parse(r)
			dbgPkt.Printf("b64: %d", b64)
			return int16(b64)
		default:
			panic(fmt.Errorf("%w: unknown packet direction when reading short on shockwave session",
				errors.ErrUnsupported))
		}
	}
	assertCanRead(r, 2)
	value = int16(binary.BigEndian.Uint16(r.p.Data[*r.pos:]))
	dbgPkt.Printf("%d", value)
	*r.pos += 2
	return
}

// Reads an int from the current position and advances it.
//
// Read as a VL64 on Shockwave, otherwise as an int32.
func (r packetReader) ReadInt() (value int) {
	if r.p.Client == Shockwave {
		var vl64 VL64
		vl64.Parse(r)
		dbgPkt.Printf("vl64: %d", vl64)
		return int(vl64)
	}
	assertCanRead(r, 4)
	value = int(int32(binary.BigEndian.Uint32(r.p.Data[*r.pos:])))
	dbgPkt.Printf("%d", value)
	*r.pos += 4
	return
}

// Reads a float from the current position and advances it.
//
// Read as a string and parsed to a float on Flash and Shockwave sessions, otherwise as a float32.
func (r packetReader) ReadFloat() float32 {
	switch r.p.Client {
	case Flash, Shockwave:
		s := r.ReadString()
		value, err := strconv.ParseFloat(s, 32)
		if err != nil {
			panic(fmt.Errorf("failed to parse float: %w", err))
		}
		dbgPkt.Printf("string: %f", value)
		return float32(value)
	default:
		assertCanRead(r, 4)
		bits := binary.BigEndian.Uint32(r.p.Data[*r.pos:])
		value := math.Float32frombits(bits)
		dbgPkt.Printf("%f", value)
		*r.pos += 4
		return value
	}
}

// Reads a long from the current position and advances it.
//
// Only supported on Unity sessions.
func (r packetReader) ReadLong() (value int64) {
	if r.p.Client != Unity {
		panic(fmt.Errorf("%w: attempt to read long on client: %s", errors.ErrUnsupported, r.p.Client))
	}
	assertCanRead(r, 8)
	x := binary.BigEndian.Uint64(r.p.Data[*r.pos:])
	ptr := unsafe.Pointer(&x)
	value = *(*int64)(ptr)
	dbgPkt.Printf("%d", value)
	*r.pos += 8
	return
}

// Reads a string from the current position and advances it.
//
// Read as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (r packetReader) ReadString() (value string) {
	if r.p.Client == Shockwave && r.p.Header.Dir == In {
		i := *r.pos
		for i < len(r.p.Data) && r.p.Data[i] != 2 {
			i++
		}
		// some packets don't have a terminator byte for the final string..
		// if i >= len(r.p.Data) {
		// 	panic(fmt.Errorf("unterminated string"))
		// }
		value = string(r.p.Data[*r.pos:i])
		*r.pos = min(len(r.p.Data), i+1)
	} else {
		// Read length without advancing the current position.
		length := int(uint16(r.Copy().ReadShort()))
		assertCanRead(r, 2+length)
		value = string(r.p.Data[*r.pos+2 : *r.pos+2+length])
		*r.pos += 2 + length
	}
	dbgPkt.Printf("%q", value)
	return
}

// Reads into the specified variables from the current position and advances it.
// The provided variables must be a pointer type or implement [Parsable].
func (r packetReader) Read(vars ...any) {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Errorf("packet read failed: %v", r))
		}
	}()
	for _, v := range vars {
		if !r.readInterface(v) {
			r.readReflect(reflect.ValueOf(v))
		}
	}
}

func (r packetReader) readReflect(v reflect.Value) {
	if v.CanAddr() {
		if parsable, ok := v.Addr().Interface().(Parsable); ok {
			dbgPkt.Printf("Parsable: %s", reflect.TypeOf(v).Name())
			parsable.Parse(r)
			return
		}
	}
	switch v.Kind() {
	case reflect.Pointer:
		r.readReflect(v.Elem())
	case reflect.Array:
		n := v.Len()
		dbgPkt.Printf("array[%d]", n)
		for i := 0; i < n; i++ {
			r.readReflect(v.Index(i))
		}
	case reflect.Slice:
		t := v.Type()
		var len Length
		len.Parse(r)
		dbgPkt.Printf("slice[%d]", len)
		slc := reflect.MakeSlice(t, int(len), int(len))
		for i := 0; i < int(len); i++ {
			r.readReflect(slc.Index(i))
		}
		v.Set(slc)
	case reflect.Struct:
		n := v.NumField()
		dbgPkt.Printf("struct: %s", v.Type().Name())
		for i := 0; i < n; i++ {
			if v := v.Field(i); v.CanSet() {
				r.readReflect(v)
			}
		}
	case reflect.Interface:
		if r.readInterface(v.Interface()) {
			return
		}
	default:
		if v.CanAddr() && v.CanInterface() {
			if r.readInterface(v.Addr().Interface()) {
				return
			}
		}
		panic(fmt.Errorf("cannot read unsupported type: %+v", v.Type()))
	}
}

func (r packetReader) readInterface(v any) bool {
	switch v := v.(type) {
	case Parsable:
		dbgPkt.Printf("Parsable: %s", reflect.TypeOf(v).Name())
		v.Parse(r)
	case *bool:
		*v = r.ReadBool()
	case *byte:
		*v = r.ReadByte()
	case *int16:
		*v = r.ReadShort()
	case *uint16:
		*v = uint16(r.ReadShort())
	case *int:
		*v = r.ReadInt()
	case *uint:
		*v = uint(r.ReadInt())
	case *int32:
		*v = int32(r.ReadInt())
	case *uint32:
		*v = uint32(r.ReadInt())
	case *float32:
		*v = r.ReadFloat()
	case *float64:
		*v = float64(r.ReadFloat())
	case *int64:
		*v = r.ReadLong()
	case *uint64:
		*v = uint64(r.ReadLong())
	case *string:
		*v = r.ReadString()
	default:
		return false
	}
	return true
}

/* Skipping */

// Skips the types specified by the provided values, advancing the current position.
func (r packetReader) Skip(values ...any) {
	for _, v := range values {
		switch v := v.(type) {
		case Parsable:
			v.Parse(r)
		case bool:
			r.ReadBool()
		case int8, uint8:
			r.ReadByte()
		case int16, uint16:
			r.ReadShort()
		case int, int32, uint32:
			r.ReadInt()
		case int64, uint64:
			r.ReadLong()
		case float32, float64:
			r.ReadFloat()
		case string:
			r.ReadString()
		default:
			r.readReflect(reflect.ValueOf(v))
		}
	}
}
