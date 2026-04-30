package binding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringBinder_Mappable(t *testing.T) {
	m := StringBinder{}
	tests := []struct {
		dest     any
		expected bool
	}{
		{string(""), false},
		{new(string), true},
		{[]string{}, false},
		{new([]string), true},
		{int(1), false},
		{nil, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, m.Mappable(tt.dest), "Mappable fail for %v", tt.dest)
	}
}

func TestStringBinder_Bind(t *testing.T) {
	m := StringBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst string
		err := m.Bind("foo", &dst)
		assert.NoError(t, err)
		assert.Equal(t, "foo", dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst int
		err := m.Bind("foo", &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})

	t.Run("NilPointer", func(t *testing.T) {
		err := m.Bind("foo", (*string)(nil))
		assert.Error(t, err)
		assert.Equal(t, "destination cannot be nil", err.Error())
	})
}

func TestStringBinder_BindMany(t *testing.T) {
	m := StringBinder{}

	t.Run("Success", func(t *testing.T) {
		var dst []string
		err := m.BindMany([]string{"foo", "bar"}, &dst)
		assert.NoError(t, err)
		assert.Equal(t, []string{"foo", "bar"}, dst)
	})

	t.Run("Unmappable", func(t *testing.T) {
		var dst []int
		err := m.BindMany([]string{"foo", "bar"}, &dst)
		assert.Error(t, err)
		assert.Equal(t, "invalid destination type for binder", err.Error())
	})
}

func TestStringBinder_BindT(t *testing.T) {
	m := StringBinder{}

	tests := []struct {
		name     string
		src      string
		dst      *string
		wantErr  bool
		errMsg   string
		expected string
	}{
		{"Success", "foo", new(string), false, "", "foo"},
		{"Nil Destination", "foo", nil, true, "destination cannot be nil", ""},
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

func TestStringBinder_BindManyT(t *testing.T) {
	m := StringBinder{}

	tests := []struct {
		name     string
		src      []string
		dst      *[]string
		wantErr  bool
		errMsg   string
		expected []string
	}{
		{"Success", []string{"foo", "bar"}, new([]string), false, "", []string{"foo", "bar"}},
		{"Nil Destination", []string{"foo"}, nil, true, "destination cannot be nil", nil},
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
