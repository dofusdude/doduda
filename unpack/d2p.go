package unpack

import "io"

type D2P struct {
	Stream           io.ReadWriteSeeker
	baseOffset       uint32
	baseLength       uint32
	indexesOffset    uint32
	numberIndexes    uint32
	propertiesOffset uint32
	numberProperties uint32
	filesPosition    map[string]map[string]uint32
	files            map[string][]byte
	properties       map[string]string
}

func NewD2P(stream io.ReadWriteSeeker) *D2P {
	d2p := &D2P{
		Stream:        stream,
		filesPosition: make(map[string]map[string]uint32),
		files:         make(map[string][]byte),
		properties:    make(map[string]string),
	}

	d2pFileBinary := NewBinaryStream(stream, true)
	bytesHeader := string(d2pFileBinary.ReadBytes(2))
	if bytesHeader != "\x02\x01" {
		panic("Invalid D2P file")
	}

	stream.Seek(-24, io.SeekEnd)
	d2p.baseOffset = d2pFileBinary.ReadUint32()
	d2p.baseLength = d2pFileBinary.ReadUint32()
	d2p.indexesOffset = d2pFileBinary.ReadUint32()
	d2p.numberIndexes = d2pFileBinary.ReadUint32()
	d2p.propertiesOffset = d2pFileBinary.ReadUint32()
	d2p.numberProperties = d2pFileBinary.ReadUint32()

	stream.Seek(int64(d2p.indexesOffset), io.SeekStart)
	for i := uint32(0); i < d2p.numberIndexes; i++ {
		fileName := d2pFileBinary.ReadString()
		offset := d2pFileBinary.ReadUint32()
		length := d2pFileBinary.ReadUint32()
		d2p.filesPosition[fileName] = map[string]uint32{
			"offset": offset + d2p.baseOffset,
			"length": length,
		}
	}

	stream.Seek(int64(d2p.propertiesOffset), io.SeekStart)

	for i := uint32(0); i < d2p.numberProperties; i++ {
		pptyType := d2pFileBinary.ReadString()
		pptyValue := d2pFileBinary.ReadString()
		d2p.properties[pptyType] = pptyValue
	}

	for fileName, position := range d2p.filesPosition {
		d2p.Stream.Seek(int64(position["offset"]), io.SeekStart)
		file := d2pFileBinary.ReadBytes(int(position["length"]))
		d2p.files[fileName] = file
	}

	return d2p
}

func (d2p *D2P) GetFiles() map[string]map[string]interface{} {
	toReturn := make(map[string]map[string]interface{})
	for fileName, position := range d2p.filesPosition {
		object_ := make(map[string]interface{})
		object_["position"] = position
		if d2p.files != nil {
			object_["binary"] = d2p.files[fileName]
		}
		toReturn[fileName] = object_
	}
	return toReturn
}
