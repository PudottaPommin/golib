package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64Binder_Mappable(t *testing.T) {
	m := Float64Binder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{float64(0), true},
		{new(float64), true},
		{[]float64{}, true},
		{new([]float64), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestFloat64Binder_Bind(t *testing.T) {
	m := Float64Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst float64
		err := m.Bind("1.23", &dst)
		assert.NoError(t, err)
		assert.InDelta(t, 1.23, dst, 0.0001)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("1.23", &dst)
		assert.Error(t, err)
		assert.Equal(t, "destination is not mappable", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("1.23", (*float64)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestFloat64Binder_BindMany(t *testing.T) {
	m := Float64Binder{}

	t.Run("Success", func(t *testing.T) {
		var dst []float64
		err := m.BindMany([]string{"1.1", "2.2"}, &dst)
		assert.NoError(t, err)
		assert.Len(t, dst, 2)
		assert.InDelta(t, 1.1, dst[0], 0.0001)
		assert.InDelta(t, 2.2, dst[1], 0.0001)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"1.1", "2.2"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "destination is not mappable", err.Error())
	})
}

func TestFloat64Binder_BindT(t *testing.T) {
	m := Float64Binder{}

	tests := []struct {
		name     string
		src      string
		dst      *float64
		wantErr  bool
		errMsg   string
		expected float64
	}{
		{"Success", "1.23", new(float64), false, "", 1.23},
		{"Parse Error", "abc", new(float64), true, "failed to bind value to float64", 0},
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

func TestFloat64Binder_BindManyT(t *testing.T) {
	m := Float64Binder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]float64
		wantErr  bool
		errMsg   string
		expected []float64
	}{
		{"Success", []string{"1.1", "2.2"}, new([]float64), false, "", []float64{1.1, 2.2}},
		{"Parse Error", []string{"1.1", "abc"}, new([]float64), true, "failed to bind value to float64", nil},
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
