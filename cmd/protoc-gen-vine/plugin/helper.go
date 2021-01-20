// Copyright 2021 lack
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	"fmt"
	"strings"

	"github.com/lack-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// service tag
	_get    = "get"
	_post   = "post"
	_put    = "put"
	_patch  = "patch"
	_delete = "delete"
	_body   = "body"

	// message tag
	_ignore = "ignore"

	// field common tag
	_required = "required"
	_default  = "default"
	_in       = "in"
	_enum     = "enum"
	_notIn    = "not_in"

	// string tag
	_minLen   = "min_len"
	_maxLen   = "max_len"
	_prefix   = "prefix"
	_suffix   = "suffix"
	_contains = "contains"
	_number   = "number"
	_email    = "email"
	_ip       = "ip"
	_ipv4     = "ipv4"
	_ipv6     = "ipv6"
	_crontab  = "crontab"
	_uuid     = "uuid"
	_uri      = "uri"
	_domain   = "domain"
	_pattern  = "pattern"

	// int32, int64, uint32, uint64, float32, float64 tag
	_ne  = "ne"
	_eq  = "eq"
	_lt  = "lt"
	_lte = "lte"
	_gt  = "gt"
	_gte = "gte"

	// bytes tag
	_maxBytes = "max_bytes"
	_minBytes = "min_bytes"

	// repeated tag: required, min_len, max_len
	// message tag: required
)

type Tag struct {
	Key   string
	Value string
}

func extractTags(comments []*generator.Comment) map[string]*Tag {
	if comments == nil || len(comments) == 0 {
		return nil
	}
	tags := make(map[string]*Tag, 0)
	for _, c := range comments {
		if c.Tag != TagString || len(c.Text) == 0 {
			continue
		}
		if strings.HasPrefix(c.Text, _pattern) {
			if i := strings.Index(c.Text, "="); i == -1 {
				panic("invalid pattern format")
			} else {
				pa := c.Text[:i]
				pe := c.Text[i+1:]
				if pa != "`" || pe != "`" {
					panic("invalid pattern value")
				}
				key := strings.TrimSpace(pa)
				value := strings.TrimSpace(pe)
				if len(value) == 0 {
					panic(fmt.Sprintf("tag '%s' missing value", key))
				}
				tags[key] = &Tag{
					Key:   key,
					Value: value,
				}
			}
			continue
		}
		parts := strings.Split(c.Text, ";")
		for _, p := range parts {
			tag := new(Tag)
			p = strings.TrimSpace(p)
			if i := strings.Index(p, "="); i > 0 {
				tag.Key = strings.TrimSpace(p[:i])
				v := strings.TrimSpace(p[i+1:])
				if v == "" {
					panic(fmt.Sprintf("tag '%s' missing value", tag.Key))
				}
				tag.Value = v
			} else {
				tag.Key = p
			}
			tags[tag.Key] = tag
		}
	}

	return tags
}
