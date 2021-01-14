package config

import (
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var confs map[string]map[string]interface{}

const prefix string = "./conf/"
const suffix string = ".yml"
const defaultappname string = "app"

func parseYAML(filename string) map[string]interface{} {
	if confs == nil {
		confs = make(map[string]map[string]interface{})
	}
	confname := strings.TrimSuffix(filename, suffix)
	confname = strings.TrimPrefix(confname, prefix)
	conf := confs[confname]
	if conf == nil {
		conf = make(map[string]interface{})
		confs[confname] = conf
	}

	yamlAbsPath, err := filepath.Abs(filename)
	if err != nil {
		panic(err)
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		panic(err)
	}

	// put the file's contents as yaml to the default configuration(c)
	if err := yaml.Unmarshal(data, conf); err != nil {
		panic(err)
	}

	return conf
}

/**
根据启动参数来确定当前应用的名称，从而能够读取特定的配置参数
*/
var appName string

func GetAppName() string {
	if appName == "" {
		appName = "app"
		fmt.Errorf("No AppName, will be set:%v", appName)
	} else {
		fmt.Printf("AppName is:%v", appName)
	}

	return appName
}

func init() {
	appname := flag.String("appname", "", "app name(default app)")
	flag.Parse()
	if len(*appname) == 0 {
		fmt.Errorf("'appname' is required")
		panic("'appname' is required")
	} else {
		appName = *appname
	}

	//读取公共配置
	parseYAML(prefix + defaultappname + suffix)
	files, err := ioutil.ReadDir(prefix)
	if err != nil {
		fmt.Errorf("%v path is not exist:%v", prefix, err)
	}
	//读取各应用的单独配置
	for _, file := range files {
		filename := file.Name()
		if filename != defaultappname+".yml" && filename != "iris.yml" {
			if strings.HasSuffix(filename, suffix) {
				parseYAML(prefix + filename)
			}
		}
	}
}

func Get(name string) (interface{}, error) {
	v, err := get(name, appName)
	if err != nil {
		if appName != defaultappname {
			v, err = get(name, defaultappname)
		}
	}

	return v, err
}

func get(name string, appname string) (interface{}, error) {
	data, ok := confs[appname]
	if !ok {
		fmt.Errorf("app conf is not exist:%v", appname)
		if appname == defaultappname {
			fmt.Errorf("default app conf is not exist:%v", defaultappname)
			panic(defaultappname)
		}

		return nil, errors.New("NoAppName")
	}
	path := strings.Split(name, ".")
	var v interface{}
	for key, value := range path {
		v, ok = data[value]
		if !ok {
			break
		}

		if (key + 1) == len(path) {
			return v, nil
		}
		if v != nil && reflect.TypeOf(v).String() == "map[string]interface {}" {
			data = v.(map[string]interface{})
		}
	}

	return "", fmt.Errorf("Not Exist")
}

func GetString(name string, defaultValue ...string) (string, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return "", nil
		}
		if old != nil {
			return fmt.Sprintf("%v", old), nil
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return "", fmt.Errorf("Failed")
}

func GetInt(name string, defaultValue ...int) (int, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b int
		b, ok := old.(int)
		if ok {
			return b, nil
		} else {
			b, err = strconv.Atoi(fmt.Sprintf("%v", old))
			if err == nil {
				return b, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetInt64(name string, defaultValue ...int64) (int64, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b int64
		b, ok := old.(int64)
		if ok {
			return b, nil
		} else {
			b, err = strconv.ParseInt(fmt.Sprintf("%v", old), 10, 64)
			if err == nil {
				return b, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetInt32(name string, defaultValue ...int32) (int32, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b int32
		b, ok := old.(int32)
		if ok {
			return b, nil
		} else {
			b, err := strconv.ParseInt(fmt.Sprintf("%v", old), 10, 32)
			if err == nil {
				i := int32(b)

				return i, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetUint64(name string, defaultValue ...uint64) (uint64, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b uint64
		b, ok := old.(uint64)
		if ok {
			return b, nil
		} else {
			b, err := strconv.ParseInt(fmt.Sprintf("%v", old), 10, 64)
			if err == nil {
				i := uint64(b)

				return i, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetUint32(name string, defaultValue ...uint32) (uint32, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b uint32
		b, ok := old.(uint32)
		if ok {
			return b, nil
		} else {
			b, err := strconv.ParseInt(fmt.Sprintf("%v", old), 10, 32)
			if err == nil {
				i := uint32(b)

				return i, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetUint16(name string, defaultValue ...uint16) (uint16, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b uint16
		b, ok := old.(uint16)
		if ok {
			return b, nil
		} else {
			b, err := strconv.ParseInt(fmt.Sprintf("%v", old), 10, 16)
			if err == nil {
				i := uint16(b)

				return i, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetBool(name string, defaultValue ...bool) (bool, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return false, nil
		}
		var b bool
		b, ok := old.(bool)
		if ok {
			return b, nil
		} else {
			b, err := strconv.ParseBool(fmt.Sprintf("%v", old))
			if err == nil {
				return b, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return false, fmt.Errorf("Failed")
}

func GetFloat(name string, defaultValue ...float64) (float64, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b float64
		b, ok := old.(float64)
		if ok {
			return b, nil
		} else {
			b, err := strconv.ParseFloat(fmt.Sprintf("%v", old), 64)
			if err == nil {
				return b, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetUint(name string, defaultValue ...uint) (uint, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return 0, nil
		}
		var b uint
		b, ok := old.(uint)
		if ok {
			return b, nil
		} else {
			b, err := strconv.ParseUint(fmt.Sprintf("%v", old), 10, 64)
			if err == nil {
				return uint(b), nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}

	return 0, fmt.Errorf("Failed")
}

func GetTime(name string, defaultValue ...time.Time) (time.Time, error) {
	old, err := Get(name)
	if err == nil {
		if old == nil {
			return time.Time{}, nil
		}
		var b time.Time
		b, ok := old.(time.Time)
		if ok {
			return b, nil
		} else {
			b, err := time.Parse(time.RFC3339Nano, fmt.Sprintf("%v", old))
			if err == nil {
				return b, nil
			}
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0], nil
	}
	var t time.Time
	return t, fmt.Errorf("Failed")
}
