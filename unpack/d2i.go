package unpack

import (
	"fmt"
	"io"
)

type D2I struct {
	Stream io.ReadWriteSeeker
	Obj    map[string]map[string]interface{}
}

func NewD2I(stream io.ReadWriteSeeker) *D2I {
	return &D2I{
		Stream: stream,
		Obj:    make(map[string]map[string]interface{}),
	}
}

func (d *D2I) Read() map[string]map[string]interface{} {
	raw := NewBinaryStream(d.Stream, true)

	indexs := make(map[int]int)
	unDiacriticalIndex := make(map[int]int)

	d.Obj["texts"] = make(map[string]interface{})
	d.Obj["nameText"] = make(map[string]interface{})
	d.Obj["idText"] = make(map[string]interface{})

	indexesPointer := raw.ReadInt32()
	d.Stream.Seek(int64(indexesPointer), io.SeekStart)

	i := 0
	indexesLength := raw.ReadInt32()
	for i < int(indexesLength) {
		key := int(raw.ReadInt32())
		diacriticalText := raw.ReadBool()
		pointer := int(raw.ReadInt32())
		indexs[pointer] = key

		if diacriticalText {
			i += 4
			unDiacriticalIndex[key] = int(raw.ReadInt32())
		} else {
			unDiacriticalIndex[key] = pointer
		}
		i += 9
	}

	indexesLength = raw.ReadInt32()
	for indexesLength > 0 {
		position, _ := d.Stream.Seek(0, io.SeekCurrent)
		textKey := raw.ReadString()
		pointer := int(raw.ReadInt32())
		d.Obj["nameText"][textKey] = indexs[pointer]
		position_updated, _ := d.Stream.Seek(0, io.SeekCurrent)
		indexesLength = (indexesLength - (int32(position_updated) - int32(position)))
	}

	i = 0
	indexesLength = raw.ReadInt32()
	for indexesLength > 0 {
		position, _ := d.Stream.Seek(0, io.SeekCurrent)
		i += 1
		d.Obj["idText"][fmt.Sprint(raw.ReadInt32())] = i
		position_updated, _ := d.Stream.Seek(0, io.SeekCurrent)
		indexesLength = (indexesLength - (int32(position_updated) - int32(position)))
	}

	for pointer, key := range indexs {
		d.Stream.Seek(int64(pointer), io.SeekStart)
		d.Obj["texts"][fmt.Sprint(key)] = raw.ReadString()
	}

	return d.Obj
}
