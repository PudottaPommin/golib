package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolBinder_Mappable(t *testing.T) {
	m := BoolBinder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{bool(true), true},
		{new(bool), true},
		{[]bool{}, true},
		{new([]bool), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestBoolBinder_Bind(t *testing.T) {
	m := BoolBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst bool
		err := m.Bind("true", &dst)
		assert.NoError(t, err)
		assert.True(t, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("true", &dst)
		assert.Error(t, err)
		assert.Equal(t, "destination is not mappable", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("true", (*bool)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestBoolBinder_BindMany(t *testing.T) {
	m := BoolBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst []bool
		err := m.BindMany([]string{"true", "false"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []bool{true, false}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"true", "false"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "destination is not mappable", err.Error())
	})
}

func TestBoolBinder_BindT(t *testing.T) {
	m := BoolBinder{}

	tests := []struct {
		name     string
		src      string
		dst      *bool
		wantErr  bool
		errMsg   string
		expected bool
	}{
		{"Success True", "true", new(bool), false, "", true},
		{"Success False", "false", new(bool), false, "", false},
		{"Parse Error", "not-a-bool", new(bool), true, "failed to bind value to bool", false},
		{"Nil Destination", "true", nil, true, "destination cannot be nil", false},
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

func TestBoolBinder_BindManyT(t *testing.T) {
	m := BoolBinder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]bool
		wantErr  bool
		errMsg   string
		expected []bool
	}{
		{"Success", []string{"true", "false"}, new([]bool), false, "", []bool{true, false}},
		{"Parse Error", []string{"true", "invalid"}, new([]bool), true, "failed to bind value to bool", nil},
		{"Nil Destination", []string{"true"}, nil, true, "destination cannot be nil", nil},
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
