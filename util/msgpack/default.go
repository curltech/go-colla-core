package msgpack

import (
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
)

func Marshal(v interface{}) ([]byte, error) {
	b, err := msgpack.Marshal(v) // 将结构体转化为二进制流
	if err != nil {
		fmt.Printf("msgpack marshal failed,err:%v", err)
		return nil, err
	}

	return b, nil
}

func Unmarshal(data []byte, v interface{}) error {
	err := msgpack.Unmarshal(data, v) // 将二进制流转化回结构体
	if err != nil {
		fmt.Printf("msgpack unmarshal failed,err:%v", err)
		return err
	}

	return nil
}
