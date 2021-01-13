package json

import (
	"encoding/json"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/util/reflect"
	jsoniter "github.com/json-iterator/go"
)

var lib, _ = config.GetString("app.json", "std")

func Marshal(v interface{}) ([]byte, error) {
	var err error
	var r []byte
	if lib == "std" {
		r, err = json.Marshal(v)
	} else if lib == "iter" {
		var json_iterator = jsoniter.ConfigCompatibleWithStandardLibrary
		r, err = json_iterator.Marshal(v)
	}
	if err != nil {
		return nil, err
	}

	return r, nil
}

func Unmarshal(data []byte, v interface{}) error {
	var err error
	if !reflect.IsPtr(v) {
		panic("MustPtr")
	}
	if lib == "std" {
		//decoder := json.NewDecoder(strings.NewReader(string(data)))
		//decoder.UseNumber()
		//err := decoder.Decode(v)
		err = json.Unmarshal(data, v)
	} else if lib == "iter" {
		var json_iterator = jsoniter.ConfigCompatibleWithStandardLibrary
		//decoder := jsoniter.NewDecoder(strings.NewReader(string(data)))
		//decoder.UseNumber()
		//err := decoder.Decode(v)
		err = json_iterator.Unmarshal(data, v)
	}
	if err != nil {
		return err
	}

	return nil
}

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
