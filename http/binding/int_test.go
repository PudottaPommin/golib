package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntBinder_Mappable(t *testing.T) {
	m := IntBinder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{int(0), true},
		{new(int), true},
		{[]int{}, true},
		{new([]int), true},
		{int8(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestIntBinder_Bind(t *testing.T) {
	m := IntBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst int
		err := m.Bind("1234567", &dst)
		assert.NoError(t, err)
		assert.Equal(t, 1234567, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int8
		err := m.Bind("1234567", &dst)
		assert.Error(t, err)
		assert.Equal(t, "destination is not mappable", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("1234567", (*int)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestIntBinder_BindMany(t *testing.T) {
	m := IntBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int8
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "destination is not mappable", err.Error())
	})
}

func TestIntBinder_BindT(t *testing.T) {
	m := IntBinder{}

	tests := []struct {
		name     string
		src      string
		dst      *int
		wantErr  bool
		errMsg   string
		expected int
	}{
		{"Success", "1234567", new(int), false, "", 1234567},
		{"Parse Error", "abc", new(int), true, "failed to bind value to int", 0},
		// Overflow test for int is platform dependent, but 10^20 should overflow both 32 and 64 bit int.
		{"Overflow Error", "100000000000000000000", new(int), true, "failed to bind value to int", 0},
		{"Nil Destination", "1234567", nil, true, "destination cannot be nil", 0},
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

func TestIntBinder_BindManyT(t *testing.T) {
	m := IntBinder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]int
		wantErr  bool
		errMsg   string
		expected []int
	}{
		{"Success", []string{"1", "2"}, new([]int), false, "", []int{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]int), true, "failed to bind value to int", nil},
		{"Overflow Error", []string{"1", "100000000000000000000"}, new([]int), true, "failed to bind value to int", nil},
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
