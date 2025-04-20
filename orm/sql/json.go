package sql

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JsonColumn[
	T any] struct {
	val   T
	Valid bool
}

func (j JsonColumn[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}
	return json.Marshal(j.val)
}

func (j *JsonColumn[T]) Scan(src any) error {
	//    int64
	//    float64
	//    bool
	//    []byte
	//    string
	//    time.Time
	//    nil - for NULL values

	var bs []byte
	switch val := src.(type) {
	case string:
		// "" is not considered
		bs = []byte(val)
	case []byte:
		// []byte{} is not considered
		bs = val
	case nil:
		return nil
	default:
		return errors.New("unknown type for JsonColumn[T]")
	}

	err := json.Unmarshal(bs, &j.val)
	if err != nil {
		return err
	}
	j.Valid = true
	return nil
}
