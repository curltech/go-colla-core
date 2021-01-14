package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"github.com/curltech/go-colla-core/logger"
	"io/ioutil"
)

func GzipCompress(data []byte) []byte {
	buf := bytes.NewBuffer(nil)
	gzipWrite := gzip.NewWriter(buf)
	defer gzipWrite.Close()
	_, err := gzipWrite.Write(data)
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}
	err = gzipWrite.Flush()
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}

	return buf.Bytes()
}

func FlateCompress(data []byte, level int) []byte {
	buf := bytes.NewBuffer(nil)
	flateWrite, err := flate.NewWriter(buf, level)
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}
	defer flateWrite.Close()
	// 写入待压缩内容
	_, err = flateWrite.Write(data)
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}
	err = flateWrite.Flush()
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}

	return buf.Bytes()
}

func GzipUncompress(data []byte) []byte {
	// 一个缓存区压缩的内容
	buf := bytes.NewBuffer(data)
	// 解压刚压缩的内容
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}
	defer gzipReader.Close()
	bs, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}

	return bs
}

func FlateUncompress(data []byte) []byte {
	// 一个缓存区压缩的内容
	buf := bytes.NewBuffer(data)
	// 解压刚压缩的内容
	flateReader := flate.NewReader(buf)
	defer flateReader.Close()
	bs, err := ioutil.ReadAll(flateReader)
	if err != nil {
		logger.Errorf("%v", err)
		panic(err)
	}

	return bs
}
