package unpack

import (
	"fmt"
	"io"
	"log"
)

type GameDataProcess struct {
	stream           *BinaryStream
	sortIndex        map[string]int64
	queryableField   []string
	searchFieldIndex map[string]int64
	searchFieldType  map[string]int32
	searchFieldCount map[string]int32
}

func (process *GameDataProcess) parseStream(D2OFileBinary *BinaryStream) {
	length := int64(D2OFileBinary.ReadInt32())
	off := D2OFileBinary.Position() + length + 4
	for length > 0 {
		available := D2OFileBinary.BytesAvailable()
		string := D2OFileBinary.ReadString()
		process.queryableField = append(process.queryableField, string)
		process.searchFieldIndex[string] = int64(D2OFileBinary.ReadInt32()) + off
		process.searchFieldType[string] = D2OFileBinary.ReadInt32()
		process.searchFieldCount[string] = D2OFileBinary.ReadInt32()
		length = length - (available - D2OFileBinary.BytesAvailable())
	}
}

func NewGameDataProcess(D2OFileBinary *BinaryStream) *GameDataProcess {
	process := new(GameDataProcess)
	process.stream = D2OFileBinary
	process.sortIndex = make(map[string]int64)
	process.searchFieldIndex = make(map[string]int64)
	process.searchFieldType = make(map[string]int32)
	process.searchFieldCount = make(map[string]int32)
	process.parseStream(D2OFileBinary)
	return process
}

type GameDataField struct {
	name             string
	innerReadMethods []func(*BinaryStream, int) interface{}
	innerTypeNames   []string
	d2oReader        *D2OReader
	readData         func(*BinaryStream, int) interface{}
}

func NewGameDataField(name string, d2oReader *D2OReader) *GameDataField {
	field := new(GameDataField)
	field.name = name
	field.d2oReader = d2oReader
	field.innerReadMethods = make([]func(*BinaryStream, int) interface{}, 0)
	field.innerTypeNames = make([]string, 0)
	return field
}

func (field *GameDataField) readType(D2OFileBinary *BinaryStream) {
	readID := D2OFileBinary.ReadInt32()
	field.readData = field.getReadMethod(readID, D2OFileBinary)
}

// readInteger
func (field *GameDataField) readInteger(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	return D2OFileBinary.ReadInt32()
}

// readBoolean
func (field *GameDataField) readBoolean(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	return D2OFileBinary.ReadBool()
}

// readString
func (field *GameDataField) readString(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	str := D2OFileBinary.ReadString()
	if str == "null" {
		str = ""
	}
	return str
}

// readNumber
func (field *GameDataField) readNumber(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	return D2OFileBinary.ReadFloat64()
}

// readI18n
func (field *GameDataField) readI18n(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	return D2OFileBinary.ReadInt32()
}

// readUnsignedInteger
func (field *GameDataField) readUnsignedInteger(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	return D2OFileBinary.ReadUint32()
}

// readVector
func (field *GameDataField) readVector(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	vectorSize := int(D2OFileBinary.ReadInt32())
	vector := make([]interface{}, vectorSize)
	for i := 0; i < vectorSize; i++ {
		vector[i] = field.innerReadMethods[vecIndex](D2OFileBinary, vecIndex+1)
	}
	return vector
}

// readObject
func (field *GameDataField) readObject(D2OFileBinary *BinaryStream, vecIndex int) interface{} {
	objectID := D2OFileBinary.ReadInt32()
	if objectID == -1431655766 {
		return nil
	}

	obj := field.d2oReader.GetClassDefinition(objectID)
	return obj.read(D2OFileBinary)
}

func (field *GameDataField) getReadMethod(readID int32, D2OFileBinary *BinaryStream) func(*BinaryStream, int) interface{} {
	if readID == -1 {
		return field.readInteger
	} else if readID == -2 {
		return field.readBoolean
	} else if readID == -3 {
		return field.readString
	} else if readID == -4 {
		return field.readNumber
	} else if readID == -5 {
		return field.readI18n
	} else if readID == -6 {
		return field.readUnsignedInteger
	} else if readID == -99 {
		field.innerTypeNames = append(field.innerTypeNames, D2OFileBinary.ReadString())
		field.innerReadMethods = append([]func(*BinaryStream, int) interface{}{field.getReadMethod(D2OFileBinary.ReadInt32(), D2OFileBinary)}, field.innerReadMethods...)
		return field.readVector
	} else {
		if readID > 0 {
			return field.readObject
		}
		panic("Unknown type " + string(readID) + ".")
	}
}

type GameDataClassDefinition struct {
	class     string
	fields    []*GameDataField
	d2oReader *D2OReader
}

