// Code generated by microgen. DO NOT EDIT.

package transport

import (
	"context"
	service "github.com/devimteam/microgen/examples/usersvc/pkg/usersvc"
	endpoint "github.com/go-kit/kit/endpoint"
	metrics "github.com/go-kit/kit/metrics"
	"strconv"
	"time"
)

var _ service.UserService = &Endpoints{}

type Endpoints struct {
	CreateUser_Endpoint      endpoint.Endpoint
	UpdateUser_Endpoint      endpoint.Endpoint
	GetUser_Endpoint         endpoint.Endpoint
	FindUsers_Endpoint       endpoint.Endpoint
	CreateComment_Endpoint   endpoint.Endpoint
	GetComment_Endpoint      endpoint.Endpoint
	GetUserComments_Endpoint endpoint.Endpoint
}

func (E Endpoints) CreateUser(arg_0 context.Context, arg_1 service.User) (res_0 string, res_1 error) {
	request := CreateUser_Request{User: arg_1}
	response, res_1 := E.CreateUser_Endpoint(arg_0, &request)
	return response.(*CreateUser_Response).Id, res_1
}
func (E Endpoints) UpdateUser(arg_0 context.Context, arg_1 service.User) (res_0 error) {
	request := UpdateUser_Request{User: arg_1}
	_, res_0 = E.UpdateUser_Endpoint(arg_0, &request)
	return res_0
}
func (E Endpoints) GetUser(arg_0 context.Context, arg_1 string) (res_0 service.User, res_1 error) {
	request := GetUser_Request{Id: arg_1}
	response, res_1 := E.GetUser_Endpoint(arg_0, &request)
	return response.(*GetUser_Response).User, res_1
}
func (E Endpoints) FindUsers(arg_0 context.Context) (res_0 []*service.User, res_1 error) {
	request := FindUsers_Request{}
	response, res_1 := E.FindUsers_Endpoint(arg_0, &request)
	return response.(*FindUsers_Response).Results, res_1
}
func (E Endpoints) CreateComment(arg_0 context.Context, arg_1 service.Comment) (res_0 string, res_1 error) {
	request := CreateComment_Request{Comment: arg_1}
	response, res_1 := E.CreateComment_Endpoint(arg_0, &request)
	return response.(*CreateComment_Response).Id, res_1
}
func (E Endpoints) GetComment(arg_0 context.Context, arg_1 string) (res_0 service.Comment, res_1 error) {
	request := GetComment_Request{Id: arg_1}
	response, res_1 := E.GetComment_Endpoint(arg_0, &request)
	return response.(*GetComment_Response).Comment, res_1
}
func (E Endpoints) GetUserComments(arg_0 context.Context, arg_1 string) (res_0 []service.Comment, res_1 error) {
	request := GetUserComments_Request{UserId: arg_1}
	response, res_1 := E.GetUserComments_Endpoint(arg_0, &request)
	return response.(*GetUserComments_Response).List, res_1
}
func Latency(dur metrics.Histogram) func(endpoints Endpoints) Endpoints {
	return func(endpoints Endpoints) Endpoints {
		return Endpoints{
			CreateComment_Endpoint:   latency(dur, "CreateComment")(endpoints.CreateComment_Endpoint),
			CreateUser_Endpoint:      latency(dur, "CreateUser")(endpoints.CreateUser_Endpoint),
			FindUsers_Endpoint:       latency(dur, "FindUsers")(endpoints.FindUsers_Endpoint),
			GetComment_Endpoint:      latency(dur, "GetComment")(endpoints.GetComment_Endpoint),
			GetUserComments_Endpoint: latency(dur, "GetUserComments")(endpoints.GetUserComments_Endpoint),
			GetUser_Endpoint:         latency(dur, "GetUser")(endpoints.GetUser_Endpoint),
			UpdateUser_Endpoint:      latency(dur, "UpdateUser")(endpoints.UpdateUser_Endpoint),
		}
	}
}
func latency(dur metrics.Histogram, methodName string) endpoint.Middleware {
	dur := dur.With("method", methodName)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (request interface{}, err error) {
			defer func(begin time.Time) {
				dur.With("success", strconv.FormatBool(err == nil)).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return next(ctx, request)
		}
	}
}
