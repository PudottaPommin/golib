package binding_test

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	. "github.com/pudottapommin/golib/http/binding"
	"github.com/stretchr/testify/assert"
)

func TestFuncBinder(t *testing.T) {
	timeBinder := NewFuncBinder(func(src string) (time.Time, error) {
		return time.Parse("2006-01-02", src)
	})
	wrapped := WrapBinder(timeBinder)

	t.Run("Mappable", func(t *testing.T) {
		var tm time.Time
		var slice []time.Time
		var other string
		assert.True(t, wrapped.Mappable(&tm))
		assert.True(t, wrapped.Mappable(&slice))
		assert.False(t, wrapped.Mappable(&other))
		assert.False(t, wrapped.Mappable(tm))
	})

	t.Run("Bind Success", func(t *testing.T) {
		var tm time.Time
		err := timeBinder.Bind("2026-08-20", &tm)
		assert.NoError(t, err)
		assert.Equal(t, 2026, tm.Year())
		assert.Equal(t, time.August, tm.Month())
		assert.Equal(t, 20, tm.Day())
	})

	t.Run("Bind Error", func(t *testing.T) {
		var tm time.Time
		err := timeBinder.Bind("invalid-date", &tm)
		assert.Error(t, err)
	})

	t.Run("Bind Nil Destination", func(t *testing.T) {
		err := timeBinder.Bind("2026-08-20", nil)
		assert.ErrorIs(t, err, ErrDestinationNil)
	})

	t.Run("Wrapped Dynamic Bind Invalid Type", func(t *testing.T) {
		var s string
		err := wrapped.Bind("2026-08-20", &s)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})

	t.Run("BindMany Success", func(t *testing.T) {
		var list []time.Time
		err := timeBinder.BindMany([]string{"2026-01-01", "2026-12-31"}, &list)
		assert.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, time.January, list[0].Month())
		assert.Equal(t, time.December, list[1].Month())
	})

	t.Run("BindMany Error", func(t *testing.T) {
		var list []time.Time
		err := timeBinder.BindMany([]string{"2026-01-01", "bad-date"}, &list)
		assert.Error(t, err)
	})

	t.Run("BindMany Nil Destination", func(t *testing.T) {
		err := timeBinder.BindMany([]string{"2026-01-01"}, nil)
		assert.ErrorIs(t, err, ErrDestinationNil)
	})

	t.Run("Wrapped Dynamic BindMany Invalid Type", func(t *testing.T) {
		var list []string
		err := wrapped.BindMany([]string{"2026-01-01"}, &list)
		assert.ErrorIs(t, err, ErrDestinationTypeInvalid)
	})
}

type EventForm struct {
	Date time.Time `form:"date,required"`
	Code int       `form:"code"`
}

func TestFuncBinder_IntegrationWithFormBinder(t *testing.T) {
	timeBinder := NewFuncBinder(func(src string) (time.Time, error) {
		if src == "" {
			return time.Time{}, errors.New("empty date")
		}
		return time.Parse("2006-01-02", src)
	})

	fb := NewFormBinder[EventForm](FormWithGenericBinder[EventForm](timeBinder))

	form := url.Values{}
	form.Set("date", "2026-08-20")
	form.Set("code", "42")

	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	req.Form = form

	var dst EventForm
	err := fb.BindTo(req, &dst)
	assert.NoError(t, err)
	assert.Equal(t, 2026, dst.Date.Year())
	assert.Equal(t, 42, dst.Code)
}

func BenchmarkFuncBinder(b *testing.B) {
	intFuncBinder := NewFuncBinder(func(src string) (int, error) {
		return strconv.Atoi(src)
	})
	src := "12345"
	var dst int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = intFuncBinder.Bind(src, &dst)
	}
}
