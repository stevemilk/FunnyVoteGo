package util

import (
	"fmt"
	"reflect"
	"time"

	"encoding/json"

	"math/big"
	"strconv"
	"strings"

	"github.com/glog"
)

// Struct2String convert struct to string
func Struct2String(st interface{}) string {
	vt := reflect.TypeOf(st)
	fmt.Println(vt)
	vv := reflect.ValueOf(st)
	fmt.Println(vv)
	var str = ""
	for i := 0; i < vt.NumField(); i++ {
		f := vt.Field(i)
		fmt.Println(f)
		v := vv.Field(i)
		fmt.Println(v)
		chKey := f.Tag.Get("json")
		fmt.Println("kind:", v.Kind())
		switch v.Kind() {
		case reflect.String:
			if s, ok := v.Interface().(string); ok && s != "" {
				str += "\"" + chKey + "\"" + ":" + "\"" + s + "\"" + ","
			}
		case reflect.Int:
			if i, ok := v.Interface().(int); ok && i != 0 {
				s := strconv.Itoa(i)
				str += "\"" + chKey + "\"" + ":" + "\"" + s + "\"" + ","

			}
		case reflect.Uint64:
			if u64, ok := v.Interface().(uint64); ok && u64 != 0 {
				s := strconv.Itoa(int(u64))
				str += "\"" + chKey + "\"" + ":" + "\"" + s + "\"" + ","

			}
		case reflect.Int64:
			if i64, ok := v.Interface().(int64); ok && i64 != 0 {
				s := strconv.Itoa(int(i64))
				str += "\"" + chKey + "\"" + ":" + "\"" + s + "\"" + ","

			}

		case reflect.Uint:
			if u, ok := v.Interface().(uint); ok && u != 0 {
				s := strconv.Itoa(int(u))
				str += "\"" + chKey + "\"" + ":" + "\"" + s + "\"" + ","

			}

		case reflect.Slice:
			fmt.Println("slice: ", v.Interface())
			l := v.Len()
			stri := ""

			for i := 0; i < l; i++ {
				vi := v.Index(i).Interface()
				vtype := reflect.TypeOf(vi)
				fmt.Println(vtype.Name())

				if vtype.Name() == "string" {
					stri += "\"" + vi.(string) + "\"" + ","
				} else if vtype.Name() == "uint" {
					ss := strconv.Itoa(int(vi.(uint)))
					stri += "\"" + ss + "\"" + ","
				} else if vtype.Name() == "int" {
					ss := strconv.Itoa(vi.(int))
					stri += "\"" + ss + "\"" + ","
				}

			}
			stri = stri[:len(stri)-1]
			stri = "[" + stri + "]"
			fmt.Println("stri:", stri)
			str += "\"" + chKey + "\"" + ":" + stri + ","
		default:
			glog.Error("unsupport common query type: " + string(chKey))
			return ""

		}
	}
	str = str[:len(str)-1]
	str = "{" + str + "}"
	fmt.Println(str)
	return str
}

// Struct2Map convert struct to map
func Struct2Map(st interface{}) map[string]interface{} {
	vt := reflect.TypeOf(st)
	vv := reflect.ValueOf(st)
	var data = make(map[string]interface{})
	for i := 0; i < vt.NumField(); i++ {
		f := vt.Field(i)
		v := vv.Field(i)
		chKey := f.Tag.Get("json")
		switch v.Kind() {
		case reflect.String:
			if s, ok := v.Interface().(string); ok && s != "" {
				data[chKey] = s
			}
		case reflect.Int:
			if i, ok := v.Interface().(int); ok && i != 0 {
				data[chKey] = i
			}
		case reflect.Struct:
			if t, ok := v.Interface().(time.Time); ok && t != (time.Time{}) {
				data[chKey] = t
			}
		case reflect.Uint64:
			if u64, ok := v.Interface().(uint64); ok && u64 != 0 {
				data[chKey] = u64
			}
		case reflect.Int64:
			if u64, ok := v.Interface().(int64); ok && u64 != 0 {
				data[chKey] = u64
			}
		case reflect.Uint:
			if u, ok := v.Interface().(uint); ok && u != 0 {
				data[chKey] = u
			}
		case reflect.Float32:
			if u, ok := v.Interface().(float32); ok && u != 0 {
				data[chKey] = u
			}
		case reflect.Float64:
			if u, ok := v.Interface().(float64); ok && u != 0 {
				data[chKey] = u
			}
		case reflect.Bool:
			if u, ok := v.Interface().(bool); ok {
				data[chKey] = u
			}

		default:
			glog.Error("unsupport common query type: " + string(chKey))
		}
	}
	return data
}

