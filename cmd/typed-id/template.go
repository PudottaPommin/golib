package main

import (
	"fmt"
	"strings"
)

func generateTextPreset(t *TypeInfo) string {
	var b strings.Builder
	name := t.Name
	bitSize := BitSizeForType(t.Underlying)

	switch t.Kind {
	case KindString:
		b.WriteString(fmt.Sprintf(`// String implements fmt.Stringer.
func (id %s) String() string {
	return string(id)
}

// MarshalText implements encoding.TextMarshaler.
func (id %s) MarshalText() ([]byte, error) {
	return []byte(id), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *%s) UnmarshalText(text []byte) error {
	*id = %s(text)
	return nil
}

`, name, name, name, name))

	case KindInt:
		b.WriteString(fmt.Sprintf(`// String implements fmt.Stringer.
func (id %s) String() string {
	return strconv.FormatInt(int64(id), 10)
}

// MarshalText implements encoding.TextMarshaler.
func (id %s) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(id), 10)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *%s) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, %d)
	if err != nil {
		return err
	}
	*id = %s(v)
	return nil
}

`, name, name, name, bitSize, name))

	case KindUint:
		b.WriteString(fmt.Sprintf(`// String implements fmt.Stringer.
func (id %s) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// MarshalText implements encoding.TextMarshaler.
func (id %s) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatUint(uint64(id), 10)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *%s) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 10, %d)
	if err != nil {
		return err
	}
	*id = %s(v)
	return nil
}

`, name, name, name, bitSize, name))
	case KindUUID:
		b.WriteString(fmt.Sprintf(`// String implements fmt.Stringer.
func (id %s) String() string {
	return uuid.UUID(id).String()
}

// MarshalText implements encoding.TextMarshaler.
func (id %s) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *%s) UnmarshalText(text []byte) error {
	return (*uuid.UUID)(id).UnmarshalText(text)
}

		`, name, name, name))
	}

	return b.String()
}

func generateJSONPreset(t *TypeInfo) string {
	var b strings.Builder
	name := t.Name

	switch t.Kind {
	case KindString:
		b.WriteString(fmt.Sprintf(`// MarshalJSONTo implements json.MarshalerTo.
func (id %[1]s) MarshalJSONTo(in *jsontext.Encoder) error {
	return in.WriteToken(jsontext.String(string(id)))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (id *%[1]s) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var s string
	if err := json.UnmarshalDecode(dec, &s); err != nil {
		return fmt.Errorf("%[1]s failed UnmarshalJSONFrom: %%w", err)
	}
	*id = %[1]s(s)
	return nil
}

`, name))
	case KindUUID:
		b.WriteString(fmt.Sprintf(`// MarshalJSONTo implements json.MarshalerTo.
func (id %[1]s) MarshalJSONTo(in *jsontext.Encoder) error {
	return in.WriteToken(jsontext.String(uuid.UUID(id).String()))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (id *%[1]s) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var s string
	if err := json.UnmarshalDecode(dec, &s); err != nil {
		return fmt.Errorf("%[1]s failed UnmarshalJSONFrom: %%w", err)
	}
	uid, err := uuid.Parse(s)
	if err != nil {
		return fmt.Errorf("%[1]s failed UnmarshalJSONFrom: %%w", err)
	}
	*id = %[1]s(uid)
	return nil
}

`, name))
	case KindInt:
		b.WriteString(fmt.Sprintf(`// MarshalJSONTo implements json.MarshalerTo.
func (id %[1]s) MarshalJSONTo(in *jsontext.Encoder) error {
	return in.WriteToken(jsontext.Int(int64(id)))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (id *%[1]s) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var n int64
	if err := json.UnmarshalDecode(dec, &n); err != nil {
		return fmt.Errorf("%[1]s failed UnmarshalJSONFrom: %%w", err)
	}
	*id = %[1]s(n)
	return nil
}

`, name))

	case KindUint:
		b.WriteString(fmt.Sprintf(`// MarshalJSONTo implements json.MarshalerTo.
func (id %[1]s) MarshalJSONTo(in *jsontext.Encoder) error {
	return in.WriteToken(jsontext.Uint(uint64(id)))
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (id *%[1]s) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var n uint64
	if err := json.UnmarshalDecode(dec, &n); err != nil {
		return fmt.Errorf("%[1]s failed UnmarshalJSONFrom: %%w", err)
	}
	*id = %[1]s(n)
	return nil
}

`, name))
	}

	return b.String()
}

