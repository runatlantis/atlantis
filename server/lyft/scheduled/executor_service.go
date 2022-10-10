package scheduled

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

type ExecutorService struct {
	log logging.Logger

	// jobs
	garbageCollector      JobDefinition
	rateLimitPublisher    JobDefinition
	runtimeStatsPublisher JobDefinition
}

func NewExecutorService(
	workingDirIterator events.WorkDirIterator,
	statsScope tally.Scope,
	log logging.Logger,
	closedPullCleaner events.PullCleaner,
	openPullCleaner events.PullCleaner,
	githubClient *vcs.GithubClient,
) *ExecutorService {

	scheduledScope := statsScope.SubScope("scheduled")
	garbageCollector := &GarbageCollector{
		workingDirIterator: workingDirIterator,
		stats:              scheduledScope.SubScope("garbagecollector"),
		log:                log,
		closedPullCleaner:  closedPullCleaner,
		openPullCleaner:    openPullCleaner,
	}

	garbageCollectorJob := JobDefinition{
		Job: garbageCollector,

		Period: 30 * time.Minute,
	}

	rateLimitPublisher := &RateLimitStatsPublisher{
		client: githubClient,
		stats:  scheduledScope.SubScope("ratelimitpublisher"),
		log:    log,
	}

	rateLimitPublisherJob := JobDefinition{
		Job: rateLimitPublisher,

		// since rate limit api doesn't contribute to the rate limit we can call this every minute
		Period: 1 * time.Minute,
	}

	runtimeStatsPublisher := NewRuntimeStats(scheduledScope)

	runtimeStatsPublisherJob := JobDefinition{
		Job:    runtimeStatsPublisher,
		Period: 10 * time.Second,
	}

	return &ExecutorService{
		log:                   log,
		garbageCollector:      garbageCollectorJob,
		rateLimitPublisher:    rateLimitPublisherJob,
		runtimeStatsPublisher: runtimeStatsPublisherJob,
	}
}

type JobDefinition struct {
	Job    Job
	Period time.Duration
}

func (s *ExecutorService) Run() {
	s.log.Info("Scheduled Executor Service started")

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	s.runScheduledJob(ctx, &wg, s.garbageCollector)
	s.runScheduledJob(ctx, &wg, s.rateLimitPublisher)
	s.runScheduledJob(ctx, &wg, s.runtimeStatsPublisher)

	interrupt := make(chan os.Signal, 1)

	// Stop on SIGINTs and SIGTERMs.
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt

	s.log.Warn("Received interrupt. Attempting to Shut down scheduled executor service")

	cancel()
	wg.Wait()

	s.log.Warn("All jobs completed, exiting.")
}

func (s *ExecutorService) runScheduledJob(ctx context.Context, wg *sync.WaitGroup, jd JobDefinition) {
	ticker := time.NewTicker(jd.Period)
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer ticker.Stop()

		// Ensure we recover from any panics to keep the jobs isolated.
		// Keep the recovery outside the select to ensure that we don't infinitely panic.
		defer func() {
			if r := recover(); r != nil {
				s.log.Error(fmt.Sprintf("Recovered from panic: %v", r))
			}
		}()

		for {
			select {
			case <-ctx.Done():
				s.log.Warn("Received interrupt, cancelling job")
				return
			case <-ticker.C:
				jd.Job.Run()
			}
		}
	}()

}

type Job interface {
	Run()
}

type RateLimitStatsPublisher struct {
	log    logging.Logger
	stats  tally.Scope
	client *vcs.GithubClient
}

func (r *RateLimitStatsPublisher) Run() {
	errCounter := r.stats.Counter(metrics.ExecutionErrorMetric)
	rateLimitRemainingCounter := r.stats.Counter("ratelimitremaining")

	rateLimits, err := r.client.GetRateLimits()

	if err != nil {
		errCounter.Inc(1)
		return
	}

	rateLimitRemainingCounter.Inc(int64(rateLimits.GetCore().Remaining))
}

