package main

import (
	"context"
	"time"

	"github.com/zippyra/backend/shared/kafka"
	"github.com/zippyra/backend/shared/logger"
)

type SLAWarningJob struct {
	repo          SupportRepository
	kafkaProducer *kafka.Producer
	ticker        *time.Ticker
	stopCh        chan struct{}
}

func NewSLAWarningJob(repo SupportRepository, kafkaProducer *kafka.Producer) *SLAWarningJob {
	return &SLAWarningJob{
		repo:          repo,
		kafkaProducer: kafkaProducer,
		stopCh:        make(chan struct{}),
	}
}

func (j *SLAWarningJob) Start(interval time.Duration) {
	j.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-j.ticker.C:
				j.RunOnce(context.Background())
			case <-j.stopCh:
				return
			}
		}
	}()
}

func (j *SLAWarningJob) Stop() {
	if j.ticker != nil {
		j.ticker.Stop()
	}
	close(j.stopCh)
}

func (j *SLAWarningJob) RunOnce(ctx context.Context) {
	tickets, err := j.repo.ListTicketsNearSLA(ctx)
	if err != nil {
		logger.Error("[SLAWarningJob] Failed to list tickets near SLA: %v", err)
		return
	}

	for _, t := range tickets {
		if j.kafkaProducer != nil {
			_ = j.kafkaProducer.PublishEvent(ctx, "support.ticket_sla_warning", t.ID, map[string]interface{}{
				"ticket_id":         t.ID,
				"assigned_agent_id": t.AssignedAgentID,
				"category":          t.Category,
				"subject":           t.Subject,
				"sla_due_at":        t.SLADueAt,
			})
		}
		_ = j.repo.MarkSLAWarned(ctx, t.ID)
		logger.Warn("[SLAWarningJob] Published SLA warning for ticket %s (due at: %v)", t.ID, t.SLADueAt)
	}
}
