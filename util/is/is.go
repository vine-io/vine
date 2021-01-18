// Copyright 2021 lack
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package is

import (
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/lack-io/vine/util/sched/cron"
)

type Empty struct{}

const (
	ReEmail = `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`

	ReUuid = `[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}`

	ReDomain = `[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$`

	ReURL = `[a-zA-z]+://[^\s]*`

	ReIPv4 = `^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$`

	ReIPv6 = `^([\da-fA-F]{1,4}:){6}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^::([\da-fA-F]{1,4}:){0,4}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:):([\da-fA-F]{1,4}:){0,3}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){2}:([\da-fA-F]{1,4}:){0,2}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){3}:([\da-fA-F]{1,4}:){0,1}((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){4}:((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$|^([\da-fA-F]{1,4}:){7}[\da-fA-F]{1,4}$|^:((:[\da-fA-F]{1,4}){1,6}|:)$|^[\da-fA-F]{1,4}:((:[\da-fA-F]{1,4}){1,5}|:)$|^([\da-fA-F]{1,4}:){2}((:[\da-fA-F]{1,4}){1,4}|:)$|^([\da-fA-F]{1,4}:){3}((:[\da-fA-F]{1,4}){1,3}|:)$|^([\da-fA-F]{1,4}:){4}((:[\da-fA-F]{1,4}){1,2}|:)$|^([\da-fA-F]{1,4}:){5}:([\da-fA-F]{1,4})?$|^([\da-fA-F]{1,4}:){6}:$`
)

func Email(s string) bool {
	return Re(ReEmail, s)
}

func IPv4(s string) bool {
	return Re(ReIPv4, s)
}

func IPv6(s string) bool {
	return Re(ReIPv6, s)
}

func IP(s string) bool {
	return IPv4(s) || IPv6(s)
}

func Domain(s string) bool {
	return Re(ReDomain, s)
}

func Uuid(s string) bool {
	return Re(ReUuid, s)
}

func Number(s string) bool {
	return Re(`^[0-9]+$`, s)
}

func URL(s string) bool {
	return Re(ReURL, s)
}

func Re(re string, text string) bool {
	ok, _ := regexp.MatchString(re, text)
	return ok
}

func Zero(v interface{}) bool {
	return reflect.ValueOf(v).IsZero()
}

func In(arr interface{}, item interface{}) bool {
	v := reflect.ValueOf(arr)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(v.Index(i).Interface(), item) {
				return true
			}
		}
	}
	return false
}

func NotIn(arr interface{}, item interface{}) bool {
	v := reflect.ValueOf(arr)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if reflect.DeepEqual(v.Index(i).Interface(), item) {
				return false
			}
		}
	}
	return true
}

func Crontab(s string) bool {
	_, err := cron.Parse(s)
	return err == nil
}

func MargeErr(errs ...error) error {
	if errs == nil {
		return nil
	}
	parts := make([]string, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		parts = append(parts, err.Error())
	}
	return errors.New(strings.Join(parts, "\n"))
}