var gcStaleClosedPullTemplate = template.Must(template.New("").Parse(
	"Pull Request has been closed for 30 days. Atlantis GC has deleted the locks and plans for the following projects and workspaces:\n" +
		"{{ range . }}\n" +
		"- dir: `{{ .RepoRelDir }}` {{ .Workspaces }}{{ end }}"))

var gcStaleOpenPullTemplate = template.Must(template.New("").Parse(
	"Pull Request has not been updated for 30 days. Atlantis GC has deleted the locks and plans for the following projects and workspaces:\n" +
		"{{ range . }}\n" +
		"- dir: `{{ .RepoRelDir }}` {{ .Workspaces }}{{ end }}"))

type GCStalePullTemplate struct {
	template *template.Template
}

func NewGCStaleClosedPull() events.PullCleanupTemplate {
	return &GCStalePullTemplate{
		template: gcStaleClosedPullTemplate,
	}
}

func NewGCStaleOpenPull() events.PullCleanupTemplate {
	return &GCStalePullTemplate{
		template: gcStaleOpenPullTemplate,
	}
}

func (t *GCStalePullTemplate) Execute(wr io.Writer, data interface{}) error {
	return t.template.Execute(wr, data)
}

type GarbageCollector struct {
	workingDirIterator events.WorkDirIterator
	stats              tally.Scope
	log                logging.Logger
	closedPullCleaner  events.PullCleaner
	openPullCleaner    events.PullCleaner
}

func (g *GarbageCollector) Run() {
	errCounter := g.stats.Counter(metrics.ExecutionErrorMetric)

	pulls, err := g.workingDirIterator.ListCurrentWorkingDirPulls()

	if err != nil {
		g.log.Error(fmt.Sprintf("error listing pulls %s", err))
		errCounter.Inc(1)
	}

	openPullsCounter := g.stats.Counter("pulls.open")
	updatedthirtyDaysAgoOpenPullsCounter := g.stats.Counter("pulls.open.updated.thirtydaysago")
	closedPullsCounter := g.stats.Counter("pulls.closed")
	fiveMinutesAgoClosedPullsCounter := g.stats.Counter("pulls.closed.fiveminutesago")

	// we can make this shorter, but this allows us to see trends more clearly
	// to determine if there is an issue or not
	thirtyDaysAgo := time.Now().Add(-720 * time.Hour)
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	for _, pull := range pulls {
		if pull.State == models.OpenPullState {
			openPullsCounter.Inc(1)

			if pull.UpdatedAt.Before(thirtyDaysAgo) {
				updatedthirtyDaysAgoOpenPullsCounter.Inc(1)

				g.log.Warn("Pull hasn't been updated for more than 30 days.", map[string]interface{}{
					"repository": pull.BaseRepo.FullName,
					"pull-num":   strconv.Itoa(pull.Num),
				})

				err := g.openPullCleaner.CleanUpPull(pull.BaseRepo, pull)

				if err != nil {
					g.log.Error("Error cleaning up open pulls that haven't been updated in 30 days", map[string]interface{}{
						"repository": pull.BaseRepo.FullName,
						"pull-num":   strconv.Itoa(pull.Num),
						"err":        err,
					})
					errCounter.Inc(1)
					return
				}
			}
			continue
		}

		// assume only other state is closed
		closedPullsCounter.Inc(1)

		// Let's clean up any closed pulls within 5 minutes of closing to ensure that
		// any locks are released.
		if pull.ClosedAt.Before(fiveMinutesAgo) {
			fiveMinutesAgoClosedPullsCounter.Inc(1)

			g.log.Warn("Pull closed for more than 5 minutes but data still on disk", map[string]interface{}{
				"repository": pull.BaseRepo.FullName,
				"pull-num":   strconv.Itoa(pull.Num),
			})

			err := g.closedPullCleaner.CleanUpPull(pull.BaseRepo, pull)

			if err != nil {
				g.log.Error("Error cleaning up 5 minutes old closed pulls", map[string]interface{}{
					"repository": pull.BaseRepo.FullName,
					"pull-num":   strconv.Itoa(pull.Num),
					"err":        err,
				})
				errCounter.Inc(1)
				return
			}
		}
	}
}
