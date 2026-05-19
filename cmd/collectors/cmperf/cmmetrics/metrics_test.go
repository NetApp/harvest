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

func appendStringField(dst []byte, fieldNum uint64, value string) []byte {
	tag := (fieldNum << 3) | 2 // wire type 2 for length-delimited
	dst = binary.AppendUvarint(dst, tag)
	dst = binary.AppendUvarint(dst, uint64(len(value)))
	dst = append(dst, value...)
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

func mustList64(t *testing.T, counter CounterType) []uint64 {
	t.Helper()
	value, ok := counter.List64()
	if !ok {
		t.Fatalf("expected []uint64 value for counter type %v", counter.Type)
	}
	return value
}

func TestMessages(t *testing.T) {

	path := "testdata/wqd.pb"
	obs := make([]*MetricsFileRecord, 0, 5)

	for aMsg, err := range Messages(path) {
		assert.Nil(t, err)
		assert.NotNil(t, aMsg)
		obs = append(obs, aMsg)
	}

	assert.Equal(t, len(obs), 5)

	version := obs[0]
	schema := obs[1]
	rec1 := obs[2].batch
	rec2 := obs[3].batch
	summary := obs[4]

	assert.Equal(t, version.version.FormatVersion, 1)
	assert.Equal(t, schema.schema.Name, "workload_queue_dblade")
	assert.Equal(t, schema.schema.CounterSchema[0].Name, "instance_name")
	assert.Equal(t, schema.schema.CounterSchema[15].Name, "cache_miss_rate")
	assert.Equal(t, schema.schema.CounterSchema[15].BaseIndex, 16)

	assert.Equal(t, rec1.Timestamp, 1779210300000)
	assert.Equal(t, rec2.Timestamp, 1779210300000)
	assert.Equal(t, rec1.Period, 60)
	assert.Equal(t, rec1.Data.Name, "workload_queue_dblade")
	assert.Equal(t, rec1.Data.Instances[0].Name, "<none>")
	assert.Equal(t, rec1.Data.Instances[0].UUID, "<none>")
	assert.True(t, rec1.Data.Instances[0].Counters[13].IsUint64())
	assert.Equal(t, mustUint64Value(t, rec1.Data.Instances[0].Counters[13]), 5957)
	assert.Equal(t, mustList64(t, rec1.Data.Instances[0].Counters[4]), []uint64{9190949671, 0, 0, 0})

	assert.Equal(t, len(summary.summary.Statuses), 2)
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

func TestHandleCounterSchemaLabels(t *testing.T) {
	payload := appendStringField(nil, 1, "counter-1")
	payload = appendVarintField(payload, 2, 1)
	payload = appendVarintField(payload, 3, 3)
	payload = appendStringField(payload, 6, "node")
	payload = appendStringField(payload, 6, "svm")
	payload = appendStringField(payload, 7, "read")
	payload = appendStringField(payload, 7, "write")

	schema, err := handleCounterSchema(payload)
	assert.Nil(t, err)
	assert.Equal(t, schema.LabelsX, []string{"node", "svm"})
	assert.Equal(t, schema.LabelsY, []string{"read", "write"})
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
