package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt8Binder_Mappable(t *testing.T) {
	m := Int8Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{int8(0), false},
		{new(int8), true},
		{[]int8{}, false},
		{new([]int8), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestInt8Binder_Bind(t *testing.T) {
	m := Int8Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst int8
		err := m.Bind("123", &dst)
		assert.NoError(t, err)
		assert.Equal(t, int8(123), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("123", &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("123", (*int8)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestInt8Binder_BindMany(t *testing.T) {
	m := Int8Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []int8
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int8{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})
}

func TestInt8Binder_BindT(t *testing.T) {
	m := Int8Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *int8
		wantErr  bool
		errMsg   string
		expected int8
	}{
		{"Success", "127", new(int8), false, "", 127},
		{"Parse Error", "abc", new(int8), true, "failed to bind value to int8", 0},
		{"Overflow Error", "128", new(int8), true, "failed to bind value to int8", 0},
		{"Nil Destination", "127", nil, true, "destination cannot be nil", 0},
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

func TestInt8Binder_BindManyT(t *testing.T) {
	m := Int8Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]int8
		wantErr  bool
		errMsg   string
		expected []int8
	}{
		{"Success", []string{"1", "2"}, new([]int8), false, "", []int8{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]int8), true, "failed to bind value to int8", nil},
		{"Overflow Error", []string{"1", "128"}, new([]int8), true, "failed to bind value to int8", nil},
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
