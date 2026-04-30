package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint8Binder_Mappable(t *testing.T) {
	m := Uint8Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{uint8(0), false},
		{new(uint8), true},
		{[]uint8{}, false},
		{new([]uint8), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestUint8Binder_Bind(t *testing.T) {
	m := Uint8Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst uint8
		err := m.Bind("123", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint8(123), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("123", &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("123", (*uint8)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestUint8Binder_BindMany(t *testing.T) {
	m := Uint8Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []uint8
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []uint8{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})
}

func TestUint8Binder_BindT(t *testing.T) {
	m := Uint8Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *uint8
		wantErr  bool
		errMsg   string
		expected uint8
	}{
		{"Success", "255", new(uint8), false, "", 255},
		{"Parse Error", "abc", new(uint8), true, "failed to bind value to uint8", 0},
		{"Overflow Error", "256", new(uint8), true, "failed to bind value to uint8", 0},
		{"Nil Destination", "255", nil, true, "destination cannot be nil", 0},
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

func TestUint8Binder_BindManyT(t *testing.T) {
	m := Uint8Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]uint8
		wantErr  bool
		errMsg   string
		expected []uint8
	}{
		{"Success", []string{"1", "2"}, new([]uint8), false, "", []uint8{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]uint8), true, "failed to bind value to uint8", nil},
		{"Overflow Error", []string{"1", "256"}, new([]uint8), true, "failed to bind value to uint8", nil},
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
