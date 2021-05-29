// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package utils

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

var daoSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	daoSourceDir = regexp.MustCompile(`utils.utils\.go`).ReplaceAllString(file, "")
}

func FileWithLineNum() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)

		if ok && (!strings.HasPrefix(file, daoSourceDir) || strings.HasSuffix(file, "_test.go")) {
			idx := strings.Index(file, "src/")
			if idx != -1 {
				file = file[idx+4:]
			}
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}
	return ""
}

func IsValidDBNameChar(c rune) bool {
	return !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '.' && c != '*' && c != '_' && c != '$' && c != '@'
}

func CheckTruth(val interface{}) bool {
	if v, ok := val.(bool); ok {
		return v
	}

	if v, ok := val.(string); ok {
		v = strings.ToLower(v)
		return v != "false"
	}

	return !reflect.ValueOf(val).IsZero()
}

func ToStringKey(values ...interface{}) string {
	results := make([]string, len(values))

	for idx, value := range values {
		if valuer, ok := value.(driver.Valuer); ok {
			value, _ = valuer.Value()
		}

		switch v := value.(type) {
		case string:
			results[idx] = v
		case []byte:
			results[idx] = string(v)
		case uint:
			results[idx] = strconv.FormatUint(uint64(v), 10)
		default:
			results[idx] = fmt.Sprint(reflect.Indirect(reflect.ValueOf(v)).Interface())
		}
	}

	return strings.Join(results, "_")
}

func AssertEqual(src, dst interface{}) bool {
	if !reflect.DeepEqual(src, dst) {
		if valuer, ok := src.(driver.Valuer); ok {
			src, _ = valuer.Value()
		}

		if valuer, ok := dst.(driver.Valuer); ok {
			dst, _ = valuer.Value()
		}

		return reflect.DeepEqual(src, dst)
	}
	return true
}

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	}
	return ""
}
