package main_test

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UUID uuid.UUID

var _ json.UnmarshalerFrom

func (id UUID) MarshalJSONTo(in *jsontext.Encoder) error {
	t := jsontext.String(uuid.UUID(id).String())
	return in.WriteToken(t)
}

func (id *UUID) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	k := dec.PeekKind()
	switch k {
	case jsontext.KindString:
		t, err := dec.ReadToken()
		if err != nil {
			return err
		}
		uid, err := uuid.Parse(t.String())
		if err != nil {
			return err
		}
		*id = UUID(uid)
		return nil
	}
	return nil
}

func TestJSON_UUID(t *testing.T) {
	id := UUID(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"))
	b, err := json.Marshal(id)
	require.NoError(t, err)
	assert.Equal(t, `"123e4567-e89b-12d3-a456-426614174000"`, string(b))

	var id2 UUID
	require.NoError(t, json.Unmarshal(b, &id2))
	assert.Equal(t, id, id2)
}

type TInt int

func (i TInt) MarshalJSONTo(in *jsontext.Encoder) error {
	return in.WriteToken(jsontext.Int(int64(i)))
}

func TestJSON_TInt(t *testing.T) {
	i := TInt(123)
	b, err := json.Marshal(i)
	require.NoError(t, err)
	assert.Equal(t, "123", string(b))
}
