package excel

import (
	"errors"
	utilreflect "github.com/curltech/go-colla-core/util/reflect"
	"github.com/kataras/golog"
	"reflect"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func Read(filename string, rowsSlicePtrs ...interface{}) error {
	xlsx, err := excelize.OpenFile(filename)
	if err != nil {
		golog.Errorf("filename:%v can't open", filename)
		return err
	}
	count := xlsx.SheetCount
	length := len(rowsSlicePtrs)
	for c := 0; c < count; c++ {
		if c == length {
			break
		}
		sheetname := xlsx.GetSheetName(c + 1)
		rows := xlsx.GetRows(sheetname)
		var head []string
		sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtrs[c]))
		var isSlice = sliceValue.Kind() == reflect.Slice
		if !isSlice {
			return errors.New("needs a pointer to a slice")
		}
		sliceElementType := sliceValue.Type().Elem()
		for i := 0; i < len(rows); i++ {
			row := rows[i]
			if i == 0 {
				head = row
				continue
			}
			//创建新的元素
			var pv reflect.Value
			if sliceElementType.Kind() == reflect.Ptr {
				if sliceElementType.Elem().Kind() == reflect.Struct {
					pv = reflect.New(sliceElementType.Elem())
					sliceValue.Set(reflect.Append(sliceValue, pv))
				} else {
					continue
				}
			} else if sliceElementType.Kind() == reflect.Struct {
				pv = reflect.New(sliceElementType)
				sliceValue.Set(reflect.Append(sliceValue, pv))
			} else {
				continue
			}
			value := pv.Elem()
			for j := 0; j < len(row); j++ {
				fieldname := head[j]
				err := utilreflect.Set(value, fieldname, row[j])
				if err != nil {
					golog.Errorf("sheetname:%v,row:%v,col:%v,fieldname:%v can't set value:%v", sheetname, i, j, fieldname, value)
					continue
				}
			}
		}
	}
	return nil
}

func Write(rowsSlicePtr interface{}) ([]byte, error) {
	f := excelize.NewFile()
	// Create a new sheet.
	sheetname := "result"
	index := f.NewSheet(sheetname)
	f.SetCellValue(sheetname, "A2", "Hello world.")
	f.SetCellValue(sheetname, "B2", 100)
	f.SetActiveSheet(index)
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
