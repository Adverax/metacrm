// Copyright 2019 Adverax. All Rights Reserved.
// This file is part of project
//
//      http://github.com/adverax/echo
//
// Licensed under the MIT (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://github.com/adverax/echo/blob/master/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
)

func ToString(val interface{}) (res string, valid bool) {
	switch v := val.(type) {
	case string:
		return v, true
	case int:
		return strconv.FormatInt(int64(v), 10), true
	case uint:
		return strconv.FormatInt(int64(v), 10), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint8:
		return strconv.FormatInt(int64(v), 10), true
	case uint16:
		return strconv.FormatInt(int64(v), 10), true
	case uint32:
		return strconv.FormatInt(int64(v), 10), true
	case uint64:
		return strconv.FormatInt(int64(v), 10), true
	case float32:
		return strconv.FormatFloat(float64(v), 'e', 8, 64), true
	case float64:
		return strconv.FormatFloat(v, 'e', 8, 64), true
	case bool:
		if v {
			return "1", true
		} else {
			return "0", true
		}
	case []byte:
		return string(v), true
	case json.Number:
		return string(v), true
	case time.Time:
		return v.Format(DateTimeFormat), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res string
		err := ConvertAssign(&res, v)
		if err != nil {
			return "", false
		}
		return res, true
	}
}

func ToTime(val interface{}) (res time.Time, valid bool) {
	switch v := val.(type) {
	case string:
		val, err := time.ParseInLocation(DateTimeFormat, v, time.UTC)
		if err != nil {
			return
		}
		return val, true
	case int64:
		return time.Unix(v, 0), true
	case uint64:
		return time.Unix(int64(v), 0), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		rv := reflect.ValueOf(val)
		switch rv.Kind() {
		case reflect.Int64:
			return time.Unix(rv.Int(), 0), true
		case reflect.Uint64:
			return time.Unix(int64(rv.Uint()), 0), true
		default:
			return
		}
	}
}

