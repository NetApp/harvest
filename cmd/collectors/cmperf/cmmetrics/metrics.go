package cmmetrics

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"iter"
	"math"
	"os"

	"github.com/VictoriaMetrics/easyproto"
)

type ObjectCollection struct {
	Timestamp uint64
	Period    uint32
	Node      string
	Schema    ObjectSchema
	Data      ObjectData
}

type ObjectSchema struct {
	Name          string
	CounterSchema []CounterSchema
}

type CounterSchema struct {
	Name    string
	Index   uint32
	Type    uint8
	DimX    uint32
	DimY    uint32
	LabelsX []string
	LabelsY []string
}

type ObjectData struct {
	Name      string
	Instances []ObjectInstance
}

type ObjectInstance struct {
	Name     string
	UUID     string
	Counters []CounterType
}

type CounterKind uint8

const (
	_ CounterKind = iota
	CounterKindString
	CounterKindUint32
	CounterKindUint64
	CounterKindList32
	CounterKindList64
	CounterKindListString
)

type CounterType struct {
	Index  uint32
	Type   CounterKind
	_      [3]byte // explicit byte-alignment padding
	scalar uint64
	ref    any
}

func (c *CounterType) setString(value string) {
	c.Type = CounterKindString
	c.scalar = 0
	c.ref = value
}

func (c *CounterType) setUint32(value uint32) {
	c.Type = CounterKindUint32
	c.scalar = uint64(value)
	c.ref = nil
}

func (c *CounterType) setUint64(value uint64) {
	c.Type = CounterKindUint64
	c.scalar = value
	c.ref = nil
}

func (c *CounterType) setList32(value []uint32) {
	c.Type = CounterKindList32
	c.scalar = 0
	c.ref = value
}

func (c *CounterType) setList64(value []uint64) {
	c.Type = CounterKindList64
	c.scalar = 0
	c.ref = value
}

func (c *CounterType) setListString(value []string) {
	c.Type = CounterKindListString
	c.scalar = 0
	c.ref = value
}

func (c *CounterType) IsString() bool {
	return c.Type == CounterKindString
}

func (c *CounterType) IsUint32() bool {
	return c.Type == CounterKindUint32
}

func (c *CounterType) IsUint64() bool {
	return c.Type == CounterKindUint64
}

func (c *CounterType) IsList32() bool {
	return c.Type == CounterKindList32
}

func (c *CounterType) IsList64() bool {
	return c.Type == CounterKindList64
}

func (c *CounterType) IsListString() bool {
	return c.Type == CounterKindListString
}

func (c *CounterType) StringValue() (string, bool) {
	if c.Type != CounterKindString {
		return "", false
	}
	value, ok := c.ref.(string)
	return value, ok
}

func (c *CounterType) Uint32Value() (uint32, bool) {
	if c.Type != CounterKindUint32 {
		return 0, false
	}
	if c.scalar > math.MaxUint32 {
		return 0, false
	}
	return uint32(c.scalar), true
}

func (c *CounterType) Uint64Value() (uint64, bool) {
	if c.Type != CounterKindUint64 {
		return 0, false
	}
	return c.scalar, true
}

func (c *CounterType) List32() ([]uint32, bool) {
	if c.Type != CounterKindList32 {
		return nil, false
	}
	value, ok := c.ref.([]uint32)
	return value, ok
}

func (c *CounterType) List64() ([]uint64, bool) {
	if c.Type != CounterKindList64 {
		return nil, false
	}
	value, ok := c.ref.([]uint64)
	return value, ok
}

func (c *CounterType) ListString() ([]string, bool) {
	if c.Type != CounterKindListString {
		return nil, false
	}
	value, ok := c.ref.([]string)
	return value, ok
}

func Messages(path string) iter.Seq2[*ObjectCollection, error] {
	return func(yield func(*ObjectCollection, error) bool) {
		f, err := os.Open(path)

		if err != nil {
			yield(nil, err)
			return
		}
		defer f.Close()

		br := bufio.NewReaderSize(f, 64*1024)
		buf := make([]byte, 0, 4096)

		for {
			msgLen, err := binary.ReadUvarint(br)
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				yield(nil, fmt.Errorf("read varint: %w", err))
				return
			}
			if uint64(cap(buf)) < msgLen {
				buf = make([]byte, msgLen)
			}
			buf = buf[:msgLen]
			if _, err := io.ReadFull(br, buf); err != nil {
				yield(nil, err)
				return
			}

			msgCopy := bytes.Clone(buf)
			cm, err := readProto(msgCopy)
			if err != nil {
				yield(nil, err)
				return
			}

			if !yield(cm, nil) {
				return // caller broke out of the loop
			}
		}
	}
}

