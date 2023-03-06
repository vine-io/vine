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

package plugin

import (
	"fmt"
	"strings"

	"github.com/vine-io/vine/cmd/generator"
)

var TagString = "gen"

const (
	// service
	_openapi         = "openapi"
	_termURL         = "term_url"
	_contactName     = "contact_name"
	_contactEmail    = "contact_email"
	_licenseName     = "license_name"
	_licenseUrl      = "license_url"
	_externalDocDesc = "external_doc_desc"
	_externalDocUrl  = "external_doc_url"
	_version         = "version"

	// method tag
	_get      = "get"
	_post     = "post"
	_put      = "put"
	_patch    = "patch"
	_delete   = "delete"
	_entity   = "entity"
	_body     = "body"
	_summary  = "summary"
	_security = "security"
	_result   = "result"

	// field common tag
	_inline    = "inline"
	_required  = "required"
	_default   = "default"
	_example   = "example"
	_in        = "in"
	_enum      = "enum"
	_readOnly  = "ro"
	_writeOnly = "wo"
	//_notIn    = "not_in"

	// string tag
	_minLen   = "min_len"
	_maxLen   = "max_len"
	_email    = "email"
	_date     = "date"
	_dateTime = "date-time"
	_password = "password"
	_byte     = "byte"
	_binary   = "binary"
	_ip       = "ip"
	_ipv4     = "ipv4"
	_ipv6     = "ipv6"
	// _crontab  = "crontab"
	_uuid     = "uuid"
	_uri      = "uri"
	_hostname = "hostname"
	_pattern  = "pattern"

	// int32, int64, uint32, uint64, float32, float64 tag
	// _ne  = "ne"
	// _eq  = "eq"
	_lt  = "lt"
	_lte = "lte"
	_gt  = "gt"
	_gte = "gte"

	// bytes tag
	// _maxBytes = "max_bytes"
	// _minBytes = "min_bytes"
)

type Tag struct {
	Key   string
	Value string
}

// extractTags extracts the maps of *Tag from []*generator.Comment
func (g *vine) extractTags(comments []*generator.Comment) map[string]*Tag {
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
				g.gen.Fail("invalid pattern format")
			} else {
				pa := string(c.Text[i+1])
				pe := string(c.Text[len(c.Text)-1])
				if pa != "`" || pe != "`" {
					g.gen.Fail(fmt.Sprintf("invalid pattern value, pa=%s, pe=%s", pa, pe))
				}
				key := strings.TrimSpace(c.Text[:i])
				value := strings.TrimSpace(c.Text[i+1:])
				if len(value) == 0 {
					g.gen.Fail("tag '%s' missing value", key)
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
					g.gen.Fail(fmt.Sprintf("tag '%s' missing value", tag.Key))
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

// extractDesc extracts descriptions from []*generator.Comment
func extractDesc(comments []*generator.Comment) []string {
	if comments == nil || len(comments) == 0 {
		return nil
	}
	desc := make([]string, 0)
	for _, c := range comments {
		if c.Tag == "" {
			text := strings.TrimSpace(c.Text)
			if len(text) == 0 {
				continue
			}
			desc = append(desc, text)
		}
	}
	return desc
}

func TrimString(s string, c string) string {
	s = strings.TrimPrefix(s, c)
	s = strings.TrimSuffix(s, c)
	return s
}

// fullStringSlice converts [1,2,3] => `"1", "2", "3"`
func fullStringSlice(s string) string {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	out := make([]string, 0)
	for _, a := range parts {
		a = strings.TrimSpace(a)
		if len(a) == 0 {
			continue
		}
		if !strings.HasPrefix(a, "\"") {
			a = "\"" + a
		}
		if !strings.HasSuffix(a, "\"") {
			a = a + "\""
		}
		out = append(out, a)
	}
	return strings.Join(out, ",")
}
