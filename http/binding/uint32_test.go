package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint32Binder_Mappable(t *testing.T) {
	m := Uint32Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{uint32(0), false},
		{new(uint32), true},
		{[]uint32{}, false},
		{new([]uint32), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestUint32Binder_Bind(t *testing.T) {
	m := Uint32Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst uint32
		err := m.Bind("12345", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint32(12345), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("12345", &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("12345", (*uint32)(nil))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationNil)
	})
}

func TestUint32Binder_BindMany(t *testing.T) {
	m := Uint32Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []uint32
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []uint32{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})
}

func TestUint32Binder_BindT(t *testing.T) {
	m := Uint32Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *uint32
		wantErr  bool
		errMsg   string
		expected uint32
	}{
		{"Success", "4294967295", new(uint32), false, "", 4294967295},
		{"Parse Error", "abc", new(uint32), true, "failed to bind value to uint32", 0},
		{"Overflow Error", "4294967296", new(uint32), true, "failed to bind value to uint32", 0},
		{"Nil Destination", "4294967295", nil, true, "destination cannot be nil", 0},
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

func TestUint32Binder_BindManyT(t *testing.T) {
	m := Uint32Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]uint32
		wantErr  bool
		errMsg   string
		expected []uint32
	}{
		{"Success", []string{"1", "2"}, new([]uint32), false, "", []uint32{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]uint32), true, "failed to bind value to uint32", nil},
		{"Overflow Error", []string{"1", "4294967296"}, new([]uint32), true, "failed to bind value to uint32", nil},
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
