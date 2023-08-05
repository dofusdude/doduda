package unpack

import (
	"encoding/binary"
	"io"
)

type BinaryStream struct {
	baseStream io.ReadWriteSeeker
	bigEndian  bool
}

func NewBinaryStream(baseStream io.ReadWriteSeeker, bigEndian bool) *BinaryStream {
	return &BinaryStream{
		baseStream: baseStream,
		bigEndian:  bigEndian,
	}
}

// Comment functions

func (bs *BinaryStream) Position() int64 {
	position, _ := bs.baseStream.Seek(0, io.SeekCurrent)
	return position
}

func (bs *BinaryStream) SetPosition(position int64) {
	bs.baseStream.Seek(position, io.SeekStart)
}

func (bs *BinaryStream) BytesAvailable() int64 {
	eof, _ := bs.baseStream.Seek(0, io.SeekEnd)
	position := bs.Position()
	bs.baseStream.Seek(position, io.SeekStart)
	return eof - position
}

// Write functions

func (bs *BinaryStream) writeBytes(data []byte) {
	bs.baseStream.Write(data)
}

func (bs *BinaryStream) writeValue(format string, data interface{}) {
	var byteOrder binary.ByteOrder
	if bs.bigEndian {
		byteOrder = binary.BigEndian
	} else {
		byteOrder = binary.LittleEndian
	}

	binary.Write(bs.baseStream, byteOrder, data)
}

func (bs *BinaryStream) WriteBool(value bool) {
	if value {
		bs.WriteChar(1)
	} else {
		bs.WriteChar(0)
	}
}

func (bs *BinaryStream) WriteChar(value int8) {
	bs.writeValue("b", value)
}

func (bs *BinaryStream) WriteUchar(value uint8) {
	bs.writeValue("B", value)
}

func (bs *BinaryStream) WriteInt16(value int16) {
	bs.writeValue("h", value)
}

func (bs *BinaryStream) WriteUint16(value uint16) {
	bs.writeValue("H", value)
}

func (bs *BinaryStream) WriteInt32(value int32) {
	bs.writeValue("i", value)
}

func (bs *BinaryStream) WriteUint32(value uint32) {
	bs.writeValue("I", value)
}

func (bs *BinaryStream) WriteInt64(value int64) {
	bs.writeValue("q", value)
}

func (bs *BinaryStream) WriteUint64(value uint64) {
	bs.writeValue("Q", value)
}

func (bs *BinaryStream) WriteFloat32(value float32) {
	bs.writeValue("f", value)
}

func (bs *BinaryStream) WriteFloat64(value float64) {
	bs.writeValue("d", value)
}

func (bs *BinaryStream) WriteString(value string) {
	length := uint16(len(value))
	bs.WriteUint16(length)
	bs.writeBytes([]byte(value))
}

// Read functions

func (bs *BinaryStream) readValue(format string, data interface{}) {
	var byteOrder binary.ByteOrder
	if bs.bigEndian {
		byteOrder = binary.BigEndian
	} else {
		byteOrder = binary.LittleEndian
	}

	binary.Read(bs.baseStream, byteOrder, data)
}

func (bs *BinaryStream) ReadBool() bool {
	var value int8
	bs.readValue("b", &value)
	return value == 1
}

func (bs *BinaryStream) ReadChar() int8 {
	var value int8
	bs.readValue("b", &value)
	return value
}

func (bs *BinaryStream) ReadUchar() uint8 {
	var value uint8
	bs.readValue("B", &value)
	return value
}

func (bs *BinaryStream) ReadInt16() int16 {
	var value int16
	bs.readValue("h", &value)
	return value
}

func (bs *BinaryStream) ReadUint16() uint16 {
	var value uint16
	bs.readValue("H", &value)
	return value
}

func (bs *BinaryStream) ReadInt32() int32 {
	var value int32
	bs.readValue("i", &value)
	return value
}

func (bs *BinaryStream) ReadUint32() uint32 {
	var value uint32
	bs.readValue("I", &value)
	return value
}

func (bs *BinaryStream) ReadInt64() int64 {
	var value int64
	bs.readValue("q", &value)
	return value
}

func (bs *BinaryStream) ReadUint64() uint64 {
	var value uint64
	bs.readValue("Q", &value)
	return value
}

func (bs *BinaryStream) ReadFloat32() float32 {
	var value float32
	bs.readValue("f", &value)
	return value
}

func (bs *BinaryStream) ReadFloat64() float64 {
	var value float64
	bs.readValue("d", &value)
	return value
}

func (bs *BinaryStream) ReadString() string {
	length := bs.ReadUint16()
	data := make([]byte, length)
	bs.baseStream.Read(data)
	return string(data)
}

func (bs *BinaryStream) ReadBytes(length int) []byte {
	data := make([]byte, length)
	bs.ReadBytesIntoBuffer(data)
	return data
}

func (bs *BinaryStream) ReadBytesIntoBuffer(buffer []byte) {
	_, err := io.ReadFull(bs.baseStream, buffer)
	if err != nil {
		panic(err)
	}
}

func (bs *BinaryStream) Seek(position int64, whence int) (int64, error) {
	return bs.baseStream.Seek(position, whence)
}

func (bs *BinaryStream) ReadShort() int16 {
	var value int16
	bs.readValue("h", &value)
	return value
}
