package goearth

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// PacketWriter provides methods to write values to an underlying packet.
type PacketWriter interface {
	// Returns a [PacketReader] referencing this writer's packet and position.
	Reader() PacketReader
	// Gets the [ClientType] of the underlying packet.
	Client() ClientType
	// Gets the [Header] of the underlying packet.
	Header() Header
	// Gets the position of the writer.
	Pos() int
	// Sets the position of the writer.
	Seek(pos int)
	// Gets the length of the underlying packet.
	Length() int
	// Allocates `n` bytes from the current position and returns a slice referencing the allocated byte range.
	Allocate(n int) []byte
	// Writes a bool at the current position and advances it.
	//
	// Written as a VL64 on Shockwave, otherwise as a byte.
	WriteBool(value bool)
	// Writes a byte at the current position and advances it.
	WriteByte(value byte)
	// Writes a slice of bytes at the current position and advances it.
	WriteBytes(value []byte)
	// Writes a short at the current position and advances it.
	//
	// Written as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
	WriteShort(value int16)
	// Writes an int at the current position and advances it.
	//
	// Written as a VL64 on Shockwave, otherwise as an int32.
	WriteInt(value int)
	// Writes a float at the current position and advances it.
	//
	// Written as a string on Flash and Shockwave sessions, otherwise as a float32.
	WriteFloat(value float32)
	// Writes a long at the current position and advances it.
	//
	// Only supported on Unity sessions.
	WriteLong(value int64)
	// Writes a string at the current position and advances it.
	//
	// Written as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
	// otherwise as a short length-prefixed UTF-8 string.
	WriteString(value string)
	// Writes the specified values at the current position and advances it.
	Write(values ...any)
	// Modifies a string at the specified position and advances it.
	ModifyString(transform func(string) string)
	// Replaces a string at the specified position and advances it.
	ReplaceString(value string)
	// Skips the types specified by the provided values, advancing the current position.
	Skip(values ...any)
}

type packetWriter struct {
	p   *Packet
	pos *int
}

// Returns a [PacketReader] referencing this writer's packet and position.
func (w packetWriter) Reader() PacketReader {
	return packetReader{w.p, w.pos}
}

// Gets the [ClientType] of the underlying packet.
func (w packetWriter) Client() ClientType {
	return w.p.Client
}

// Gets the [Header] of the underlying packet.
func (w packetWriter) Header() Header {
	return w.p.Header
}

// Gets the position of the writer.
func (w packetWriter) Pos() int {
	return *w.pos
}

// Sets the position of the writer.
func (w packetWriter) Seek(pos int) {
	if pos < 0 {
		panic(fmt.Errorf("attempt to seek to a negative value"))
	}
	if pos > w.p.Length() {
		panic(fmt.Errorf("attempt to seek past the packet length"))
	}
	*w.pos = pos
}

// Gets the length of the underlying packet.
func (w packetWriter) Length() int {
	return w.p.Length()
}

// Ensures the packet has enough capacity to store `n` bytes from the current position.
func (w packetWriter) ensureLength(pos int, n int) {
	if pos < 0 {
		panic("position cannot be < 0.")
	}
	if pos > len(w.p.Data) {
		panic("position cannot be > packet length.")
	}
	if length := (pos + n); length > len(w.p.Data) {
		if cap(w.p.Data) > len(w.p.Data) {
			w.p.Data = w.p.Data[:min(cap(w.p.Data), length)]
		}
		if length > len(w.p.Data) {
			w.p.Data = append(w.p.Data, make([]byte, length-len(w.p.Data))...)
		}
	}
}

// Allocates `n` bytes from the current position and returns a slice referencing the allocated byte range.
func (w packetWriter) Allocate(n int) []byte {
	w.ensureLength(*w.pos, n)
	*w.pos += n
	return w.p.Data[*w.pos-n : *w.pos]
}

// Writes a bool at the current position and advances it.
//
// Written as a VL64 on Shockwave, otherwise as a byte.
func (w packetWriter) WriteBool(value bool) {
	b := uint8(0)
	if value {
		b = 1
	}
	if w.p.Client == Shockwave {
		VL64(b).Compose(w)
	} else {
		w.Allocate(1)[0] = b
	}
}

// Writes a byte at the current position and advances it.
func (w packetWriter) WriteByte(value byte) {
	w.Allocate(1)[0] = value
}

// Writes a slice of bytes at the current position and advances it.
func (w packetWriter) WriteBytes(value []byte) {
	copy(w.Allocate(len(value)), value)
}