// JSONString2Map convert struct to map
func JSONString2Map(str string) (map[string]string, error) {
	result := make(map[string]string)
	err := json.Unmarshal([]byte(str), &result)
	return result, err
}

// JSONString2MapInterface convert struct to map
func JSONString2MapInterface(str string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(str), &result)
	return result, err
}

// Map2JSON conver map to json
func Map2JSON(jsonmap map[string]string) (string, error) {
	jbytes, err := json.Marshal(jsonmap)
	return string(jbytes), err
}

//ABIchangeType change args's type into type in ABI
func ABIchangeType(param []interface{}, arg interface{}, t string) []interface{} {
	//变长则弥补0, 定长直接转换
	if strings.HasPrefix(t, "bytes") {
		if strings.Contains(t, "[]") {
			if len(t) > 7 {
				l := len(t)
				le := t[5 : l-2]
				length, err := strconv.Atoi(le)
				if err != nil {
					glog.Error(err)
				}
				if length > 32 {
					glog.Error("[]bytes too long: ", length)
					return nil
				}
				switch length {
				case 1:
					var bb [][1]byte
					for _, v := range arg.([]interface{}) {
						var b [1]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 2:
					var bb [][2]byte
					for _, v := range arg.([]interface{}) {
						var b [2]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 3:
					var bb [][3]byte
					for _, v := range arg.([]interface{}) {
						var b [3]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 4:
					var bb [][4]byte
					for _, v := range arg.([]interface{}) {
						var b [4]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 5:
					var bb [][5]byte
					for _, v := range arg.([]interface{}) {
						var b [5]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 6:
					var bb [][6]byte
					for _, v := range arg.([]interface{}) {
						var b [6]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 7:
					var bb [][7]byte
					for _, v := range arg.([]interface{}) {
						var b [7]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 8:
					var bb [][8]byte
					for _, v := range arg.([]interface{}) {
						var b [8]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 9:
					var bb [][9]byte
					for _, v := range arg.([]interface{}) {
						var b [9]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 10:
					var bb [][10]byte
					for _, v := range arg.([]interface{}) {
						var b [10]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 11:
					var bb [][11]byte
					for _, v := range arg.([]interface{}) {
						var b [11]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 12:
					var bb [][12]byte
					for _, v := range arg.([]interface{}) {
						var b [12]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 13:
					var bb [][13]byte
					for _, v := range arg.([]interface{}) {
						var b [13]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 14:
					var bb [][14]byte
					for _, v := range arg.([]interface{}) {
						var b [14]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 15:
					var bb [][15]byte
					for _, v := range arg.([]interface{}) {
						var b [15]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 16:
					var bb [][16]byte
					for _, v := range arg.([]interface{}) {
						var b [16]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 17:
					var bb [][17]byte
					for _, v := range arg.([]interface{}) {
						var b [17]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 18:
					var bb [][18]byte
					for _, v := range arg.([]interface{}) {
						var b [18]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 19:
					var bb [][19]byte
					for _, v := range arg.([]interface{}) {
						var b [19]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 20:
					var bb [][20]byte
					for _, v := range arg.([]interface{}) {
						var b [20]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 21:
					var bb [][21]byte
					for _, v := range arg.([]interface{}) {
						var b [21]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 22:
					var bb [][22]byte
					for _, v := range arg.([]interface{}) {
						var b [22]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 23:
					var bb [][23]byte
					for _, v := range arg.([]interface{}) {
						var b [23]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 24:
					var bb [][24]byte
					for _, v := range arg.([]interface{}) {
						var b [24]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 25:
					var bb [][25]byte
					for _, v := range arg.([]interface{}) {
						var b [25]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 26:
					var bb [][26]byte
					for _, v := range arg.([]interface{}) {
						var b [26]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 27:
					var bb [][27]byte
					for _, v := range arg.([]interface{}) {
						var b [27]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 28:
					var bb [][28]byte
					for _, v := range arg.([]interface{}) {
						var b [28]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 29:
					var bb [][29]byte
					for _, v := range arg.([]interface{}) {
						var b [29]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 30:
					var bb [][30]byte
					for _, v := range arg.([]interface{}) {
						var b [30]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 31:
					var bb [][31]byte
					for _, v := range arg.([]interface{}) {
						var b [31]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				case 32:
					var bb [][32]byte
					for _, v := range arg.([]interface{}) {
						var b [32]byte
						copy(b[:], v.(string))
						bb = append(bb, b)
					}
					param = append(param, bb)
					return param
				}

			}
			var bb [][]byte
			for _, v := range arg.([]interface{}) {
				var b []byte
				copy(b[:], v.(string))
				bb = append(bb, b)
			}
			param = append(param, bb)
			return param
		}
		s := arg.(string)
		if len(t) > 5 {
			l := len(t)
			le := t[5:l]
			length, err := strconv.Atoi(le)
			if err != nil {
				glog.Error(err)
			}
			if length > 32 {
				glog.Error("[]bytes too long: ", length)
				return nil
			}
			switch length {
			case 1:
				var b [1]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 2:
				var b [2]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 3:
				var b [3]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 4:
				var b [4]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 5:
				var b [4]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 6:
				var b [6]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 7:
				var b [7]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 8:
				var b [8]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 9:
				var b [9]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 10:
				var b [10]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 11:
				var b [11]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 12:
				var b [12]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 13:
				var b [13]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 14:
				var b [14]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 15:
				var b [15]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 16:
				var b [16]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 17:
				var b [17]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 18:
				var b [18]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 19:
				var b [19]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 20:
				var b [20]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 21:
				var b [21]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 22:
				var b [22]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 23:
				var b [23]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 24:
				var b [24]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 25:
				var b [25]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 26:
				var b [26]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 27:
				var b [27]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 28:
				var b [28]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 29:
				var b [29]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 30:
				var b [30]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 31:
				var b [31]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			case 32:
				var b [32]byte
				copy(b[:], s)
				param = append(param, b)
				return param
			}
		}
		b := []byte(s)
		param = append(param, b)
		return param
	}

	// uint部分,包括数组
	if strings.Contains(t, "uint") {
		if strings.Contains(t, "[]") {
			if len(t) > 6 {
				l := len(t)
				le := t[4 : l-2]
				length, err := strconv.Atoi(le)
				if err != nil {
					glog.Error(err)
				}
				if length > 256 {
					glog.Error("uint too long: ", length)
					return nil
				}
				switch length {
				case 8:
					var tmp []uint8
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]uint8 error: ", err)
							return param
						}
						tmp = append(tmp, uint8(uintNum))
					}
					param = append(param, tmp)
					return param
				case 16:
					var tmp []uint16
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]uint16 error: ", err)
							return param
						}
						tmp = append(tmp, uint16(uintNum))
					}
					param = append(param, tmp)
					return param
				case 32:
					var tmp []uint32
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]uint32 error: ", err)
							return param
						}
						tmp = append(tmp, uint32(uintNum))
					}
					param = append(param, tmp)
					return param
				case 64:
					var tmp []uint64
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]uint64 error: ", err)
							return param
						}
						tmp = append(tmp, uint64(uintNum))
					}
					param = append(param, tmp)
					return param
				default:
					var tmp []*big.Int
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]uintBig error: ", err)
							return param
						}
						tmp = append(tmp, big.NewInt(int64(uintNum)))
					}
					param = append(param, tmp)
					return param
				}

			}
			var tmp []uint
			for _, v := range arg.([]interface{}) {
				uintNum, err := strconv.Atoi(v.(string))
				if err != nil {
					glog.Error("[]uint error: ", err)
					return param
				}
				tmp = append(tmp, uint(uintNum))
			}
			param = append(param, tmp)
			return param

		}
		s := arg.(string)
		uintNum, err := strconv.Atoi(s)
		if err != nil {
			glog.Error(err)
		}
		if len(t) > 4 {
			l := len(t)
			le := t[4:l]
			length, err := strconv.Atoi(le)
			if err != nil {
				glog.Error(err)
			}
			if length > 256 {
				glog.Error("uint too long: ", length)
				return nil
			}
			switch length {
			case 8:
				param = append(param, uint8(uintNum))
				return param
			case 16:
				param = append(param, uint16(uintNum))
				return param
			case 32:
				param = append(param, uint32(uintNum))
				return param
			case 64:
				param = append(param, uint64(uintNum))
				return param
			default:
				u := big.NewInt(int64(uintNum))
				param = append(param, u)
				return param
			}

		}

		param = append(param, uint(uintNum))
		return param
	}

	// int部分,包括数组
	if strings.Contains(t, "int") {
		if strings.Contains(t, "[]") {
			if len(t) > 5 {
				l := len(t)
				le := t[3 : l-2]
				length, err := strconv.Atoi(le)
				if err != nil {
					glog.Error(err)
				}
				if length > 256 {
					glog.Error("int too long: ", length)
					return nil
				}
				switch length {
				case 8:
					var tmp []int8
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]int8 error: ", err)
							return param
						}
						tmp = append(tmp, int8(uintNum))
					}
					param = append(param, tmp)
					return param
				case 16:
					var tmp []int16
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]int16 error: ", err)
							return param
						}
						tmp = append(tmp, int16(uintNum))
					}
					param = append(param, tmp)
					return param
				case 32:
					var tmp []int32
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]int32 error: ", err)
							return param
						}
						tmp = append(tmp, int32(uintNum))
					}
					param = append(param, tmp)
					return param
				case 64:
					var tmp []int64
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]int64 error: ", err)
							return param
						}
						tmp = append(tmp, int64(uintNum))
					}
					param = append(param, tmp)
					return param
				default:
					var tmp []*big.Int
					for _, v := range arg.([]interface{}) {
						uintNum, err := strconv.Atoi(v.(string))
						if err != nil {
							glog.Error("[]uintBig error: ", err)
							return param
						}
						tmp = append(tmp, big.NewInt(int64(uintNum)))
					}
					param = append(param, tmp)
					return param
				}

			}
			var tmp []int
			for _, v := range arg.([]interface{}) {
				uintNum, err := strconv.Atoi(v.(string))
				if err != nil {
					glog.Error("[]int error: ", err)
					return param
				}
				tmp = append(tmp, int(uintNum))
			}
			param = append(param, tmp)
			return param

		}
		s := arg.(string)
		uintNum, err := strconv.Atoi(s)
		if err != nil {
			glog.Error(err)
		}
		if len(t) > 3 {
			l := len(t)
			le := t[3:l]
			length, err := strconv.Atoi(le)
			if err != nil {
				glog.Error(err)
			}
			if length > 256 {
				glog.Error("uint too long: ", length)
				return nil
			}
			switch length {
			case 8:
				param = append(param, int8(uintNum))
				return param
			case 16:
				param = append(param, int16(uintNum))
				return param
			case 32:
				param = append(param, int32(uintNum))
				return param
			case 64:
				param = append(param, int64(uintNum))
				return param
			default:
				u := big.NewInt(int64(uintNum))
				param = append(param, u)
				return param
			}

		}

		param = append(param, int(uintNum))
		return param
	}

	param = append(param, arg.(string))
	return param
}
