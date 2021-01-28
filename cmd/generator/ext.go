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
			case 2:
				index, _ := strconv.Atoi(parts[1])
				if len(file.tagServices) > index {
					file.tagServices[index].Comments = parseTagComment(comment)
				}
			case 4:
				index, _ := strconv.Atoi(parts[1])
				mIndex, _ := strconv.Atoi(parts[3])
				if len(file.tagServices) > index && len(file.tagServices[index].Methods) > mIndex {
					file.tagServices[index].Methods[mIndex].Comments = parseTagComment(comment)
				}
			case 6:
				//
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
				//
			}
		}
	}
}

func parseTagComment(comment *descriptor.SourceCodeInfo_Location) []*Comment {
	comments := make([]*Comment, 0)
	if comment.LeadingComments == nil {
		return comments
	}
	for _, item := range strings.Split(*comment.LeadingComments, "\n") {
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