package gob

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func Marshal(v interface{}) ([]byte, error) {
	// encode
	buf := new(bytes.Buffer)   // 创建一个buffer区
	enc := gob.NewEncoder(buf) // 创建新的需要转化二进制区域对象
	err := enc.Encode(v)       // 将数据转化为二进制流
	if err != nil {
		fmt.Println("gob encode failed, err:", err)
		return nil, err
	}
	b := buf.Bytes() // 将二进制流赋值给变量b
	fmt.Println(b)

	return b, nil
}

func Unmarshal(data []byte, v interface{}) error {
	// decode
	dec := gob.NewDecoder(bytes.NewBuffer(data)) // 创建一个对象 把需要转化的对象放入
	err := dec.Decode(v)                         // 进行流转化
	if err != nil {
		fmt.Println("gob decode failed, err", err)
		return err
	}
	fmt.Println(v)

	return nil
}
