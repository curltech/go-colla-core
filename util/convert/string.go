package convert

import (
	"fmt"
	"strconv"
	"time"
)

const (
	DataPattern_Super  = "super"
	DataPattern_High   = "high"
	DataPattern_Normal = "normal"
	DataPattern_Low    = "low"
)

func ToString(value interface{}, dataType string, pattern string) (string, error) {
	if value == nil {
		return "", nil
	}
	var prec int
	var layout string
	if pattern == DataPattern_Super {
		prec = 8
		layout = time.RFC3339Nano
	} else if pattern == DataPattern_High {
		prec = 6
		layout = time.RFC3339
	} else if pattern == DataPattern_Low {
		prec = 4
		layout = time.RFC3339
	} else {
		prec = 2
		layout = time.RFC3339Nano
	}
	if dataType == "time" || dataType == "time.Time" || dataType == "*time.Time" || dataType == "timestamp" || dataType == "Timestamp" || dataType == "Date" || dataType == "date" {
		v := value.(time.Time)
		return v.Format(layout), nil
	} else if dataType == "float64" || dataType == "float" || dataType == "BigDecimal" || dataType == "Float" || dataType == "Double" || dataType == "double" {
		v := value.(float64)

		return strconv.FormatFloat(v, 'f', prec, 64), nil
	}

	return fmt.Sprintf("%v", value), nil
}

func ToObject(value string, dataType string) (interface{}, error) {
	switch dataType {
	case "int", "Integer":
		return strconv.Atoi(value)
	case "int64", "Long", "long":
		return strconv.ParseInt(value, 10, 64)
	case "int32":
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return int32(0), err
		}
		return int32(v), err
	case "uint64":
		return strconv.ParseUint(value, 10, 64)
	case "uint32":
		v, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return uint32(0), err
		}
		return uint32(v), err
	case "bool", "Boolean", "boolean":
		return strconv.ParseBool(value)
	case "float64", "BigDecimal", "Float", "Double", "double", "float":
		return strconv.ParseFloat(value, 64)
	case "uint":
		return strconv.ParseUint(value, 10, 0)
	case "time", "Date", "date", "time.Time", "*time.Time", "timestamp", "Timestamp":
		t, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			t, err = time.Parse(time.RFC3339, value)
			if err != nil {
				t, err = time.Parse("2006-01-02 15:04:05.000000", value)
				if err != nil {
					t, err = time.Parse(time.RFC1123Z, value)
				}
			}
		}

		return &t, err
	}

	return value, nil
}
