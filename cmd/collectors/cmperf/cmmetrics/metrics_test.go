package cmmetrics

import (
	"encoding/binary"
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

func appendRepeatedUintField(dst []byte, fieldNum uint64, values ...uint64) []byte {
	tag := fieldNum << 3 // make room for the wire type in the lower 3 bits
	for _, value := range values {
		dst = binary.AppendUvarint(dst, tag)
		dst = binary.AppendUvarint(dst, value)
	}
	return dst
}

func appendVarintField(dst []byte, fieldNum uint64, value uint64) []byte {
	tag := fieldNum << 3 // make room for the wire type in the lower 3 bits
	dst = binary.AppendUvarint(dst, tag)
	dst = binary.AppendUvarint(dst, value)
	return dst
}

func mustStringValue(t *testing.T, counter CounterType) string {
	t.Helper()
	value, ok := counter.StringValue()
	if !ok {
		t.Fatalf("expected string value for counter type %v", counter.Type)
	}
	return value
}

func mustUint32Value(t *testing.T, counter CounterType) uint32 {
	t.Helper()
	value, ok := counter.Uint32Value()
	if !ok {
		t.Fatalf("expected uint32 value for counter type %v", counter.Type)
	}
	return value
}

func mustUint64Value(t *testing.T, counter CounterType) uint64 {
	t.Helper()
	value, ok := counter.Uint64Value()
	if !ok {
		t.Fatalf("expected uint64 value for counter type %v", counter.Type)
	}
	return value
}

func mustList32(t *testing.T, counter CounterType) []uint32 {
	t.Helper()
	value, ok := counter.List32()
	if !ok {
		t.Fatalf("expected []uint32 value for counter type %v", counter.Type)
	}
	return value
}

func TestMessages(t *testing.T) {

	path := "testdata/test1.pb"
	obs := make([]*ObjectCollection, 0, 2)

	for aMsg, err := range Messages(path) {
		assert.Nil(t, err)
		assert.NotNil(t, aMsg)
		obs = append(obs, aMsg)
	}

	assert.Equal(t, len(obs), 2)

	assert.Equal(t, obs[0].Timestamp, uint64(1776429464001))
	assert.Equal(t, obs[0].Period, uint32(60))
	assert.Equal(t, obs[0].Node, "cm-test")
	assert.Equal(t, obs[0].Schema.Name, "cm-test")
	assert.Equal(t, len(obs[0].Schema.CounterSchema), 2)
	assert.Equal(t, obs[0].Schema.CounterSchema[0].Name, "counter-1")
	assert.Equal(t, obs[0].Schema.CounterSchema[0].Index, uint32(1))
	assert.Equal(t, obs[0].Schema.CounterSchema[0].Type, uint8(3))
	assert.Equal(t, obs[0].Schema.CounterSchema[0].DimX, uint32(0))
	assert.Equal(t, obs[0].Schema.CounterSchema[0].DimY, uint32(0))
	assert.Equal(t, len(obs[0].Schema.CounterSchema[0].LabelsX), 0)
	assert.Equal(t, len(obs[0].Schema.CounterSchema[0].LabelsY), 0)

	assert.Equal(t, obs[0].Data.Name, "od-1")
	assert.Equal(t, len(obs[0].Data.Instances), 2)
	assert.Equal(t, obs[0].Data.Instances[0].Name, "vol1")
	assert.Equal(t, obs[0].Data.Instances[0].UUID, "06b3c803-ff78-11eb-ba17-00a098e24321")
	assert.Equal(t, len(obs[0].Data.Instances[0].Counters), 4)
	assert.Equal(t, obs[0].Data.Instances[0].Counters[0].Index, uint32(1))
	assert.Equal(t, mustUint64Value(t, obs[0].Data.Instances[0].Counters[0]), 42)
	assert.False(t, obs[0].Data.Instances[0].Counters[0].IsUint32())
	assert.True(t, obs[0].Data.Instances[0].Counters[0].IsUint64())
	assert.False(t, obs[0].Data.Instances[0].Counters[0].IsList32())
	assert.False(t, obs[0].Data.Instances[0].Counters[0].IsList64())
	assert.False(t, obs[0].Data.Instances[0].Counters[0].IsListString())

	assert.Equal(t, obs[0].Data.Instances[1].Counters[3].Index, 4)
	assert.Equal(t, obs[0].Data.Instances[1].Counters[2].scalar, 144)

	listString, b := obs[1].Data.Instances[0].Counters[3].ListString()
	assert.True(t, b)
	assert.Equal(t, listString, []string{"abc", "def", "ghi"})
	assert.Equal(t, mustList32(t, obs[1].Data.Instances[0].Counters[4]), []uint32{100, 200, 300})
}

func TestCounterTypeAccessors(t *testing.T) {
	var counter CounterType

	counter.setUint32(7)
	assert.Equal(t, counter.IsUint32(), true)
	assert.Equal(t, mustUint32Value(t, counter), uint32(7))
	_, ok := counter.StringValue()
	assert.Equal(t, ok, false)

	counter.setUint64(11)
	assert.Equal(t, counter.IsUint64(), true)
	assert.Equal(t, mustUint64Value(t, counter), uint64(11))
	_, ok = counter.List32()
	assert.Equal(t, ok, false)

	counter.setString("value")
	assert.Equal(t, counter.IsString(), true)
	assert.Equal(t, mustStringValue(t, counter), "value")

	counter.setList32([]uint32{1, 2, 3})
	assert.Equal(t, counter.IsList32(), true)
	assert.Equal(t, mustList32(t, counter), []uint32{1, 2, 3})

	counter.setList64([]uint64{4, 5})
	assert.Equal(t, counter.IsList64(), true)
	value64, ok := counter.List64()
	assert.Equal(t, ok, true)
	assert.Equal(t, value64, []uint64{4, 5})

	counter.setListString([]string{"a", "b"})
	assert.Equal(t, counter.IsListString(), true)
	valueString, ok := counter.ListString()
	assert.Equal(t, ok, true)
	assert.Equal(t, valueString, []string{"a", "b"})
}

func TestHandleArrayCountersUnpackedRepeatedFields(t *testing.T) {
	t.Run("uint32", func(t *testing.T) {
		payload := appendRepeatedUintField(nil, 1, 100, 200, 300)
		counters, err := handleArrayCounter32(payload)
		assert.Nil(t, err)
		assert.Equal(t, counters, []uint32{100, 200, 300})
	})

	t.Run("uint64", func(t *testing.T) {
		payload := appendRepeatedUintField(nil, 1, 1, 1<<33, 1<<34)
		counters, err := handleArrayCounter64(payload)
		assert.Nil(t, err)
		assert.Equal(t, counters, []uint64{1, 1 << 33, 1 << 34})
	})
}

func TestReadProtoRejectsMalformedNestedMessages(t *testing.T) {
	t.Run("schema", func(t *testing.T) {
		payload := appendVarintField(nil, 5, 1)
		msg, err := readProto(payload)
		assert.Nil(t, msg)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "failed to read object schema message")
	})

	t.Run("data", func(t *testing.T) {
		payload := appendVarintField(nil, 6, 1)
		msg, err := readProto(payload)
		assert.Nil(t, msg)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "failed to read object data message")
	})
}