func NewGameDataClassDefinition(class_pkg string, class_name string, d2oReader *D2OReader) *GameDataClassDefinition {
	return &GameDataClassDefinition{
		class:     class_pkg + "." + class_name,
		fields:    make([]*GameDataField, 0),
		d2oReader: d2oReader,
	}
}

func (classDef *GameDataClassDefinition) read(D2OFileBinary *BinaryStream) interface{} {
	obj := make(map[string]interface{})
	for _, field := range classDef.fields {
		obj[field.name] = field.readData(D2OFileBinary, 0)
	}
	return obj
}

func (classDef *GameDataClassDefinition) addField(name string, D2OFileBinary *BinaryStream) {
	field := NewGameDataField(name, classDef.d2oReader)
	field.readType(D2OFileBinary)
	classDef.fields = append(classDef.fields, field)
}

// D2OReader represents a reader for D2O files.
type D2OReader struct {
	stream            io.Reader
	streamStartIndex  int64
	classes           map[int32]*GameDataClassDefinition
	counter           int
	D2OFileBinary     *BinaryStream
	gameDataProcessor *GameDataProcess
}

// NewD2OReader creates a new D2OReader instance with the given stream.
func NewD2OReader(stream io.ReadWriteSeeker) (*D2OReader, error) {
	reader := &D2OReader{
		stream:           stream,
		classes:          make(map[int32]*GameDataClassDefinition),
		streamStartIndex: 7,
		counter:          0,
	}

	D2OFileBinary := BinaryStream{baseStream: stream, bigEndian: true}
	reader.D2OFileBinary = &D2OFileBinary

	stringHeader := string(D2OFileBinary.ReadBytes(3))
	baseOffset := int32(0)
	if stringHeader != "D2O" {
		D2OFileBinary.Seek(0, io.SeekStart)
		stringHeader := D2OFileBinary.ReadString()
		if stringHeader != "AKSF" {
			return nil, fmt.Errorf("malformed game data file")
		}
		D2OFileBinary.ReadUint16()
		baseOffset = D2OFileBinary.ReadInt32()
		D2OFileBinary.Seek(int64(baseOffset), io.SeekCurrent)
		reader.streamStartIndex = D2OFileBinary.Position() + 7
		stringHeaderBytes := string(D2OFileBinary.ReadBytes(3))
		if stringHeaderBytes != "D2O" {
			return nil, fmt.Errorf("malformed game data file")
		}
	}

	offset := D2OFileBinary.ReadInt32()
	D2OFileBinary.Seek(int64(baseOffset+offset), io.SeekStart)
	indexNumber := D2OFileBinary.ReadInt32()
	index := 0
	indexOffsets := make(map[int32]int32)

	for int32(index) < indexNumber {
		indexID := D2OFileBinary.ReadInt32()
		offset := D2OFileBinary.ReadInt32()
		indexOffsets[indexID] = baseOffset + offset
		reader.counter += 1
		index = index + 8
	}

	classNumber := D2OFileBinary.ReadInt32()
	classIndex := 0

	for int32(classIndex) < classNumber {
		classID := D2OFileBinary.ReadInt32()
		reader.readClassDefinition(classID, &D2OFileBinary)
		classIndex += 1
	}

	if D2OFileBinary.BytesAvailable() > 0 {
		reader.gameDataProcessor = NewGameDataProcess(&D2OFileBinary)
	}

	return reader, nil
}

func (dr *D2OReader) GetObjects() []interface{} {
	if dr.counter == 0 {
		return nil
	}

	D2OFileBinary := dr.D2OFileBinary
	D2OFileBinary.SetPosition(dr.streamStartIndex)
	objects := make([]interface{}, 0)
	i := 0

	for i < dr.counter {
		classId := D2OFileBinary.ReadInt32()
		class := dr.classes[classId]
		if class == nil {
			log.Fatal("class is nil")
		}

		object := class.read(D2OFileBinary)
		objects = append(objects, object)
		i += 1
	}

	return objects
}

// GetClassDefinition returns the class definition for a given object_id.
func (dr *D2OReader) GetClassDefinition(objectID int32) *GameDataClassDefinition {
	return dr.classes[objectID]
}

func (dr *D2OReader) readClassDefinition(classID int32, D2OFileBinary *BinaryStream) {
	className := D2OFileBinary.ReadString()
	classPackage := D2OFileBinary.ReadString()
	classDef := NewGameDataClassDefinition(classPackage, className, dr)
	fieldNumber := D2OFileBinary.ReadInt32()
	fieldIndex := 0

	for int32(fieldIndex) < fieldNumber {
		field := D2OFileBinary.ReadString()
		classDef.addField(field, D2OFileBinary)
		fieldIndex += 1
	}

	dr.classes[classID] = classDef
}