func ToInt(val interface{}) (res int, valid bool) {
	switch v := val.(type) {
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case int:
		return int(v), true
	case uint:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return int(vv), true
	case json.Number:
		vv, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, false
		}
		return int(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res int
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToInt8(val interface{}) (res int8, valid bool) {
	switch v := val.(type) {
	case int8:
		return int8(v), true
	case int16:
		return int8(v), true
	case int32:
		return int8(v), true
	case int64:
		return int8(v), true
	case uint8:
		return int8(v), true
	case uint16:
		return int8(v), true
	case uint32:
		return int8(v), true
	case uint64:
		return int8(v), true
	case int:
		return int8(v), true
	case uint:
		return int8(v), true
	case float32:
		return int8(v), true
	case float64:
		return int8(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return 0, false
		}
		return int8(vv), true
	case json.Number:
		vv, err := strconv.ParseInt(string(v), 10, 8)
		if err != nil {
			return 0, false
		}
		return int8(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res int8
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToInt16(val interface{}) (res int16, valid bool) {
	switch v := val.(type) {
	case int8:
		return int16(v), true
	case int16:
		return int16(v), true
	case int32:
		return int16(v), true
	case int64:
		return int16(v), true
	case uint8:
		return int16(v), true
	case uint16:
		return int16(v), true
	case uint32:
		return int16(v), true
	case uint64:
		return int16(v), true
	case int:
		return int16(v), true
	case uint:
		return int16(v), true
	case float32:
		return int16(v), true
	case float64:
		return int16(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return 0, false
		}
		return int16(vv), true
	case json.Number:
		vv, err := strconv.ParseInt(string(v), 10, 16)
		if err != nil {
			return 0, false
		}
		return int16(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res int16
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToInt32(val interface{}) (res int32, valid bool) {
	switch v := val.(type) {
	case int8:
		return int32(v), true
	case int16:
		return int32(v), true
	case int32:
		return int32(v), true
	case int64:
		return int32(v), true
	case uint8:
		return int32(v), true
	case uint16:
		return int32(v), true
	case uint32:
		return int32(v), true
	case uint64:
		return int32(v), true
	case int:
		return int32(v), true
	case uint:
		return int32(v), true
	case float32:
		return int32(v), true
	case float64:
		return int32(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return 0, false
		}
		return int32(vv), true
	case json.Number:
		vv, err := strconv.ParseInt(string(v), 10, 32)
		if err != nil {
			return 0, false
		}
		return int32(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res int32
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToInt64(val interface{}) (res int64, valid bool) {
	switch v := val.(type) {
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case int:
		return int64(v), true
	case uint:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return vv, true
	case json.Number:
		vv, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, false
		}
		return vv, true
	case time.Time:
		return v.Unix(), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res int64
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToUint(val interface{}) (res uint, valid bool) {
	switch v := val.(type) {
	case int8:
		return uint(v), true
	case int16:
		return uint(v), true
	case int32:
		return uint(v), true
	case int64:
		return uint(v), true
	case uint8:
		return uint(v), true
	case uint16:
		return uint(v), true
	case uint32:
		return uint(v), true
	case uint64:
		return uint(v), true
	case int:
		return uint(v), true
	case uint:
		return uint(v), true
	case float32:
		return uint(v), true
	case float64:
		return uint(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(vv), true
	case json.Number:
		vv, err := strconv.ParseUint(string(v), 10, 64)
		if err != nil {
			return 0, false
		}
		return uint(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res uint
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToUint8(val interface{}) (res uint8, valid bool) {
	switch v := val.(type) {
	case int8:
		return uint8(v), true
	case int16:
		return uint8(v), true
	case int32:
		return uint8(v), true
	case int64:
		return uint8(v), true
	case uint8:
		return uint8(v), true
	case uint16:
		return uint8(v), true
	case uint32:
		return uint8(v), true
	case uint64:
		return uint8(v), true
	case int:
		return uint8(v), true
	case uint:
		return uint8(v), true
	case float32:
		return uint8(v), true
	case float64:
		return uint8(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return 0, false
		}
		return uint8(vv), true
	case json.Number:
		vv, err := strconv.ParseUint(string(v), 10, 8)
		if err != nil {
			return 0, false
		}
		return uint8(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res uint8
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToUint16(val interface{}) (res uint16, valid bool) {
	switch v := val.(type) {
	case int8:
		return uint16(v), true
	case int16:
		return uint16(v), true
	case int32:
		return uint16(v), true
	case int64:
		return uint16(v), true
	case uint8:
		return uint16(v), true
	case uint16:
		return uint16(v), true
	case uint32:
		return uint16(v), true
	case uint64:
		return uint16(v), true
	case int:
		return uint16(v), true
	case uint:
		return uint16(v), true
	case float32:
		return uint16(v), true
	case float64:
		return uint16(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return 0, false
		}
		return uint16(vv), true
	case json.Number:
		vv, err := strconv.ParseUint(string(v), 10, 16)
		if err != nil {
			return 0, false
		}
		return uint16(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res uint16
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToUint32(val interface{}) (res uint32, valid bool) {
	switch v := val.(type) {
	case int8:
		return uint32(v), true
	case int16:
		return uint32(v), true
	case int32:
		return uint32(v), true
	case int64:
		return uint32(v), true
	case uint8:
		return uint32(v), true
	case uint16:
		return uint32(v), true
	case uint32:
		return uint32(v), true
	case uint64:
		return uint32(v), true
	case int:
		return uint32(v), true
	case uint:
		return uint32(v), true
	case float32:
		return uint32(v), true
	case float64:
		return uint32(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, false
		}
		return uint32(vv), true
	case json.Number:
		vv, err := strconv.ParseUint(string(v), 10, 32)
		if err != nil {
			return 0, false
		}
		return uint32(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res uint32
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToUint64(val interface{}) (res uint64, valid bool) {
	switch v := val.(type) {
	case int8:
		return uint64(v), true
	case int16:
		return uint64(v), true
	case int32:
		return uint64(v), true
	case int64:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return uint64(v), true
	case int:
		return uint64(v), true
	case uint:
		return uint64(v), true
	case float32:
		return uint64(v), true
	case float64:
		return uint64(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return vv, true
	case json.Number:
		vv, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, false
		}
		return uint64(vv), true
	case time.Time:
		return uint64(v.Unix()), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res uint64
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToFloat32(val interface{}) (res float32, valid bool) {
	switch v := val.(type) {
	case int8:
		return float32(v), true
	case int16:
		return float32(v), true
	case int32:
		return float32(v), true
	case int64:
		return float32(v), true
	case uint8:
		return float32(v), true
	case uint16:
		return float32(v), true
	case uint32:
		return float32(v), true
	case uint64:
		return float32(v), true
	case int:
		return float32(v), true
	case uint:
		return float32(v), true
	case float32:
		return float32(v), true
	case float64:
		return float32(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return 0, false
		}
		return float32(vv), true
	case json.Number:
		vv, err := strconv.ParseFloat(string(v), 32)
		if err != nil {
			return 0, false
		}
		return float32(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res float32
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToFloat64(val interface{}) (res float64, valid bool) {
	switch v := val.(type) {
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case int:
		return float64(v), true
	case uint:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return float64(v), true
	case bool:
		if v {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		vv, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return vv, true
	case json.Number:
		vv, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return 0, false
		}
		return float64(vv), true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res float64
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToBoolean(val interface{}) (res bool, valid bool) {
	switch v := val.(type) {
	case int8:
		return v != 0, true
	case int16:
		return v != 0, true
	case int32:
		return v != 0, true
	case int64:
		return v != 0, true
	case uint8:
		return v != 0, true
	case uint16:
		return v != 0, true
	case uint32:
		return v != 0, true
	case uint64:
		return v != 0, true
	case int:
		return v != 0, true
	case uint:
		return v != 0, true
	case float32:
		return v != 0, true
	case float64:
		return v != 0, true
	case bool:
		return v, true
	//case string:
	//	return v != "", true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res bool
		err := ConvertAssign(&res, v)
		if err != nil {
			return false, false
		}
		return res, true
	}
}

func ToDuration(val interface{}) (res time.Duration, valid bool) {
	switch v := val.(type) {
	case int8:
		return time.Duration(v), true
	case int16:
		return time.Duration(v), true
	case int32:
		return time.Duration(v), true
	case int64:
		return time.Duration(v), true
	case uint8:
		return time.Duration(v), true
	case uint16:
		return time.Duration(v), true
	case uint32:
		return time.Duration(v), true
	case uint64:
		return time.Duration(v), true
	case int:
		return time.Duration(v), true
	case uint:
		return time.Duration(v), true
	case float32:
		return time.Duration(v), true
	case float64:
		return time.Duration(v), true
	case string:
		var err error
		res, err = time.ParseDuration(v)
		if err != nil {
			return 0, false
		}
		return res, true
	case json.RawMessage:
		err := jsonUnmarshal(v, &res)
		return res, err == nil
	default:
		var res time.Duration
		err := ConvertAssign(&res, v)
		if err != nil {
			return 0, false
		}
		return res, true
	}
}

func ToJson(val interface{}) (res json.RawMessage, valid bool) {
	switch v := val.(type) {
	case int8:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case int16:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case int32:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case int64:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case uint8:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case uint16:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case uint32:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case uint64:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case int:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case uint:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case float32:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case float64:
		return json.RawMessage(fmt.Sprintf("%v", v)), true
	case string:
		return json.RawMessage(v), true
	case json.RawMessage:
		return v, true
	default:
		var res json.RawMessage
		res, err := json.Marshal(v)
		//err := ConvertAssign(&res, v)
		if err != nil {
			return nil, false
		}
		return res, true
	}
}

func IsEqualMaps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, val := range a {
		if v, has := b[key]; has {
			if v != val {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

func jsonUnmarshal(data json.RawMessage, value interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()
	return dec.Decode(value)
}

var converters = make(map[reflect.Type]func(value interface{}) (reflect.Value, bool))

func Register(tp reflect.Type, fn func(value interface{}) (reflect.Value, bool)) {
	converters[tp] = fn
}

func To(value interface{}, to reflect.Type) (reflect.Value, bool) {
	if value == nil {
		return reflect.Value{}, false
	}

	switch v := value.(type) {
	case reflect.Value:
		if v.Type().ConvertibleTo(to) {
			return v.Convert(to), true
		}
	default:
		if reflect.TypeOf(value).ConvertibleTo(to) {
			return reflect.ValueOf(value).Convert(to), true
		}
	}

	if fn, ok := converters[to]; ok {
		return fn(value)
	}

	return reflect.Value{}, false
}
