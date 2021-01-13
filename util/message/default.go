package message

import (
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/util/gob"
	"github.com/curltech/go-colla-core/util/json"
	"github.com/curltech/go-colla-core/util/msgpack"
)

var Marshal func(v interface{}) ([]byte, error)

var Unmarshal func(data []byte, v interface{}) error

func TextMarshal(v interface{}) (string, error) {
	r, err := Marshal(v)
	if err == nil {
		return string(r), nil
	}

	return "", err
}

func TextUnmarshal(data string, v interface{}) error {
	d := []byte(data)
	return Unmarshal(d, v)
}

type Serializable interface {
	Marshal() []byte

	Unmarshal(data []byte)

	TextMarshal() string

	TextUnmarshal(data string)
}

func init() {
	serialize, _ := config.GetString("app.serialize", "json")
	switch serialize {
	case "json":
		Marshal = json.Marshal
		Unmarshal = json.Unmarshal
	case "msgpack":
		Marshal = msgpack.Marshal
		Unmarshal = msgpack.Unmarshal
	case "gob":
		Marshal = gob.Marshal
		Unmarshal = gob.Unmarshal
	}
}
