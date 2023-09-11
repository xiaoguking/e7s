package e7s

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

//生成32位md5字串
func getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func uniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getMd5String(base64.URLEncoding.EncodeToString(b))
}

// StructToURLValues ToURLValues 将结构体转换为url.Values
func StructToURLValues(data map[string]interface{}, key string) string {

	if _, ok := data[key]; !ok {
		return ""
	}
	types := reflect.TypeOf(data[key])

	value := reflect.ValueOf(data[key])

	switch types.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(value.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(value.Bool())
	case reflect.Slice, reflect.Map, reflect.Struct:
		buf, _ := json.Marshal(value.Interface())
		return string(buf)
	default:
		// 其他类型使用fmt.Sprint方法将其转换为字符串
		return fmt.Sprint(value.Interface())
	}
}
