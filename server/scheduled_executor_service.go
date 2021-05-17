package server

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type ScheduledExecutorService struct {
	log logging.SimpleLogging

	// jobs
	garbageCollector CronDefinition
}

func NewScheduledExecutorService(
	workingDirIterator events.WorkDirIterator,
	statsScope stats.Scope,
	log logging.SimpleLogging,
) *ScheduledExecutorService {
	garbageCollector := &GarbageCollector{
		workingDirIterator: workingDirIterator,
		stats:              statsScope.Scope("scheduled.garbagecollector"),
		log:                log,
	}

	garbageCollectorCron := CronDefinition{
		Job: garbageCollector,

		// 5 minutes should probably be the lowest to prevent GH rate limits
		Period: 5 * time.Minute,
	}

	return &ScheduledExecutorService{
		log:              log,
		garbageCollector: garbageCollectorCron,
	}
}

type CronDefinition struct {
	Job    Job
	Period time.Duration
}

func (s *ScheduledExecutorService) Run() {
	s.log.Info("Scheduled Executor Service started")

	// create tickers
	garbageCollectorTicker := time.NewTicker(s.garbageCollector.Period)
	defer garbageCollectorTicker.Stop()

	interrupt := make(chan os.Signal, 1)

	// Stop on SIGINTs and SIGTERMs.
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-interrupt:
			s.log.Warn("Received interrupt. Shutting down scheduled executor service")
			return
		case <-garbageCollectorTicker.C:
			go s.garbageCollector.Job.Run()
		}
	}
}

type Job interface {
	Run()
}

type GarbageCollector struct {
	workingDirIterator events.WorkDirIterator
	stats              stats.Scope
	log                logging.SimpleLogging
}

func (g *GarbageCollector) Run() {
	pulls, err := g.workingDirIterator.ListCurrentWorkingDirPulls()

	if err != nil {
		g.log.Err("error listing pulls %s", err)
	}

	openPullsCounter := g.stats.NewCounter("pulls.open")
	closedPullsCounter := g.stats.NewCounter("pulls.closed")
	thirtyDaysAgoClosedPullsCounter := g.stats.NewCounter("pulls.closed.thirtydaysago")
	fiveMinutesAgoClosedPullsCounter := g.stats.NewCounter("pulls.closed.fiveminutesago")

	// we can make this shorter, but this allows us to see trends more clearly
	// to determine if there is an issue or not
	thirtyDaysAgo := time.Now().Add(-720 * time.Hour)
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	for _, pull := range pulls {
		logger := g.log.With(fmtLogSrc(pull.BaseRepo, pull.Num)...)
		if pull.State == models.OpenPullState {
			logger.Debug("pull #%d is open", pull.Num)
			openPullsCounter.Inc()
			continue
		}

		// assume only other state is closed
		closedPullsCounter.Inc()

		logger.Debug("pull #%d is closed but data still on disk", pull.Num)

		// TODO: update this to actually go ahead and delete things
		if pull.ClosedAt.Before(thirtyDaysAgo) {
			thirtyDaysAgoClosedPullsCounter.Inc()

			logger.Info("Pull closed for more than 30 days but data still on disk")
		}

		// This will allow us to catch leaks as soon as they happen (hopefully)
		if pull.ClosedAt.Before(fiveMinutesAgo) {
			fiveMinutesAgoClosedPullsCounter.Inc()

			logger.Info("Pull closed for more than 5 minutes but data still on disk")
		}
	}
}

// taken from other parts of the code, would be great to have this in a shared spot
func fmtLogSrc(repo models.Repo, pullNum int) []interface{} {
	return []interface{}{
		"repository", repo.FullName,
		"pull-num", strconv.Itoa(pullNum),
	}
}
