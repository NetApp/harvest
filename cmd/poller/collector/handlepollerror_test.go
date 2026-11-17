package collector

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/schedule"
	"github.com/netapp/harvest/v2/pkg/errs"
)

func newPollErrorCollector(t *testing.T) (*AbstractCollector, *schedule.Task) {
	t.Helper()
	s := schedule.New()
	err := s.NewTaskString("data", "3m", 0, nil, false, "")
	assert.Nil(t, err)
	c := &AbstractCollector{
		Name:     "Rest",
		Object:   "WebCluster",
		Logger:   slog.New(slog.DiscardHandler),
		Schedule: s,
	}
	return c, s.GetTask("data")
}

// restFetchErr wraps an ONTAP REST error the way the REST collector returns it from PollData:
// FetchAllStream adds "error making request" and handleError adds "failed to fetch data".
func restFetchErr(code errs.OntapRestCode) error {
	restErr := errs.NewRest().StatusCode(404).Message(code.Name).Code(code.Code).API("/api/cluster/web").Build()
	return fmt.Errorf("failed to fetch data: %w", fmt.Errorf("error making request %w", restErr))
}

// An API that does not exist on the cluster's ONTAP version is expected, not a failure.
// The task stands by for an hour instead of entering the failed state.
func TestHandlePollErrorAPINotFoundStandsBy(t *testing.T) {
	c, task := newPollErrorCollector(t)

	retryDelay := c.handlePollError(task, restFetchErr(errs.APINotFound), time.Second, 1)

	assert.Equal(t, retryDelay, 1)
	assert.True(t, c.Schedule.IsStandBy())
	assert.Equal(t, task.GetInterval(), 1*time.Hour)
	assert.Equal(t, c.Status, uint8(1))
	assert.Equal(t, c.Message, errs.APINotFound.Name)
}

// Contrast with APINotFound: a REST error the collector does not recognize marks the collector failed
// and leaves the schedule alone.
func TestHandlePollErrorUnknownRestErrorFails(t *testing.T) {
	c, task := newPollErrorCollector(t)

	unknown := errs.OntapRestCode{Name: "Unexpected argument", Code: 262179}
	retryDelay := c.handlePollError(task, restFetchErr(unknown), time.Second, 1)

	assert.Equal(t, retryDelay, 1)
	assert.False(t, c.Schedule.IsStandBy())
	assert.Equal(t, task.GetInterval(), 3*time.Minute)
	assert.Equal(t, c.Status, uint8(2))
}
