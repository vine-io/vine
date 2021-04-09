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

package handler

import (
	"context"
	"time"

	"github.com/lack-io/vine"
	"github.com/lack-io/vine/proto/apis/errors"
	pb "github.com/lack-io/vine/proto/services/runtime"
	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime"
)

type Runtime struct {
	// The runtime used to manage services
	Runtime runtime.Runtime
	// The client used to publish events
	Client vine.Event
}

func (r *Runtime) Read(ctx context.Context, req *pb.ReadRequest, rsp *pb.ReadResponse) error {
	options := toReadOptions(ctx, req.Options)

	services, err := r.Runtime.Read(options...)
	if err != nil {
		return errors.InternalServerError("go.vine.runtime", err.Error())
	}

	for _, service := range services {
		rsp.Services = append(rsp.Services, toProto(service))
	}

	return nil
}

func (r *Runtime) Create(ctx context.Context, req *pb.CreateRequest, rsp *pb.CreateResponse) error {
	if req.Service == nil {
		return errors.BadRequest("go.vine.runtime", "blank service")
	}

	options := toCreateOptions(ctx, req.Options)
	service := toService(req.Service)

	log.Infof("Creating service %s version %s source %s", service.Name, service.Version, service.Source)

	if err := r.Runtime.Create(service, options...); err != nil {
		return errors.InternalServerError("go.vine.runtime", err.Error())
	}

	// publish the create event
	r.Client.Publish(ctx, &pb.Event{
		Type:      "create",
		Timestamp: time.Now().Unix(),
		Service:   req.Service.Name,
		Version:   req.Service.Version,
	})

	return nil
}

func (r *Runtime) Update(ctx context.Context, req *pb.UpdateRequest, rsp *pb.UpdateResponse) error {
	if req.Service == nil {
		return errors.BadRequest("go.vine.runtime", "blank service")
	}

	service := toService(req.Service)
	options := toUpdateOptions(ctx, req.Options)

	log.Infof("Updating service %s version %s source %s", service.Name, service.Version, service.Source)

	if err := r.Runtime.Update(service, options...); err != nil {
		return errors.InternalServerError("go.vine.runtime", err.Error())
	}

	// publish the update event
	r.Client.Publish(ctx, &pb.Event{
		Type:      "update",
		Timestamp: time.Now().Unix(),
		Service:   req.Service.Name,
		Version:   req.Service.Version,
	})

	return nil
}

func (r *Runtime) Delete(ctx context.Context, req *pb.DeleteRequest, rsp *pb.DeleteResponse) error {
	if req.Service == nil {
		return errors.BadRequest("go.vine.runtime", "blank service")
	}

	service := toService(req.Service)
	options := toDeleteOptions(ctx, req.Options)

	log.Infof("Deleting service %s version %s source %s", service.Name, service.Version, service.Source)

	if err := r.Runtime.Delete(service, options...); err != nil {
		return errors.InternalServerError("go.vine.runtime", err.Error())
	}

	// publish the delete event
	r.Client.Publish(ctx, &pb.Event{
		Type:      "delete",
		Timestamp: time.Now().Unix(),
		Service:   req.Service.Name,
		Version:   req.Service.Version,
	})

	return nil
}

func (r *Runtime) Logs(ctx context.Context, req *pb.LogsRequest, stream pb.Runtime_LogsStream) error {
	opts := toLogsOptions(ctx, req.Options)

	// options passed in the request
	if req.GetCount() > 0 {
		opts = append(opts, runtime.LogsCount(req.GetCount()))
	}
	if req.GetStream() {
		opts = append(opts, runtime.LogsStream(req.GetStream()))
	}

	logStream, err := r.Runtime.Logs(&runtime.Service{
		Name: req.GetService(),
	}, opts...)
	if err != nil {
		return err
	}
	defer logStream.Stop()
	defer stream.Close()

	recordChan := logStream.Chan()
	for {
		select {
		case record, ok := <-recordChan:
			if !ok {
				return logStream.Error()
			}
			// send record
			if err := stream.Send(&pb.LogRecord{
				//Timestamp: record.Timestamp.Unix(),
				Metadata: record.Metadata,
				Message:  record.Message,
			}); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}
