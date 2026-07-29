package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat32Binder_Mappable(t *testing.T) {
	m := Float32Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{float32(0), false},
		{new(float32), true},
		{[]float32{}, false},
		{new([]float32), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestFloat32Binder_Bind(t *testing.T) {
	m := Float32Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst float32
		err := m.Bind("1.23", &dst)
		assert.NoError(t, err)
		assert.InDelta(t, float32(1.23), dst, 0.0001)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("1.23", &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("1.23", (*float32)(nil))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationNil)
	})
}

func TestFloat32Binder_BindMany(t *testing.T) {
	m := Float32Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []float32
		err := m.BindMany([]string{"1.1", "2.2"}, &dst)
		assert.NoError(t, err)
		assert.Len(t, dst, 2)
		assert.InDelta(t, float32(1.1), dst[0], 0.0001)
		assert.InDelta(t, float32(2.2), dst[1], 0.0001)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1.1", "2.2"}, &dst)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})
}

func TestFloat32Binder_BindT(t *testing.T) {
	m := Float32Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *float32
		wantErr  bool
		errMsg   string
		expected float32
	}{
		{"Success", "1.23", new(float32), false, "", 1.23},
		{"Parse Error", "abc", new(float32), true, "failed to bind value to float32", 0},
		{"Nil Destination", "1.23", nil, true, "destination cannot be nil", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.BindT(tt.src, tt.dst)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, *tt.dst, 0.0001)
			}
		})
	}
}

func TestFloat32Binder_BindManyT(t *testing.T) {
	m := Float32Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]float32
		wantErr  bool
		errMsg   string
		expected []float32
	}{
		{"Success", []string{"1.1", "2.2"}, new([]float32), false, "", []float32{1.1, 2.2}},
		{"Parse Error", []string{"1.1", "abc"}, new([]float32), true, "failed to bind value to float32", nil},
		{"Nil Destination", []string{"1.1"}, nil, true, "destination cannot be nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.BindManyT(tt.src, tt.dst)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Len(t, *tt.dst, len(tt.expected))
				for i := range tt.expected {
					assert.InDelta(t, tt.expected[i], (*tt.dst)[i], 0.0001)
				}
			}
		})
	}
}
