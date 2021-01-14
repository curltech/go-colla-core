package container

import (
	"github.com/curltech/go-colla-core/logger"
)

var controllerContainer = make(map[string]interface{})

var serviceContainer = make(map[string]interface{})

func init() {

}

func Regist(typ string, name string, beanPtr interface{}) {
	var c map[string]interface{}
	if typ == "controller" {
		c = controllerContainer
	}
	if typ == "service" {
		c = serviceContainer
	}
	_, ok := c[name]
	if !ok {
		c[name] = beanPtr
		logger.Infof("bean:%v registed", name)
	} else {
		logger.Warnf("bean:%v exist", name)
	}
}

func RegistController(name string, beanPtr interface{}) {
	Regist("controller", name, beanPtr)
}

func RegistService(name string, beanPtr interface{}) {
	Regist("service", name, beanPtr)
}

func Get(typ string, name string) interface{} {
	var c map[string]interface{}
	if typ == "controller" {
		c = controllerContainer
	}
	if typ == "service" {
		c = serviceContainer
	}
	old, ok := c[name]
	if ok {
		return old.(interface{})
	}

	return nil
}

func GetService(name string) interface{} {
	return Get("service", name)
}

func GetController(name string) interface{} {
	return Get("controller", name)
}
