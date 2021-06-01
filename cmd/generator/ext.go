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

package generator

import (
	"strconv"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

const (
	PackageType = 12
	OptionType  = 8
	ServiceType = 6
	MessageType = 4
)

const NilTag = "nil"

type ServiceDescriptor struct {
	Proto *descriptor.ServiceDescriptorProto

	Methods []*MethodDescriptor

	Comments []*Comment
}

type MethodDescriptor struct {
	Proto *descriptor.MethodDescriptorProto

	Comments []*Comment
}

type MessageDescriptor struct {
	Proto *Descriptor

	Fields []*FieldDescriptor

	Comments []*Comment
}

type FieldDescriptor struct {
	Proto *descriptor.FieldDescriptorProto

	Comments []*Comment
}

type Comment struct {
	Tag  string
	Text string
}

func extractFileDescriptor(file *FileDescriptor) {
	file.messages = make([]*MessageDescriptor, 0)
	for _, item := range file.desc {
		// ignore MapEntry message
		if strings.HasSuffix(item.GetName(), "Entry") {
			continue
		}
		md := &MessageDescriptor{
			Proto:    item,
			Fields:   []*FieldDescriptor{},
			Comments: []*Comment{},
		}
		for _, field := range item.Field {
			md.Fields = append(md.Fields, &FieldDescriptor{
				Proto:    field,
				Comments: []*Comment{},
			})
		}
		file.messages = append(file.messages, md)
	}

	file.tagServices = make([]*ServiceDescriptor, 0)
	for _, service := range file.Service {
		sv := &ServiceDescriptor{
			Proto:    service,
			Methods:  []*MethodDescriptor{},
			Comments: []*Comment{},
		}
		for _, method := range service.Method {
			sv.Methods = append(sv.Methods, &MethodDescriptor{
				Proto:    method,
				Comments: []*Comment{},
			})
		}
		file.tagServices = append(file.tagServices, sv)
	}

	for path, comment := range file.comments {
		parts := strings.Split(path, ",")
		if len(parts) == 0 {
			continue
		}
		first, _ := strconv.Atoi(parts[0])
		switch first {
		case PackageType:
		case OptionType:
		case ServiceType:
			switch len(parts) {
			// service comments
			case 2:
				index, _ := strconv.Atoi(parts[1])
				if len(file.tagServices) > index {
					file.tagServices[index].Comments = parseTagComment(comment)
				}
			// service inner method comments
			case 4:
				index, _ := strconv.Atoi(parts[1])
				mIndex, _ := strconv.Atoi(parts[3])
				if len(file.tagServices) > index && len(file.tagServices[index].Methods) > mIndex {
					file.tagServices[index].Methods[mIndex].Comments = parseTagComment(comment)
				}
			case 6:
				// do nothing
			}
		case MessageType:
			switch len(parts) {
			// message comment
			case 2:
				index, _ := strconv.Atoi(parts[1])
				if len(file.messages) > index {
					file.messages[index].Comments = parseTagComment(comment)
				}
			// message field comment
			case 4:
				index, _ := strconv.Atoi(parts[1])
				fIndex, _ := strconv.Atoi(parts[3])
				if len(file.messages) > index && len(file.messages[index].Fields) > fIndex {
					file.messages[index].Fields[fIndex].Comments = parseTagComment(comment)
				}
			case 6:
				// do nothing
			}
		}
	}
}

func parseTagComment(comment *descriptor.SourceCodeInfo_Location) []*Comment {
	comments := make([]*Comment, 0)
	if comment.LeadingComments == nil {
		return comments
	}
	for _, item := range strings.Split(comment.GetLeadingComments(), "\n") {
		var tag string
		text := strings.TrimSpace(item)
		if len(text) == 0 {
			continue
		}
		if strings.HasPrefix(text, "+") {
			if i := strings.Index(text, ":"); i > 0 {
				tag = text[1:i]
				text = text[i+1:]
			} else {
				tag = NilTag
				text = text[1:]
			}
		}
		comments = append(comments, &Comment{
			Tag:  tag,
			Text: text,
		})
	}
	return comments
}

type FileOutPut struct {
	Package       string
	Out           string
	SourcePkgPath string
	Load          bool
}

func (g *Generator) extractFileOutFile(file *FileDescriptor) (output *FileOutPut) {
	output = &FileOutPut{}
	for path, comment := range file.comments {
		parts := strings.Split(path, ",")
		if len(parts) == 0 {
			continue
		}
		first, _ := strconv.Atoi(parts[0])
		switch first {
		case PackageType:
			for _, comment := range parseTagComment(comment) {
				if comment.Tag != g.name {
					continue
				}
				text := comment.Text
				parts := strings.Split(text, "=")
				if len(parts) > 1 {
					if parts[0] == "output" {
						tt := parts[1]
						if idx := strings.Index(parts[1], ";"); idx > 0 {
							output.Out, output.Package = tt[:idx], tt[idx+1:]
						} else {
							output.Out = comment.Text
						}
						output.Load = true
						output.SourcePkgPath = strings.ReplaceAll(file.importPath.String(), "\"", "")
					}
				}
			}
		}
	}
	return
}

// ExtractMessage extract MessageDescriptor by name
func (g *Generator) ExtractMessage(name string) *MessageDescriptor {
	obj := g.ObjectNamed(name)

	for _, f := range g.AllFiles() {
		for _, m := range f.Messages() {
			if m.Proto.GoImportPath() == obj.GoImportPath() {
				for _, item := range obj.TypeName() {
					if item == m.Proto.GetName() {
						m.Proto.file = f
						return m
					}
				}
			}
		}
	}
	return nil
}

// ExtractEnum extract EnumDescriptor by name
func (g *Generator) ExtractEnum(name string) *EnumDescriptor {
	obj := g.ObjectNamed(name)
	for _, f := range g.AllFiles() {
		for _, m := range f.Enums() {
			if m.TypeName()[0] == obj.TypeName()[0] {
				return m
			}
		}
	}
	return nil
}

func isInlineText(text string) bool {
	inline := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(line, "//", ""))
		parts := strings.Split(line, ":")
		if len(parts) > 1 {
			if parts[0] == "+gen" && strings.Contains(parts[1], "inline") {
				inline = true
				break
			}
		}
	}
	return inline
}

func isInlineField(comments []*Comment) bool {
	for _, c := range comments {
		if c.Tag == "gen" && c.Text == "inline" {
			return true
		}
	}
	return false
}

func isMeta(comments []*Comment) bool {
	for _, c := range comments {
		if c.Tag == "gen" && c.Text == "meta" {
			return true
		}
	}
	return false
}
