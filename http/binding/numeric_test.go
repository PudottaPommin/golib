package binding_test

import (
	"testing"

	. "github.com/pudottapommin/golib/http/binding"
	"github.com/stretchr/testify/assert"
)

func TestNumericBinder_Mappable(t *testing.T) {
	t.Run("Int", func(t *testing.T) {
		m := WrapBinder(IntBinder{})
		assert.True(t, m.Mappable(new(int)))
		assert.True(t, m.Mappable(new([]int)))
		assert.False(t, m.Mappable(int(0)))
		assert.False(t, m.Mappable(new(string)))
		assert.False(t, m.Mappable(nil))
	})

	t.Run("Int8", func(t *testing.T) {
		m := WrapBinder(Int8Binder{})
		assert.True(t, m.Mappable(new(int8)))
		assert.True(t, m.Mappable(new([]int8)))
		assert.False(t, m.Mappable(new(int)))
	})

	t.Run("Uint", func(t *testing.T) {
		m := WrapBinder(UintBinder{})
		assert.True(t, m.Mappable(new(uint)))
		assert.True(t, m.Mappable(new([]uint)))
		assert.False(t, m.Mappable(new(int)))
	})

	t.Run("Float64", func(t *testing.T) {
		m := WrapBinder(Float64Binder{})
		assert.True(t, m.Mappable(new(float64)))
		assert.True(t, m.Mappable(new([]float64)))
		assert.False(t, m.Mappable(new(int)))
	})
}

func TestNumericBinder_Bind(t *testing.T) {
	t.Run("Int Success", func(t *testing.T) {
		var dst int
		m := IntBinder{}
		err := m.Bind("12345", &dst)
		assert.NoError(t, err)
		assert.Equal(t, 12345, dst)
	})

	t.Run("Int8 Success & Overflow", func(t *testing.T) {
		var dst int8
		m := Int8Binder{}
		assert.NoError(t, m.Bind("127", &dst))
		assert.Equal(t, int8(127), dst)
		assert.Error(t, m.Bind("128", &dst)) // overflow for int8
	})

	t.Run("Int16 Success & Overflow", func(t *testing.T) {
		var dst int16
		m := Int16Binder{}
		assert.NoError(t, m.Bind("32767", &dst))
		assert.Equal(t, int16(32767), dst)
		assert.Error(t, m.Bind("32768", &dst))
	})

	t.Run("Int32 Success & Overflow", func(t *testing.T) {
		var dst int32
		m := Int32Binder{}
		assert.NoError(t, m.Bind("2147483647", &dst))
		assert.Equal(t, int32(2147483647), dst)
		assert.Error(t, m.Bind("2147483648", &dst))
	})

	t.Run("Int64 Success", func(t *testing.T) {
		var dst int64
		m := Int64Binder{}
		assert.NoError(t, m.Bind("9223372036854775807", &dst))
		assert.Equal(t, int64(9223372036854775807), dst)
		assert.Error(t, m.Bind("9223372036854775808", &dst))
	})

	t.Run("Uint Success", func(t *testing.T) {
		var dst uint
		m := UintBinder{}
		assert.NoError(t, m.Bind("500", &dst))
		assert.Equal(t, uint(500), dst)
		assert.Error(t, m.Bind("-1", &dst))
	})

	t.Run("Uint8 Success & Overflow", func(t *testing.T) {
		var dst uint8
		m := Uint8Binder{}
		assert.NoError(t, m.Bind("255", &dst))
		assert.Equal(t, uint8(255), dst)
		assert.Error(t, m.Bind("256", &dst))
	})

	t.Run("Uint16 Success & Overflow", func(t *testing.T) {
		var dst uint16
		m := Uint16Binder{}
		assert.NoError(t, m.Bind("65535", &dst))
		assert.Equal(t, uint16(65535), dst)
		assert.Error(t, m.Bind("65536", &dst))
	})

	t.Run("Uint32 Success & Overflow", func(t *testing.T) {
		var dst uint32
		m := Uint32Binder{}
		assert.NoError(t, m.Bind("4294967295", &dst))
		assert.Equal(t, uint32(4294967295), dst)
		assert.Error(t, m.Bind("4294967296", &dst))
	})

	t.Run("Uint64 Success", func(t *testing.T) {
		var dst uint64
		m := Uint64Binder{}
		assert.NoError(t, m.Bind("18446744073709551615", &dst))
		assert.Equal(t, uint64(18446744073709551615), dst)
	})

	t.Run("Float32 Success", func(t *testing.T) {
		var dst float32
		m := Float32Binder{}
		assert.NoError(t, m.Bind("3.14", &dst))
		assert.InDelta(t, float32(3.14), dst, 0.001)
	})

	t.Run("Float64 Success", func(t *testing.T) {
		var dst float64
		m := Float64Binder{}
		assert.NoError(t, m.Bind("2.718281828", &dst))
		assert.InDelta(t, 2.718281828, dst, 0.00001)
	})

	t.Run("Nil Destination", func(t *testing.T) {
		m := IntBinder{}
		assert.ErrorIs(t, m.Bind("123", nil), ErrDestinationNil)
	})

	t.Run("Empty Source", func(t *testing.T) {
		var dst int
		m := IntBinder{}
		assert.ErrorIs(t, m.Bind("", &dst), ErrValueIsZero)
	})
}

func TestNumericBinder_BindMany(t *testing.T) {
	t.Run("Int Slice", func(t *testing.T) {
		var dst []int
		m := IntBinder{}
		err := m.BindMany([]string{"1", "2", "3"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, dst)
	})

	t.Run("Float64 Slice with Capacity Reuse", func(t *testing.T) {
		dst := make([]float64, 5)
		m := Float64Binder{}
		err := m.BindMany([]string{"1.1", "2.2"}, &dst)
		assert.NoError(t, err)
		assert.Len(t, dst, 2)
		assert.InDelta(t, 1.1, dst[0], 0.001)
		assert.InDelta(t, 2.2, dst[1], 0.001)
	})

	t.Run("Parse Error in Slice", func(t *testing.T) {
		var dst []int
		m := IntBinder{}
		err := m.BindMany([]string{"1", "invalid"}, &dst)
		assert.Error(t, err)
	})

	t.Run("Nil Destination", func(t *testing.T) {
		m := IntBinder{}
		assert.ErrorIs(t, m.BindMany([]string{"1"}, nil), ErrDestinationNil)
	})
}

func TestNumericBinder_CustomTypes(t *testing.T) {
	type CustomPort uint16
	type CustomBalance int64

	t.Run("CustomPort Bind", func(t *testing.T) {
		b := NumericBinder[CustomPort]{}
		var port CustomPort
		err := b.Bind("8080", &port)
		assert.NoError(t, err)
		assert.Equal(t, CustomPort(8080), port)
	})

	t.Run("CustomBalance BindMany", func(t *testing.T) {
		b := NumericBinder[CustomBalance]{}
		var balances []CustomBalance
		err := b.BindMany([]string{"1000", "2000"}, &balances)
		assert.NoError(t, err)
		assert.Equal(t, []CustomBalance{1000, 2000}, balances)
	})
}
