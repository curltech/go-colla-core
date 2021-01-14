package reflect

import (
	"errors"
	"github.com/curltech/go-colla-core/logger"
	"github.com/curltech/go-colla-core/util/convert"
	"reflect"
)

func GetType(value interface{}) reflect.Kind {
	typ := reflect.TypeOf(value)

	return typ.Kind()
}

func GetIndirectType(value interface{}) reflect.Kind {
	v := reflect.Indirect(reflect.ValueOf(value))

	return v.Type().Kind()
}

func GetFieldNames(obj interface{}, recursive bool) (map[string]string, error) {
	value := reflect.ValueOf(obj)
	var fieldNames = make(map[string]string, 0)
	objValue := reflect.Indirect(value)
	num := objValue.Type().NumField()
	for i := 0; i < num; i++ {
		f := objValue.Type().Field(i)
		typ := f.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if typ.Kind() == reflect.Struct && typ.String() != "time.Time" {
			if recursive {
				v := objValue.FieldByName(f.Name).Interface()
				ns, err := GetFieldNames(v, recursive)
				if err == nil && ns != nil && len(ns) > 0 {
					for n, t := range ns {
						fieldNames[n] = t
					}
				}
			}
		} else if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map || typ.Kind() == reflect.Interface || typ.Kind() == reflect.Array {

		} else {
			fieldNames[f.Name] = typ.String()
		}
	}

	return fieldNames, nil
}

func GetValue(obj interface{}, name string) (interface{}, error) {
	value := reflect.Indirect(reflect.ValueOf(obj))
	typ := value.Type().Kind()
	if reflect.Struct == typ {
		v := value.FieldByName(name)
		if v.IsValid() {
			return v.Interface(), nil
		} else {
			return nil, errors.New("NoFieldName")
		}
	} else if reflect.Map == typ {
		m := value.Interface().(map[string]interface{})

		return m[name], nil
	}

	return nil, errors.New("NotStruct")
}

func SetValue(obj interface{}, name string, val interface{}) error {
	value := reflect.Indirect(reflect.ValueOf(obj))
	typ := value.Type().Kind()
	if reflect.Struct == typ {
		field := value.FieldByName(name)
		if !field.IsValid() {
			return errors.New("FieldNameInvalid")
		}
		if field.CanSet() {
			field.Set(reflect.ValueOf(val))
		} else {
			logger.Errorf("name:%v canot set", name)
			return errors.New("canot set")
		}
	} else if reflect.Map == typ {
		m := value.Interface().(map[string]interface{})
		m[name] = val
	}

	return nil
}

func Set(obj interface{}, name string, val string) error {
	typ := reflect.TypeOf(obj)
	typName := typ.Kind().String()
	value := reflect.ValueOf(obj)
	if "ptr" == typName {
		value = value.Elem()
	}
	field := value.FieldByName(name)
	if !field.IsValid() {
		return errors.New("FieldNameInvalid")
	}
	fieldType := field.Type().String()
	o, err := convert.ToObject(val, fieldType)
	if err != nil {
		return err
	}
	if field.CanSet() {
		field.Set(reflect.ValueOf(o))
	} else {
		logger.Errorf("name:%v canot set", name)
		return errors.New("canot set")
	}

	return nil
}

func IsPtr(obj interface{}) bool {
	typ := reflect.TypeOf(obj)
	if reflect.Ptr == typ.Kind() {
		return true
	}
	return false
}

func IsSlice(obj interface{}) bool {
	typ := GetIndirectType(obj)
	if reflect.Slice == typ {
		return true
	}
	return false
}

//反射创建新对象。
func New(target interface{}) interface{} {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Ptr { //指针类型获取真正type需要调用Elem
		t = t.Elem()
	}
	ptr := reflect.New(t) // 调用反射创建对象
	//value := ptr.Elem()
	//val := value.Interface()

	return ptr.Interface()
}

func CreateArgs(service interface{}, methodName string) []interface{} {
	if service != nil {
		serviceValue := reflect.ValueOf(service)
		methodValue := serviceValue.MethodByName(methodName)
		methodType := reflect.TypeOf(methodValue)
		numIn := methodType.NumIn() //返回func类型的参数个数，如果不是函数，将会panic
		methodIn := make([]reflect.Type, numIn)
		args := make([]interface{}, numIn)
		for i := 0; i < numIn; i++ {
			methodIn[i] = methodType.In(i)
			args[i] = New(methodIn[i])
		}

		return args
	} else {
		panic("InvalidService")
	}
}

func Call(service interface{}, methodName string, args []interface{}) ([]interface{}, error) {
	if service != nil {
		serviceValue := reflect.ValueOf(service)
		methodValue := serviceValue.MethodByName(methodName)
		if methodValue.IsValid() {
			params := make([]reflect.Value, len(args))
			for i, _ := range args {
				params[i] = reflect.ValueOf(args[i])
			}
			vs := methodValue.Call(params)
			results := make([]interface{}, len(vs))
			for i, v := range vs {
				results[i] = v.Interface()
			}

			return results, nil
		} else {
			logger.Errorf("InvalidMethod:%v", methodName)

			return nil, errors.New("InvalidMethod")
		}
	} else {
		logger.Errorf("InvalidService")

		return nil, errors.New("InvalidService")
	}
}

func Invoke(fn interface{}, args []interface{}) []interface{} {
	fv := reflect.ValueOf(fn)
	params := make([]reflect.Value, len(args))
	for i, _ := range args {
		params[i] = reflect.ValueOf(args[i])
	}
	vs := fv.Call(params)
	results := make([]interface{}, len(vs))
	for i, v := range vs {
		results[i] = v.Interface()
	}

	return results
}

func ToArray(entities interface{}) []interface{} {
	val := reflect.Indirect(reflect.ValueOf(entities))
	kind := val.Type().Kind()
	if kind == reflect.Slice || kind == reflect.Array {
		rowsSlicePtr := make([]interface{}, 0)
		for i := 0; i < val.Len(); i++ {
			ele := val.Index(i)
			rowsSlicePtr = append(rowsSlicePtr, ele.Interface().(interface{}))
		}
		return rowsSlicePtr
	}

	return nil
}

//在入参的数组中创建并增加同类型的对象，并返回这个新增的对象指针
func Append(rowsSlicePtr interface{}) interface{} {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	var isSlice = sliceValue.Kind() == reflect.Slice
	if !isSlice {
		logger.Errorf("needs a pointer to a slice")
		return nil
	}
	sliceElementType := sliceValue.Type().Elem()
	//创建新的元素
	var pv reflect.Value
	if sliceElementType.Kind() == reflect.Ptr {
		if sliceElementType.Elem().Kind() == reflect.Struct {
			pv = reflect.New(sliceElementType.Elem())
			sliceValue.Set(reflect.Append(sliceValue, pv))
		} else {
			return nil
		}
	} else if sliceElementType.Kind() == reflect.Struct {
		pv = reflect.New(sliceElementType)
		sliceValue.Set(reflect.Append(sliceValue, pv))
	} else {
		return nil
	}

	value := pv.Elem()
	val := value.Interface()
	logger.Infof("%v", val)

	return &val
}
