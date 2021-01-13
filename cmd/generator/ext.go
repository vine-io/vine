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

type ServiceDescriptor struct {
	Proto *descriptor.ServiceDescriptorProto

	Methods []*MethodDescriptor

	TagComments []string
}

type MethodDescriptor struct {
	Proto *descriptor.MethodDescriptorProto

	TagsComments []string
}

type MessageDescriptor struct {
	Proto *Descriptor

	Fields []*FieldDescriptor

	TagComments []string
}

type FieldDescriptor struct {
	Proto *descriptor.FieldDescriptorProto

	TagComments []string
}

func extractFileDescriptor(file *FileDescriptor) {
	file.messages = make([]*MessageDescriptor, 0)
	for _, item := range file.desc {
		md := &MessageDescriptor{
			Proto:       item,
			Fields:      []*FieldDescriptor{},
			TagComments: []string{},
		}
		for _, field := range item.Field {
			md.Fields = append(md.Fields, &FieldDescriptor{
				Proto:       field,
				TagComments: []string{},
			})
		}
		file.messages = append(file.messages, md)
	}

	file.tagServices = make([]*ServiceDescriptor, 0)
	for _, service := range file.Service {
		sv := &ServiceDescriptor{
			Proto:       service,
			Methods:     []*MethodDescriptor{},
			TagComments: []string{},
		}
		for _, method := range service.Method {
			sv.Methods = append(sv.Methods, &MethodDescriptor{
				Proto:        method,
				TagsComments: []string{},
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
				file.tagServices[index].TagComments = parseTagComment(comment)
			case 4:
				index, _ := strconv.Atoi(parts[1])
				mIndex, _ := strconv.Atoi(parts[3])
				file.tagServices[index].Methods[mIndex].TagsComments = parseTagComment(comment)
			case 6:
				//
			}
		case MessageType:
			switch len(parts) {
			// message comment
			case 2:
				index, _ := strconv.Atoi(parts[1])
				file.messages[index].TagComments = parseTagComment(comment)
			// message field comment
			case 4:
				index, _ := strconv.Atoi(parts[1])
				fIndex, _ := strconv.Atoi(parts[3])
				file.messages[index].Fields[fIndex].TagComments = parseTagComment(comment)
			case 6:
				//
			}
		}
	}
}

func parseTagComment(comment *descriptor.SourceCodeInfo_Location) []string {
	comments := make([]string, 0)
	if comment.LeadingComments == nil {
		return comments
	}
	for _, item := range strings.Split(*comment.LeadingComments, "\n") {
		text := strings.TrimSpace(item)
		if strings.HasPrefix(text, "+") {
			comments = append(comments, text)
		}
	}
	return comments
}