func readProto(src []byte) (*ObjectCollection, error) {
	var (
		fc  easyproto.FieldContext
		err error
	)

	cm := ObjectCollection{}

	for len(src) > 0 {
		src, err = fc.NextField(src)
		if err != nil {
			return nil, err
		}
		switch fc.FieldNum {
		case 1:
			value, ok := fc.Uint64()
			if !ok {
				return nil, errors.New("failed to read timestamp")
			}
			cm.Timestamp = value
		case 2:
			period, ok := fc.Uint32()
			if !ok {
				return nil, errors.New("failed to read sample period")
			}
			cm.Period = period
		case 3:
			nodeName, ok := fc.String()
			if !ok {
				return nil, errors.New("failed to read node name")
			}
			cm.Node = nodeName
		case 5:
			data, ok := fc.MessageData()
			if !ok {
				return nil, errors.New("failed to read object schema message")
			}
			schema, err := handleObjectSchema(data)
			if err != nil {
				return nil, err
			}
			cm.Schema = schema
		case 6:
			data, ok := fc.MessageData()
			if !ok {
				return nil, errors.New("failed to read object data message")
			}
			objectData, err := handleObjectData(data)
			if err != nil {
				return nil, err
			}
			cm.Data = objectData
		}
	}
	return &cm, nil
}

func handleObjectSchema(data []byte) (ObjectSchema, error) {
	var fc easyproto.FieldContext
	objectSchema := ObjectSchema{}

	for len(data) > 0 {
		var err error
		data, err = fc.NextField(data)
		if err != nil {
			return objectSchema, err
		}
		switch fc.FieldNum {
		case 1:
			name, ok := fc.String()
			if !ok {
				return objectSchema, errors.New("failed to read object schema name")
			}
			objectSchema.Name = name
		case 2:
			value, ok := fc.MessageData()
			if !ok {
				return objectSchema, errors.New("failed to read counter schema")
			}
			cs, err := handleCounterSchema(value)
			if err != nil {
				return ObjectSchema{}, err
			}
			objectSchema.CounterSchema = append(objectSchema.CounterSchema, cs)
		}
	}

	return objectSchema, nil
}

func handleCounterSchema(value []byte) (CounterSchema, error) {
	var fc easyproto.FieldContext
	counterSchema := CounterSchema{}

	for len(value) > 0 {
		var err error

		value, err = fc.NextField(value)
		if err != nil {
			return counterSchema, err
		}
		switch fc.FieldNum {
		case 1:
			val, ok := fc.String()
			if !ok {
				return counterSchema, errors.New("failed to read counter schema name")
			}
			counterSchema.Name = val
		case 2:
			val, ok := fc.Uint32()
			if !ok {
				return counterSchema, errors.New("failed to read counter schema index")
			}
			counterSchema.Index = val
		case 3:
			val, ok := fc.Uint64()
			if !ok {
				return counterSchema, errors.New("failed to read counter schema type")
			}
			if val > math.MaxUint8 {
				return counterSchema, fmt.Errorf("counter schema type %d exceeds uint8", val)
			}
			counterSchema.Type = uint8(val)
		case 4:
			val, ok := fc.Uint32()
			if !ok {
				return counterSchema, errors.New("failed to read counter schema dim_x")
			}
			counterSchema.DimX = val
		case 5:
			val, ok := fc.Uint32()
			if !ok {
				return counterSchema, errors.New("failed to read counter schema dim_y")
			}
			counterSchema.DimY = val
		case 6:
			return counterSchema, errors.New("counter_x_labels not implemented yet")
		case 7:
			return counterSchema, errors.New("counter_y_labels not implemented yet")
		}
	}

	return counterSchema, nil
}

func handleObjectData(data []byte) (ObjectData, error) {
	var fc easyproto.FieldContext
	objectData := ObjectData{}

	for len(data) > 0 {
		var err error
		data, err = fc.NextField(data)
		if err != nil {
			return objectData, err
		}
		switch fc.FieldNum {
		case 1:
			key, ok := fc.String()
			if !ok {
				return objectData, errors.New("failed to read object name")
			}
			objectData.Name = key
		case 2:
			value, ok := fc.MessageData()
			if !ok {
				return objectData, errors.New("failed to read object instances")
			}
			instance, err := handleObjectInstance(value)
			if err != nil {
				return ObjectData{}, err
			}
			objectData.Instances = append(objectData.Instances, instance)
		}
	}

	return objectData, nil
}

