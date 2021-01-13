package collection

import (
	"github.com/goinggo/mapstructure"
	"reflect"
	"sync"
)

func MapToStruct(m map[string]interface{}, obj interface{}) error {
	//value := reflect.ValueOf(obj)
	//
	//for name,val:= range m {
	//	field := value.FieldByName(name)
	//	if field.CanSet() {
	//		field.Set(reflect.ValueOf(val))
	//	} else {
	//		return errors.New("canot set")
	//	}
	//}
	//
	//return nil

	return mapstructure.Decode(m, obj)
}

type MapOptions struct {
	OmitEmpty  bool
	OmitBool   bool
	OmitNumber bool
}

func StructToMap(obj interface{}, options *MapOptions) map[string]interface{} {
	value := reflect.ValueOf(obj)
	value = reflect.Indirect(value)
	typ := value.Type()

	var data = make(map[string]interface{})
	for i := 0; i < typ.NumField(); i++ {
		f := value.Field(i)
		if f.CanInterface() && f.IsValid() {
			v := f.Interface()
			if typ.Field(i).Anonymous {
				baseData := StructToMap(v, options)
				if baseData != nil {
					for k, v := range baseData {
						data[k] = v
					}
				}
			} else {
				if options != nil && options.OmitEmpty {
					kind := typ.Field(i).Type.Kind()
					switch kind {
					case reflect.String:
						if f.Len() == 0 {
							continue
						}
					case reflect.Ptr, reflect.Interface, reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
						if f.IsNil() {
							continue
						}
					case reflect.Int, reflect.Float64, reflect.Int64, reflect.Uint, reflect.Uint64:
						if f.IsZero() {
							continue
						}
					}
				}
				data[typ.Field(i).Name] = v
			}
		}
	}

	return data
}

type QueueNode struct {
	Data interface{}
	Next *QueueNode // 比自己老的
}

type Queue struct {
	First  *QueueNode //最老的
	Last   *QueueNode //最新的
	Length int
	mutex  sync.Mutex // 对queue关闭上锁
}

//创建链列（数据）
func (queue *Queue) Create(Data ...interface{}) {
	if len(Data) == 0 {
		return
	}

	//创建链列
	for _, v := range Data {
		queue.Push(v)
	}
}

//入列(insert)
func (queue *Queue) Push(Data interface{}) {
	if Data == nil {
		return
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	newNode := new(QueueNode)
	newNode.Data = Data
	if queue.Last != nil {
		queue.Last.Next = newNode
	}

	queue.Last = newNode
	if queue.First == nil {
		queue.First = newNode
	}
	queue.Length++
}

//出队(delete)
func (queue *Queue) Pop() interface{} {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.First != nil {
		data := queue.First.Data
		queue.First = queue.First.Next
		queue.Length--

		return data
	} else {
		return nil
	}
}