func BenchmarkCounterType(b *testing.B) {
	b.Run("uint32", func(b *testing.B) {
		b.ReportAllocs()
		var total uint64
		value := uint32(1)
		for range b.N {
			var counter CounterType
			counter.setUint32(value)
			storedValue, ok := counter.Uint32Value()
			if !ok {
				b.Fatal("expected uint32 value")
			}
			total += uint64(storedValue)
			value++
		}
		if total == 0 {
			b.Fatal("unexpected zero total")
		}
	})

	b.Run("string", func(b *testing.B) {
		b.ReportAllocs()
		value := "value"
		var total int
		for range b.N {
			var counter CounterType
			counter.setString(value)
			storedValue, ok := counter.StringValue()
			if !ok {
				b.Fatal("expected string value")
			}
			total += len(storedValue)
		}
		if total == 0 {
			b.Fatal("unexpected zero total")
		}
	})

	b.Run("list32", func(b *testing.B) {
		b.ReportAllocs()
		value := []uint32{1, 2, 3, 4}
		var total int
		for range b.N {
			var counter CounterType
			counter.setList32(value)
			storedValue, ok := counter.List32()
			if !ok {
				b.Fatal("expected []uint32 value")
			}
			total += len(storedValue)
		}
		if total == 0 {
			b.Fatal("unexpected zero total")
		}
	})
}