func generateBinaryPreset(t *TypeInfo) string {
	var b strings.Builder
	name := t.Name

	switch t.Kind {
	case KindString:
		b.WriteString(fmt.Sprintf(`// MarshalBinary implements encoding.BinaryMarshaler.
func (id %s) MarshalBinary() ([]byte, error) {
	return []byte(id), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *%s) UnmarshalBinary(data []byte) error {
	*id = %s(data)
	return nil
}

`, name, name, name))
	case KindUUID:
		b.WriteString(fmt.Sprintf(`// MarshalBinary implements encoding.BinaryMarshaler.
func (id %s) MarshalBinary() ([]byte, error) {
	return id[:], nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *%s) UnmarshalBinary(data []byte) error {
	if len(data) != 16 {
		return fmt.Errorf("invalid binary length for %%s: got %%d, expected 16", "%s", len(data))
	}
	copy(id[:], data)
	return nil
}

`, name, name, name))

	case KindInt, KindUint:
		b.WriteString(fmt.Sprintf(`// MarshalBinary implements encoding.BinaryMarshaler.
func (id %s) MarshalBinary() ([]byte, error) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *%s) UnmarshalBinary(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("invalid binary length for %%s: got %%d, expected 8", "%s", len(data))
	}
	*id = %s(binary.BigEndian.Uint64(data))
	return nil
}

`, name, name, name, name))
	}

	return b.String()
}

