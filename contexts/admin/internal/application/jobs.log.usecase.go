// Code generated by gowrap. DO NOT EDIT.
// template: templates/slog.html
// gowrap: http://github.com/hexdigest/gowrap

package application

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/go-arrower/skeleton/contexts/admin/internal/domain/jobs"
)

// LoggedJobsApplication implements JobsApplication that is instrumented with logging
type LoggedJobsApplication struct {
	logger *slog.Logger
	next   JobsApplication
}

// NewLoggedJobsApplication instruments an implementation of the JobsApplication with simple logging
func NewLoggedJobsApplication(next JobsApplication, logger *slog.Logger) LoggedJobsApplication {
	return LoggedJobsApplication{
		next:   next,
		logger: logger,
	}
}

// ListAllQueues implements JobsApplication
func (app LoggedJobsApplication) ListAllQueues(ctx context.Context, in ListAllQueuesRequest) (l1 ListAllQueuesResponse, err error) {
	cmdName := commandName(in)

	app.logger.DebugContext(ctx, "executing command",
		slog.String("command", cmdName),
	)

	// result, err := app.next(ctx, in)
	l1, err = app.next.ListAllQueues(ctx, in)

	if err == nil {
		app.logger.DebugContext(ctx, "command executed successfully",
			slog.String("command", cmdName),
		)
	} else {
		app.logger.DebugContext(ctx, "failed to execute command",
			slog.String("command", cmdName),
			slog.String("error", err.Error()),
		)
	}

	return l1, err

}

// Queues implements JobsApplication
func (app LoggedJobsApplication) Queues(ctx context.Context) (q1 jobs.QueueNames, err error) {
	cmdName := "Queues"

	app.logger.DebugContext(ctx, "executing command",
		slog.String("command", cmdName),
	)

	// result, err := app.next(ctx, in)
	q1, err = app.next.Queues(ctx)

	if err == nil {
		app.logger.DebugContext(ctx, "command executed successfully",
			slog.String("command", cmdName),
		)
	} else {
		app.logger.DebugContext(ctx, "failed to execute command",
			slog.String("command", cmdName),
			slog.String("error", err.Error()),
		)
	}

	return q1, err

}

// RescheduleJob implements JobsApplication
func (app LoggedJobsApplication) RescheduleJob(ctx context.Context, in RescheduleJobRequest) (err error) {
	cmdName := commandName(in)

	app.logger.DebugContext(ctx, "executing command",
		slog.String("command", cmdName),
	)

	// result, err := app.next(ctx, in)
	err = app.next.RescheduleJob(ctx, in)

	if err == nil {
		app.logger.DebugContext(ctx, "command executed successfully",
			slog.String("command", cmdName),
		)
	} else {
		app.logger.DebugContext(ctx, "failed to execute command",
			slog.String("command", cmdName),
			slog.String("error", err.Error()),
		)
	}

	return err

}

// ScheduleJobs implements JobsApplication
func (app LoggedJobsApplication) ScheduleJobs(ctx context.Context, in ScheduleJobsRequest) (err error) {
	cmdName := commandName(in)

	app.logger.DebugContext(ctx, "executing command",
		slog.String("command", cmdName),
	)

	// result, err := app.next(ctx, in)
	err = app.next.ScheduleJobs(ctx, in)

	if err == nil {
		app.logger.DebugContext(ctx, "command executed successfully",
			slog.String("command", cmdName),
		)
	} else {
		app.logger.DebugContext(ctx, "failed to execute command",
			slog.String("command", cmdName),
			slog.String("error", err.Error()),
		)
	}

	return err

}

// commandName extracts a printable name from cmd in the format of: functionName.
//
// structName	 								=> strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
// structname	 								=> strings.ToLower(strings.Split(fmt.Sprintf("%T", cmd), ".")[1])
// packageName.structName	 					=> fmt.Sprintf("%T", cmd)
// github.com/go-arrower/skeleton/.../package	=> fmt.Sprintln(reflect.TypeOf(cmd).PkgPath())
// structName is used, the other examples are for inspiration.
// The use case function can not be used, as it is anonymous / a closure returned by the use case constructor.
// Accessing the function name with runtime.Caller(4) will always lead to ".func1".
func commandName(cmd any) string {
	pkgPath := reflect.TypeOf(cmd).PkgPath()

	// example: github.com/go-arrower/skeleton/contexts/admin/internal/application_test
	// take string after /contexts/ and then take string before /internal/
	pkg0 := strings.Split(pkgPath, "/contexts/")

	hasContext := len(pkg0) == 2 //nolint:gomnd
	if hasContext {
		pkg1 := strings.Split(pkg0[1], "/internal/")
		if len(pkg1) == 2 { //nolint:gomnd
			context := pkg1[0]

			return fmt.Sprintf("%s.%T", context, cmd)
		}
	}

	// fallback: if the function is not called from a proper Context => packageName.structName
	return fmt.Sprintf("%T", cmd)
}
