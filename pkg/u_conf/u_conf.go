// Универсальный пакет для заполнения структур настроек подключаемых пакетов
// из единого конфигурационного файла
package u_conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"
)

var (
	bConf     []byte
	confmap   map[string]interface{}
	commonmap map[string]interface{}
)

// Получаем содержимое файла конфига
func SetConfigFile(fileName string) error {
	var err error
	bConf, err = ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	return parseConfigData(bConf)
}

// получаем данные из json
func parseConfigData(confData []byte) error {
	err := json.Unmarshal(confData, &confmap)
	if err != nil {
		return err
	}
	if confmap["common"] != nil {
		var ok bool
		commonmap, ok = confmap["common"].(map[string]interface{})
		if !ok {
			return errors.New("error in common json block")
		}
	}
	return nil
}

// ##Заполнение структуры конфига
/*
	packStruct - указатель на структуру конфига
	blockName - имя блока в json конфиге, если передать пустым, в качестве блока берется имя пакета в котором определена структура packStruct
	В json файле может быть определен необязательный блок common. В случае, если поле для структуры не найдено в блоке для конкретного пакета,
	это поле будет заполнено значением из поля с таким же именем блока common, при условии совпадения типа.
*/
func ParsePackageConfig(packStruct interface{}, blockName string) error {
	var err error

	structVal := reflect.ValueOf(packStruct).Elem()
	pack_path := structVal.Type().PkgPath()

	pack := blockName
	// Если передать пустое имя блока в конфиге
	// имя будет взято из полного наименования пакета
	if len(pack) == 0 {
		pack_slice := strings.Split(pack_path, "/")
		pack = pack_slice[len(pack_slice)-1]
	}

	if V, ok := confmap[pack]; ok {
		err = switchSetType(structVal, V, structVal.Type(), "", pack_path, reflect.Value{})
	} else {
		err = fmt.Errorf("no have block <%s> in config file", pack)
	}

	return err
}

// Рекурсивная функция заполнения полей конфига
func switchSetType(field reflect.Value, json_value interface{}, field_type reflect.Type, field_tag reflect.StructTag, pack_path string, map_key reflect.Value) error {
	var err error
	switch field_type.Kind() {
	case reflect.Slice:
		if json_value != nil {
			v_slice := reflect.ValueOf(json_value)
			t_slice := field_type.Elem()
			slice := reflect.MakeSlice(reflect.SliceOf(t_slice), v_slice.Len(), v_slice.Cap())
			for j := 0; j < v_slice.Len(); j++ {
				slice.Index(j).Set(reflect.Zero(t_slice))
				err := switchSetType(slice.Index(j), v_slice.Index(j).Interface(), t_slice, field_tag, pack_path, reflect.Value{})
				if err != nil {
					return err
				}
			}
			if field.Type().Kind() == reflect.Map {
				field.SetMapIndex(map_key, slice)
			} else {
				field.Set(slice)
			}
		}
	case reflect.Map:
		if json_value != nil {
			v_map := reflect.ValueOf(json_value)
			field.Set(reflect.MakeMap(field_type))

			for _, k := range v_map.MapKeys() {
				val_json := v_map.MapIndex(k)
				n_key := reflect.ValueOf(k.Interface().(string))
				field.SetMapIndex(n_key, reflect.Zero(field_type.Elem()))
				err := switchSetType(field, val_json.Interface(), field_type.Elem(), field_tag, pack_path, n_key)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Struct:
		var structmap map[string]interface{}
		var ok bool
		structmap, ok = json_value.(map[string]interface{})
		if !ok {
			return typeError(field_type, pack_path)
		}
		for i := 0; i < field_type.NumField(); i++ {
			tag := field_type.Field(i).Tag.Get("conf")
			// Если тэга conf нет у поля структуры либо там стоит "-",
			// то это поле пропускаем
			if tag == "" || tag == "-" {
				continue
			}
			json_value_child := structmap[tag]
			field_type_child := field_type.Field(i).Type
			// Если у структуре указатель, то поле может отсутствовать в json, иначе возвращаем ошибку.
			if json_value_child == nil && field_type_child.Kind() != reflect.Ptr {
				// Ищем такой тэг в общем(common) блоке
				json_value_child = commonmap[tag]
				if json_value_child == nil {
					return fmt.Errorf("not option(or value is null) %s for package %s", tag, pack_path)
				}
			}
			field_tag_child := field_type.Field(i).Tag
			err = switchSetType(field.Field(i), json_value_child, field_type_child, field_tag_child, pack_path, reflect.Value{})
			if err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if json_value != nil {
			field_type_child := field.Type()
			field.Set(reflect.New(field_type_child.Elem()))
			err = switchSetType(field.Elem(), json_value, field_type_child.Elem(), field_tag, pack_path, reflect.Value{})
			if err != nil {
				return err
			}
		}
	default:
		val, ok := simpleType(field_type.Kind(), field_type, json_value, field_tag)
		if !ok {
			return typeError(field_type, pack_path)
		}
		if field.CanAddr() && field.Type().Kind() != reflect.Map {
			field.Set(val)
		} else if field.Type().Kind() == reflect.Map {
			field.SetMapIndex(map_key, val)
		}
	}
	return err
}

// Получение значений для простых типов
func simpleType(kind reflect.Kind, field_type reflect.Type, json_value interface{}, field_tag reflect.StructTag) (reflect.Value, bool) {
	var value reflect.Value
	switch kind {
	case reflect.String:
		val, ok := json_value.(string)
		if !ok {
			return value, ok
		}
		value = reflect.ValueOf(val)
	case reflect.Uint:
		fl, ok := json_value.(float64)
		if !ok {
			return value, ok
		}
		val := uint(fl)
		value = reflect.ValueOf(val)
	case reflect.Uint64:
		fl, ok := json_value.(float64)
		if !ok {
			return value, ok
		}
		val := uint64(fl)
		value = reflect.ValueOf(val)
	case reflect.Int:
		fl, ok := json_value.(float64)
		if !ok {
			return value, ok
		}
		val := int(fl)
		value = reflect.ValueOf(val)
	case reflect.Int64:
		switch v := json_value.(type) {
		case float64:
			if field_type.PkgPath()+`.`+field_type.Name() == "time.Duration" {
				d_name := field_tag.Get("time")
				d := multiplyDuration(time.Duration(int64(v)), d_name)
				value = reflect.ValueOf(d)
			} else {
				value = reflect.ValueOf(int64(v))
			}
		case string:
			if field_type.PkgPath()+`.`+field_type.Name() == "time.Duration" {
				t, err := time.ParseDuration(v)
				if err != nil {
					return value, false
				}
				value = reflect.ValueOf(t)
			}
		}
	case reflect.Bool:
		b, ok := json_value.(bool)
		if !ok {
			return value, ok
		}
		value = reflect.ValueOf(b)
	default:
		return value, false
	}
	return value, true
}

// Возможность в конфиге задавать время в разных единицах.
func multiplyDuration(duration time.Duration, multiplier string) time.Duration {
	switch multiplier {
	case "Microsecond":
		return duration * time.Microsecond
	case "Millisecond":
		return duration * time.Millisecond
	case "Second":
		return duration * time.Second
	case "Minute":
		return duration * time.Minute
	case "Hour":
		return duration * time.Hour
	default:
		return duration
	}
}

// Ошибка, если значение в json не соответсвует типу в структуре.
func typeError(field_type reflect.Type, pack_path string) error {
	t := field_type.String()
	err := fmt.Errorf("in conf file does not match type %s for package %s", t, pack_path)
	return err
}