// Writes a short at the current position and advances it.
//
// Written as a VL64 on incoming Shockwave, B64 on outgoing Shockwave, otherwise as an int16.
func (w packetWriter) WriteShort(value int16) {
	if w.p.Client == Shockwave {
		switch w.p.Header.Dir {
		case In:
			VL64(value).Compose(w)
		case Out:
			B64(value).Compose(w)
		default:
			panic(fmt.Errorf("%w: unknown packet direction when writing short on shockwave session",
				errors.ErrUnsupported))
		}
	} else {
		binary.BigEndian.PutUint16(w.Allocate(2), uint16(value))
	}
}

// Writes an int at the current position and advances it.
//
// Written as a VL64 on Shockwave, otherwise as an int32.
func (w packetWriter) WriteInt(value int) {
	if w.p.Client == Shockwave {
		VL64(value).Compose(w)
	} else {
		binary.BigEndian.PutUint32(w.Allocate(4), uint32(value))
	}
}

// Writes a float at the current position and advances it.
//
// Written as a string on Flash and Shockwave sessions, otherwise as a float32.
func (w packetWriter) WriteFloat(value float32) {
	switch w.p.Client {
	case Flash, Shockwave:
		w.WriteString(strconv.FormatFloat(float64(value), 'f', -1, 32))
	case Unity:
		binary.BigEndian.PutUint32(w.Allocate(4), math.Float32bits(value))
	default:
		panic(fmt.Errorf("attempt to write float on unknown client: %s", w.p.Client))
	}
}

// Writes a long at the current position and advances it.
//
// Only supported on Unity sessions.
func (w packetWriter) WriteLong(value int64) {
	if w.p.Client == Shockwave {
		panic(fmt.Errorf("%w: attempt to write long on client: %s", errors.ErrUnsupported, w.p.Client))
	}
	binary.BigEndian.PutUint64(w.Allocate(8), uint64(value))
}

// Writes a string at the current position and advances it.
//
// Written as a UTF-8 string terminated with an 0x02 byte on (incoming) Shockwave,
// otherwise as a short length-prefixed UTF-8 string.
func (w packetWriter) WriteString(value string) {
	b := []byte(value)
	size := len(b)

	if w.p.Client == Shockwave && w.p.Header.Dir == In {
		dst := w.Allocate(1 + size)
		copy(dst, b)
		dst[len(dst)-1] = 2
	} else {
		if size >= (1 << 16) {
			panic(fmt.Errorf("string length cannot fit into a uint16"))
		}

		w.WriteShort(int16(size))
		dst := w.Allocate(size)
		copy(dst, b)
	}
}

// Writes the specified values at the current position and advances it.
func (w packetWriter) Write(values ...any) {
	for _, value := range values {
		switch v := value.(type) {
		case Composable:
			v.Compose(w)
		case bool:
			w.WriteBool(v)
		case int8:
			w.WriteByte(byte(v))
		case uint8:
			w.WriteByte(v)
		case int16:
			w.WriteShort(v)
		case uint16:
			w.WriteShort(int16(v))
		case int:
			w.WriteInt(v)
		case int32:
			w.WriteInt(int(v))
		case uint32:
			w.WriteInt(int(v))
		case float32:
			w.WriteFloat(v)
		case float64:
			w.WriteFloat(float32(v))
		case int64:
			w.WriteLong(v)
		case uint64:
			w.WriteLong(int64(v))
		case string:
			w.WriteString(v)
		case []byte:
			w.WriteBytes(v)
		default:
			r := reflect.ValueOf(v)
			if r.Kind() == reflect.Pointer {
				r = r.Elem()
			}
			switch r.Kind() {
			case reflect.Struct:
				n := r.NumField()
				for i := 0; i < n; i++ {
					w.Write(r.Field(i).Interface())
				}
			default:
				panic(fmt.Errorf("cannot write type %T to packet: (%+v)", v, v))
			}
		}
	}
}

/* Replacement */

// Modifies a string at the specified position and advances it.
func (w packetWriter) ModifyString(transform func(string) string) {
	// read original string
	start := *w.pos
	value := w.Reader().ReadString()
	end := *w.pos

	// save tail
	tail := make([]byte, len(w.p.Data)-end)
	copy(tail, w.p.Data[end:])

	// write modified string & calculate offsets
	*w.pos = start
	w.WriteString(transform(value))

	diff := *w.pos - end
	if diff < 0 {
		w.p.Data = w.p.Data[:len(w.p.Data)+diff]
	} else if diff > 0 {
		w.ensureLength(end+len(tail), diff)
	}

	// restore tail
	copy(w.p.Data[*w.pos:], tail)
}

// Replaces a string at the specified position and advances it.
func (w packetWriter) ReplaceString(value string) {
	w.ModifyString(func(_ string) string { return value })
}

/* Skipping */

// Skips the types specified by the provided values, advancing the current position.
func (w packetWriter) Skip(values ...any) {
	w.Reader().Skip(values)
}
