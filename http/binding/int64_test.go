package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt64Binder_Mappable(t *testing.T) {
	m := Int64Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{int64(0), false},
		{new(int64), true},
		{[]int64{}, false},
		{new([]int64), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestInt64Binder_Bind(t *testing.T) {
	m := Int64Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst int64
		err := m.Bind("123456789012345", &dst)
		assert.NoError(t, err)
		assert.Equal(t, int64(123456789012345), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("123456789012345", &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("123456789012345", (*int64)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestInt64Binder_BindMany(t *testing.T) {
	m := Int64Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []int64
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int64{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})
}

func TestInt64Binder_BindT(t *testing.T) {
	m := Int64Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *int64
		wantErr  bool
		errMsg   string
		expected int64
	}{
		{"Success", "9223372036854775807", new(int64), false, "", 9223372036854775807},
		{"Parse Error", "abc", new(int64), true, "failed to bind value to int64", 0},
		{"Overflow Error", "9223372036854775808", new(int64), true, "failed to bind value to int64", 0},
		{"Nil Destination", "9223372036854775807", nil, true, "destination cannot be nil", 0},
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

func TestInt64Binder_BindManyT(t *testing.T) {
	m := Int64Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]int64
		wantErr  bool
		errMsg   string
		expected []int64
	}{
		{"Success", []string{"1", "2"}, new([]int64), false, "", []int64{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]int64), true, "failed to bind value to int64", nil},
		{"Overflow Error", []string{"1", "9223372036854775808"}, new([]int64), true, "failed to bind value to int64", nil},
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
