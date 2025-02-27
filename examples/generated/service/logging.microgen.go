// Code generated by microgen 0.9.0. DO NOT EDIT.

package service

import (
	"context"
	log "github.com/go-kit/kit/log"
	service "github.com/recolabs/microgen/examples/generated"
	"time"
)

// LoggingMiddleware writes params, results and working time of method call to provided logger after its execution.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next service.StringService) service.StringService {
		return &loggingMiddleware{
			logger: logger,
			next:   next,
		}
	}
}

type loggingMiddleware struct {
	logger log.Logger
	next   service.StringService
}

func (M loggingMiddleware) Uppercase(arg0 context.Context, arg1 map[string]string) (res0 string, res1 error) {
	defer func(begin time.Time) {
		M.logger.Log(
			"method", "Uppercase",
			"request", logUppercaseRequest{StringsMap: arg1},
			"took", time.Since(begin))
	}(time.Now())
	return M.next.Uppercase(arg0, arg1)
}

func (M loggingMiddleware) Count(arg0 context.Context, arg1 string, arg2 string) (res0 int, res1 []int, res2 error) {
	defer func(begin time.Time) {
		M.logger.Log(
			"method", "Count",
			"request", logCountRequest{
				Symbol: arg2,
				Text:   arg1,
			},
			"response", logCountResponse{
				Count:     res0,
				Positions: res1,
			},
			"err", res2,
			"took", time.Since(begin))
	}(time.Now())
	return M.next.Count(arg0, arg1, arg2)
}

func (M loggingMiddleware) TestCase(arg0 context.Context, arg1 []*service.Comment) (res0 map[string]int, res1 error) {
	defer func(begin time.Time) {
		M.logger.Log(
			"method", "TestCase",
			"request", logTestCaseRequest{
				Comments:    arg1,
				LenComments: len(arg1),
			},
			"response", logTestCaseResponse{Tree: res0},
			"err", res1,
			"took", time.Since(begin))
	}(time.Now())
	return M.next.TestCase(arg0, arg1)
}

func (M loggingMiddleware) DummyMethod(arg0 context.Context) (res0 error) {
	defer func(begin time.Time) {
		M.logger.Log(
			"method", "DummyMethod",
			"err", res0,
			"took", time.Since(begin))
	}(time.Now())
	return M.next.DummyMethod(arg0)
}

func (M loggingMiddleware) IgnoredMethod() {
	M.next.IgnoredMethod()
}

func (M loggingMiddleware) IgnoredErrorMethod() (res0 error) {
	return M.next.IgnoredErrorMethod()
}

type (
	logUppercaseRequest struct {
		StringsMap map[string]string
	}
	logCountRequest struct {
		Text   string
		Symbol string
	}
	logCountResponse struct {
		Count     int
		Positions []int
	}
	logTestCaseRequest struct {
		Comments    []*service.Comment
		LenComments int `json:"len(Comments)"`
	}
	logTestCaseResponse struct {
		Tree map[string]int
	}
)
