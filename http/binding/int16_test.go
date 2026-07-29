package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt16Binder_Mappable(t *testing.T) {
	m := Int16Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{int16(0), false},
		{new(int16), true},
		{[]int16{}, false},
		{new([]int16), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestInt16Binder_Bind(t *testing.T) {
	m := Int16Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst int16
		err := m.Bind("12345", &dst)
		assert.NoError(t, err)
		assert.Equal(t, int16(12345), dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("12345", &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("12345", (*int16)(nil))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationNil)
	})
}

func TestInt16Binder_BindMany(t *testing.T) {
	m := Int16Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []int16
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []int16{1, 2}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1", "2"}, &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})
}

func TestInt16Binder_BindT(t *testing.T) {
	m := Int16Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *int16
		wantErr  bool
		errMsg   string
		expected int16
	}{
		{"Success", "32767", new(int16), false, "", 32767},
		{"Parse Error", "abc", new(int16), true, "failed to bind value to int16", 0},
		{"Overflow Error", "32768", new(int16), true, "failed to bind value to int16", 0},
		{"Nil Destination", "32767", nil, true, "destination cannot be nil", 0},
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

func TestInt16Binder_BindManyT(t *testing.T) {
	m := Int16Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]int16
		wantErr  bool
		errMsg   string
		expected []int16
	}{
		{"Success", []string{"1", "2"}, new([]int16), false, "", []int16{1, 2}},
		{"Parse Error", []string{"1", "abc"}, new([]int16), true, "failed to bind value to int16", nil},
		{"Overflow Error", []string{"1", "32768"}, new([]int16), true, "failed to bind value to int16", nil},
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
