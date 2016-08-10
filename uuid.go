package sharding

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"gopkg.in/pg.v4/types"
)

const uuidLen = 16
const uuidHexLen = 36

type UUID []byte

var _ types.ValueAppender = (UUID)(nil)
var _ sql.Scanner = (*UUID)(nil)
var _ driver.Valuer = (*UUID)(nil)

func NewUUID(shard int64, tm time.Time) UUID {
	shard = shard % 2048

	b := make([]byte, uuidLen)
	binary.BigEndian.PutUint64(b[:8], uint64(unixMicrosecond(tm)))
	rand.Read(b[8:])
	b[8] = (b[8] &^ 0x7) | byte(shard>>8)
	b[9] = byte(shard)
	return b
}

func ParseUUID(b []byte) (UUID, error) {
	if len(b) != uuidHexLen {
		return nil, fmt.Errorf("sharding: invalid UUID: %s", b)
	}
	u := make([]byte, uuidLen)
	hex.Decode(u[:4], b[:8])
	hex.Decode(u[4:6], b[9:13])
	hex.Decode(u[6:8], b[14:18])
	hex.Decode(u[8:10], b[19:23])
	hex.Decode(u[10:], b[24:])
	return u, nil
}

func (u UUID) Split() (shardId int64, tm time.Time) {
	b := []byte(u)
	tm = fromUnixMicrosecond(int64(binary.BigEndian.Uint64(b[:8])))
	shardId |= (int64(b[8]) & 0x7) << 8
	shardId |= int64(b[9])
	return
}

func (u UUID) String() string {
	b, _ := u.AppendValue(nil, 0)
	return string(b)
}

func (u UUID) AppendValue(b []byte, quote int) ([]byte, error) {
	if u == nil {
		return types.AppendNull(b, quote), nil
	}
	if len(u) != uuidLen {
		return nil, fmt.Errorf("sharding: invalid UUID: % x", b)
	}

	if quote == 2 {
		b = append(b, '"')
	} else if quote == 1 {
		b = append(b, '\'')
	}

	b = append(b, make([]byte, uuidHexLen)...)

	bb := b[len(b)-uuidHexLen:]
	hex.Encode(bb[:8], u[:4])
	bb[8] = '-'
	hex.Encode(bb[9:13], u[4:6])
	bb[13] = '-'
	hex.Encode(bb[14:18], u[6:8])
	bb[18] = '-'
	hex.Encode(bb[19:23], u[8:10])
	bb[23] = '-'
	hex.Encode(bb[24:], u[10:])

	if quote == 2 {
		b = append(b, '"')
	} else if quote == 1 {
		b = append(b, '\'')
	}

	return b, nil
}

func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

func (u *UUID) Scan(b interface{}) error {
	if b == nil {
		*u = nil
	}
	uuid, err := ParseUUID(b.([]byte))
	if err != nil {
		return err
	}
	*u = uuid
	return nil
}

func unixMicrosecond(tm time.Time) int64 {
	return tm.Unix()*1e6 + int64(tm.Nanosecond())/1e3
}

func fromUnixMicrosecond(n int64) time.Time {
	secs := n / 1e6
	return time.Unix(secs, (n-secs*1e6)*1e3)
}