package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUintBinder_Mappable(t *testing.T) {
	m := UintBinder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{uint(0), false},
		{new(uint), true},
		{[]uint{}, false},
		{new([]uint), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestUintBinder_Bind(t *testing.T) {
	m := UintBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst uint
		err := m.Bind("12345", &dst)
		assert.NoError(t, err)
		assert.Equal(t, uint(12345), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("12345", &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("12345", (*uint)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestUintBinder_BindMany(t *testing.T) {
	m := UintBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst []uint
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []uint{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})
}

func TestUintBinder_BindT(t *testing.T) {
	m := UintBinder{}

	tests := []struct {
		name     string
		src      string
		dst      *uint
		wantErr  bool
		errMsg   string
		expected uint
	}{
		{"Success", "12345", new(uint), false, "", 12345},
		{"Parse Error", "abc", new(uint), true, "failed to bind value to uint", 0},
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

func TestUintBinder_BindManyT(t *testing.T) {
	m := UintBinder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]uint
		wantErr  bool
		errMsg   string
		expected []uint
	}{
		{"Success", []string{"1", "2"}, new([]uint), false, "", []uint{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]uint), true, "failed to bind value to uint", nil},
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
