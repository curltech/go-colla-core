package content

import (
	"errors"
	"github.com/curltech/go-colla-core/config"
	"github.com/curltech/go-colla-core/logger"
	"io/ioutil"
	"os"
	"strings"
)

type ContentStream interface {
	Read(name string) ([]byte, error)
	Write(name string, data []byte) error
}

type fileContent struct {
	filePerm os.FileMode
	path     string
	gap      int
}

var FileContent fileContent

func (this *fileContent) Read(contentId string) ([]byte, error) {
	pathname, filename := this.getFilename(contentId)
	name := pathname + "/" + filename
	existed := exist(name)
	if existed {
		return ioutil.ReadFile(name)
	} else {
		logger.Errorf("filename:%v is exist", name)
		return nil, errors.New("FileNotExist")
	}
}

func exist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func (this *fileContent) Write(contentId string, data []byte) error {
	pathname, filename := this.getFilename(contentId)
	existed := exist(pathname)
	if !existed {
		os.MkdirAll(pathname, this.filePerm)
	}
	name := pathname + "/" + filename
	existed = exist(name)
	if existed {
		logger.Warnf("filename:%v is exist, will be overrided", name)
	}

	return ioutil.WriteFile(name, data, this.filePerm)
}

func (this *fileContent) getFilename(contentId string) (string, string) {
	if contentId == "" {
		panic("NoContentId")
	}
	var pathname string = this.path
	var filename string
	seg := len(contentId) / this.gap
	mod := len(contentId) % this.gap
	if mod > 0 {
		seg++
	}
	for i := 0; i < seg; i++ {
		slice := contentId[i*this.gap : (i+1)*this.gap]
		if i == seg-1 {
			filename = slice
		} else {
			pathname = pathname + "/" + slice
		}
	}

	return pathname, filename
}

func init() {
	path, _ := config.GetString("content.path")
	if path == "" {
		panic("contentpath is not exist")
	}
	path = strings.TrimSuffix(path, "/")
	gap, _ := config.GetInt("content.gap", 4)
	logger.Infof("content basepath:%v", path)
	perm, _ := config.GetUint32("content.perm", 0664)
	filePerm := os.FileMode(perm)
	FileContent = fileContent{path: path, gap: gap, filePerm: filePerm}
}
