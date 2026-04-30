package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint64Binder_Mappable(t *testing.T) {
	m := Uint64Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{uint64(0), false},
		{new(uint64), true},
		{[]uint64{}, false},
		{new([]uint64), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestUint64Binder_Bind(t *testing.T) {
	m := Uint64Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst uint64
		err := m.Bind("123456789012345", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint64(123456789012345), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("12345", &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("12345", (*uint64)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestUint64Binder_BindMany(t *testing.T) {
	m := Uint64Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []uint64
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []uint64{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})
}

func TestUint64Binder_BindT(t *testing.T) {
	m := Uint64Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *uint64
		wantErr  bool
		errMsg   string
		expected uint64
	}{
		{"Success", "18446744073709551615", new(uint64), false, "", 18446744073709551615},
		{"Parse Error", "abc", new(uint64), true, "failed to bind value to uint64", 0},
		{"Overflow Error", "18446744073709551616", new(uint64), true, "failed to bind value to uint64", 0},
		{"Nil Destination", "123", nil, true, "destination cannot be nil", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.BindT(tt.src, tt.dst)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, *tt.dst)
			}
		})
	}
}

func TestUint64Binder_BindManyT(t *testing.T) {
	m := Uint64Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]uint64
		wantErr  bool
		errMsg   string
		expected []uint64
	}{
		{"Success", []string{"1", "2"}, new([]uint64), false, "", []uint64{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]uint64), true, "failed to bind value to uint64", nil},
		{"Overflow Error", []string{"1", "18446744073709551616"}, new([]uint64), true, "failed to bind value to uint64", nil},
		{"Nil Destination", []string{"1"}, nil, true, "destination cannot be nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.BindManyT(tt.src, tt.dst)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, *tt.dst)
			}
		})
	}
}
