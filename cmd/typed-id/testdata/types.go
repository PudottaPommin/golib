package testdata

import "uuid"

type UserID string
type AccountID int64
type RoleID uint32
type SmallID int8

//typed-id: presets=sql
type SQLOnlyUUID uuid.UUID

type TokenID []byte

//typed-id:encoding=base64
type SecretID []byte

//typed-id:sql=hex
type HexSQLID []byte

//typed-id:sql=base64
type Base64SQLID []byte

//typed-id: presets=json,sql encoding=base64 sql=bytes
type CustomSecretID []byte

type Metadata map[string]any
type FlexData map[any]any