func generateSQLPreset(t *TypeInfo) string {
	var b strings.Builder
	name := t.Name
	bitSize := BitSizeForType(t.Underlying)

	switch t.Kind {
	case KindString:
		b.WriteString(fmt.Sprintf(`// Value implements driver.Valuer.
func (id %s) Value() (driver.Value, error) {
	return string(id), nil
}

// Scan implements sql.Scanner.
func (id *%s) Scan(src any) error {
	if src == nil {
		*id = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		*id = %s(v)
		return nil
	case []byte:
		*id = %s(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %%T into %%s", src, "%s")
	}
}

`, name, name, name, name, name))
	case KindUUID:
		b.WriteString(fmt.Sprintf(`// Value implements driver.Valuer.
func (id %s) Value() (driver.Value, error) {
	return uuid.UUID(id).String(), nil
}

// Scan implements sql.Scanner.
func (id *%s) Scan(src any) error {
	if src == nil {
		*id = %s(uuid.Nil())
		return nil
	}

	switch src := src.(type) {
	case []byte:
		if len(src) == 16 {
			copy(id[:], src)
			return nil
		}
		return id.UnmarshalText(src)
	case string:
		return id.UnmarshalText([]byte(src))
	default:
		return fmt.Errorf("cannot scan %%T into %%s", src, "%s")
	}
}

`, name, name, name, name))

	case KindInt:
		b.WriteString(fmt.Sprintf(`// Value implements driver.Valuer.
func (id %s) Value() (driver.Value, error) {
	return int64(id), nil
}

// Scan implements sql.Scanner.
func (id *%s) Scan(src any) error {
	if src == nil {
		*id = 0
		return nil
	}
	switch v := src.(type) {
	case int64:
		*id = %s(v)
		return nil
	case int:
		*id = %s(v)
		return nil
	case int32:
		*id = %s(v)
		return nil
	case int16:
		*id = %s(v)
		return nil
	case int8:
		*id = %s(v)
		return nil
	case uint:
		*id = %s(v)
		return nil
	case uint64:
		*id = %s(v)
		return nil
	case uint32:
		*id = %s(v)
		return nil
	case uint16:
		*id = %s(v)
		return nil
	case uint8:
		*id = %s(v)
		return nil
	case []byte:
		val, err := strconv.ParseInt(string(v), 10, %d)
		if err != nil {
			return err
		}
		*id = %s(val)
		return nil
	case string:
		val, err := strconv.ParseInt(v, 10, %d)
		if err != nil {
			return err
		}
		*id = %s(val)
		return nil
	case float64:
		*id = %s(int64(v))
		return nil
	default:
		return fmt.Errorf("cannot scan %%T into %%s", src, "%s")
	}
}

`, name, name, name, name, name, name, name, name, name, name, name, name, bitSize, name, bitSize, name, name, name))

	case KindUint:
		b.WriteString(fmt.Sprintf(`// Value implements driver.Valuer.
func (id %s) Value() (driver.Value, error) {
	return int64(id), nil
}

// Scan implements sql.Scanner.
func (id *%s) Scan(src any) error {
	if src == nil {
		*id = 0
		return nil
	}
	switch v := src.(type) {
	case uint64:
		*id = %s(v)
		return nil
	case uint:
		*id = %s(v)
		return nil
	case uint32:
		*id = %s(v)
		return nil
	case uint16:
		*id = %s(v)
		return nil
	case uint8:
		*id = %s(v)
		return nil
	case int64:
		*id = %s(v)
		return nil
	case int:
		*id = %s(v)
		return nil
	case int32:
		*id = %s(v)
		return nil
	case int16:
		*id = %s(v)
		return nil
	case int8:
		*id = %s(v)
		return nil
	case []byte:
		val, err := strconv.ParseUint(string(v), 10, %d)
		if err != nil {
			return err
		}
		*id = %s(val)
		return nil
	case string:
		val, err := strconv.ParseUint(v, 10, %d)
		if err != nil {
			return err
		}
		*id = %s(val)
		return nil
	case float64:
		*id = %s(uint64(v))
		return nil
	default:
		return fmt.Errorf("cannot scan %%T into %%s", src, "%s")
	}
}

`, name, name, name, name, name, name, name, name, name, name, name, name, bitSize, name, bitSize, name, name, name))
	}

	return b.String()
}

func generateBinderPreset(t *TypeInfo) string {
	var b strings.Builder
	name := t.Name
	bitSize := BitSizeForType(t.Underlying)

	switch t.Kind {
	case KindString:
		b.WriteString(fmt.Sprintf(`// UnmarshalBind implements binding.BindUnmarshaller.
func (id *%s) UnmarshalBind(s string) error {
	*id = %s(s)
	return nil
}

`, name, name))
	case KindUUID:
		b.WriteString(fmt.Sprintf(`// UnmarshalBind implements binding.BindUnmarshaller.
func (id *%s) UnmarshalBind(s string) error {
	gid, err := guid.ParseT[%s](s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bind %s: %%w", err)
	}
	*id = gid 
	return nil
}

`, name, name, name))

	case KindInt:
		b.WriteString(fmt.Sprintf(`/// UnmarshalBind implements binding.BindUnmarshaller.
func (id *%s) UnmarshalBind(s string) error {
	v, err := strconv.ParseInt(s, 10, %d)
	if err != nil {
		return err
	}
	*id = %s(v)
	return nil
}

`, name, bitSize, name))

	case KindUint:
		b.WriteString(fmt.Sprintf(`// UnmarshalBind implements binding.BindUnmarshaller.
func (id *%s) UnmarshalBind(s string) error {
	v, err := strconv.ParseUint(s, 10, %d)
	if err != nil {
		return err
	}
	*id = %s(v)
	return nil
}

`, name, bitSize, name))
	}
	return b.String()
}
