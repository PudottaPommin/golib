package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint16Binder_Mappable(t *testing.T) {
	m := Uint16Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{uint16(0), false},
		{new(uint16), true},
		{[]uint16{}, false},
		{new([]uint16), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestUint16Binder_Bind(t *testing.T) {
	m := Uint16Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst uint16
		err := m.Bind("12345", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint16(12345), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("12345", &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("12345", (*uint16)(nil))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationNil)
	})
}

func TestUint16Binder_BindMany(t *testing.T) {
	m := Uint16Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []uint16
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []uint16{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})
}

func TestUint16Binder_BindT(t *testing.T) {
	m := Uint16Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *uint16
		wantErr  bool
		errMsg   string
		expected uint16
	}{
		{"Success", "65535", new(uint16), false, "", 65535},
		{"Parse Error", "abc", new(uint16), true, "failed to bind value to uint16", 0},
		{"Overflow Error", "65536", new(uint16), true, "failed to bind value to uint16", 0},
		{"Nil Destination", "65535", nil, true, "destination cannot be nil", 0},
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

func TestUint16Binder_BindManyT(t *testing.T) {
	m := Uint16Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]uint16
		wantErr  bool
		errMsg   string
		expected []uint16
	}{
		{"Success", []string{"1", "2"}, new([]uint16), false, "", []uint16{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]uint16), true, "failed to bind value to uint16", nil},
		{"Overflow Error", []string{"1", "65536"}, new([]uint16), true, "failed to bind value to uint16", nil},
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
