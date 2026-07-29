package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt32Binder_Mappable(t *testing.T) {
	m := Int32Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{int32(0), false},
		{new(int32), true},
		{[]int32{}, false},
		{new([]int32), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestInt32Binder_Bind(t *testing.T) {
	m := Int32Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst int32
		err := m.Bind("1234567", &dst)
		assert.NoError(t, err)
		assert.Equal(t, int32(1234567), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("1234567", &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("1234567", (*int32)(nil))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationNil)
	})
}

func TestInt32Binder_BindMany(t *testing.T) {
	m := Int32Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []int32
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int32{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})
}

func TestInt32Binder_BindT(t *testing.T) {
	m := Int32Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *int32
		wantErr  bool
		errMsg   string
		expected int32
	}{
		{"Success", "2147483647", new(int32), false, "", 2147483647},
		{"Parse Error", "abc", new(int32), true, "failed to bind value to int32", 0},
		{"Overflow Error", "2147483648", new(int32), true, "failed to bind value to int32", 0},
		{"Nil Destination", "2147483647", nil, true, "destination cannot be nil", 0},
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

func TestInt32Binder_BindManyT(t *testing.T) {
	m := Int32Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]int32
		wantErr  bool
		errMsg   string
		expected []int32
	}{
		{"Success", []string{"1", "2"}, new([]int32), false, "", []int32{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]int32), true, "failed to bind value to int32", nil},
		{"Overflow Error", []string{"1", "2147483648"}, new([]int32), true, "failed to bind value to int32", nil},
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