func handleObjectInstance(value []byte) (ObjectInstance, error) {
	var fc easyproto.FieldContext
	objectInstance := ObjectInstance{}

	for len(value) > 0 {
		var err error
		value, err = fc.NextField(value)
		if err != nil {
			return objectInstance, err
		}
		switch fc.FieldNum {
		case 1:
			val, ok := fc.String()
			if !ok {
				return objectInstance, errors.New("failed to read instance_name value")
			}
			objectInstance.Name = val
		case 2:
			val, ok := fc.String()
			if !ok {
				return objectInstance, errors.New("failed to read instance_uuid value")
			}
			objectInstance.UUID = val
		case 3:
			data, ok := fc.MessageData()
			if !ok {
				return objectInstance, errors.New("failed to read instance counters")
			}
			counterType, err := handleCounterType(data)
			if err != nil {
				return objectInstance, err
			}
			objectInstance.Counters = append(objectInstance.Counters, counterType)
		}
	}

	return objectInstance, nil
}

func handleCounterType(data []byte) (CounterType, error) {
	var fc easyproto.FieldContext
	counterType := CounterType{}

	for len(data) > 0 {
		var err error
		data, err = fc.NextField(data)
		if err != nil {
			return counterType, err
		}
		switch fc.FieldNum {
		case 1:
			val, ok := fc.Uint32()
			if !ok {
				return counterType, errors.New("failed to read counter_index")
			}
			counterType.Index = val
		case 2:
			val, ok := fc.String()
			if !ok {
				return counterType, errors.New("failed to read string_value")
			}
			counterType.setString(val)
		case 3:
			val, ok := fc.Uint32()
			if !ok {
				return counterType, errors.New("failed to read uint32_value")
			}
			counterType.setUint32(val)
		case 4:
			val, ok := fc.Uint64()
			if !ok {
				return counterType, errors.New("failed to read uint64_value")
			}
			counterType.setUint64(val)
		case 5:
			val, ok := fc.MessageData()
			if !ok {
				return counterType, errors.New("failed to read array of uint32 values")
			}
			counters, err := handleArrayCounter32(val)
			if err != nil {
				return counterType, err
			}
			counterType.setList32(counters)
		case 6:
			val, ok := fc.MessageData()
			if !ok {
				return counterType, errors.New("failed to read array of uint64 values")
			}
			counters, err := handleArrayCounter64(val)
			if err != nil {
				return counterType, err
			}
			counterType.setList64(counters)
		case 7:
			val, ok := fc.MessageData()
			if !ok {
				return counterType, errors.New("failed to read array of string values")
			}
			counters, err := handleArrayCounterString(val)
			if err != nil {
				return counterType, err
			}
			counterType.setListString(counters)
		}
	}

	return counterType, nil
}

func handleArrayCounterString(val []byte) ([]string, error) {
	var fc easyproto.FieldContext
	var counters []string

	for len(val) > 0 {
		var err error
		val, err = fc.NextField(val)
		if err != nil {
			return nil, errors.New("cannot read sample array string data")
		}
		if fc.FieldNum != 1 {
			continue
		}

		s, ok := fc.String()
		if !ok {
			return nil, errors.New("cannot read sample array string value")
		}
		counters = append(counters, s)
	}

	return counters, nil
}

func handleArrayCounter32(val []byte) ([]uint32, error) {
	var fc easyproto.FieldContext
	var counters []uint32
	for len(val) > 0 {
		var err error
		val, err = fc.NextField(val)
		if err != nil {
			return nil, errors.New("cannot read sample array counter32 data")
		}
		if fc.FieldNum != 1 {
			continue
		}
		var ok bool
		counters, ok = fc.UnpackUint32s(counters)
		if !ok {
			return nil, errors.New("cannot read sample array counter32 values")
		}
	}

	return counters, nil
}

func handleArrayCounter64(val []byte) ([]uint64, error) {
	var fc easyproto.FieldContext
	var counters []uint64
	for len(val) > 0 {
		var err error
		val, err = fc.NextField(val)
		if err != nil {
			return nil, errors.New("cannot read sample array counter64 data")
		}
		if fc.FieldNum != 1 {
			continue
		}
		var ok bool
		counters, ok = fc.UnpackUint64s(counters)
		if !ok {
			return nil, errors.New("cannot read sample array counter64 values")
		}
	}

	return counters, nil
}
