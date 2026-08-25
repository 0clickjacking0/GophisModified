package worker

import (
	"context"
	"fmt"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
	"github.com/sirupsen/logrus"
)

// DefaultMailBatchSize is the default number of messages sent over one SMTP
// connection before another batch takes its place.
const DefaultMailBatchSize = 100

// Worker is an interface that defines the operations needed for a background worker
type Worker interface {
	Start()
	LaunchCampaign(c models.Campaign)
	SendTestEmail(s *models.EmailRequest) error
}

// DefaultWorker is the background worker that handles watching for new campaigns and sending emails appropriately.
type DefaultWorker struct {
	mailer        mailer.Mailer
	batchSize     int
	maxConcurrent int
}

// New creates a new worker object to handle the creation of campaigns
func New(options ...func(Worker) error) (Worker, error) {
	w := &DefaultWorker{
		batchSize:     DefaultMailBatchSize,
		maxConcurrent: mailer.DefaultMaxConcurrent,
	}
	for _, opt := range options {
		if err := opt(w); err != nil {
			return nil, err
		}
	}
	if w.mailer == nil {
		w.mailer = mailer.NewMailWorkerWithConcurrency(w.maxConcurrent)
	}
	return w, nil
}

// WithMailer sets the mailer for a given worker.
// By default, workers use a standard, default mailworker.
func WithMailer(m mailer.Mailer) func(*DefaultWorker) error {
	return func(w *DefaultWorker) error {
		w.mailer = m
		return nil
	}
}

// WithBatchSize configures how many messages share one SMTP connection. A
// value of zero keeps the default.
func WithBatchSize(size int) func(Worker) error {
	return func(worker Worker) error {
		if size < 0 {
			return fmt.Errorf("mail batch size must not be negative: %d", size)
		}
		w, ok := worker.(*DefaultWorker)
		if !ok {
			return fmt.Errorf("unsupported worker type %T", worker)
		}
		if size > 0 {
			w.batchSize = size
		}
		return nil
	}
}

// WithMaxConcurrent configures the global number of mail batches that may be
// processed simultaneously. A value of zero keeps the default.
func WithMaxConcurrent(maxConcurrent int) func(Worker) error {
	return func(worker Worker) error {
		if maxConcurrent < 0 {
			return fmt.Errorf("maximum mail concurrency must not be negative: %d", maxConcurrent)
		}
		w, ok := worker.(*DefaultWorker)
		if !ok {
			return fmt.Errorf("unsupported worker type %T", worker)
		}
		if maxConcurrent > 0 {
			w.maxConcurrent = maxConcurrent
		}
		return nil
	}
}

func splitMailBatches(ms []mailer.Mail, size int) [][]mailer.Mail {
	if len(ms) == 0 {
		return nil
	}
	if size <= 0 {
		size = DefaultMailBatchSize
	}
	batches := make([][]mailer.Mail, 0, (len(ms)+size-1)/size)
	for start := 0; start < len(ms); start += size {
		end := start + size
		if end > len(ms) {
			end = len(ms)
		}
		batches = append(batches, ms[start:end])
	}
	return batches
}

func (w *DefaultWorker) queueMailBatches(ms []mailer.Mail) {
	for _, batch := range splitMailBatches(ms, w.batchSize) {
		w.mailer.Queue(batch)
	}
}

// processCampaigns loads maillogs scheduled to be sent before the provided
// time and sends them to the mailer.
func (w *DefaultWorker) processCampaigns(t time.Time) error {
	ms, err := models.GetQueuedMailLogs(t.UTC())
	if err != nil {
		log.Error(err)
		return err
	}
	// Lock the MailLogs (they will be unlocked after processing)
	err = models.LockMailLogs(ms, true)
	if err != nil {
		return err
	}
	campaignCache := make(map[int64]models.Campaign)
	// We'll group the maillogs by campaign ID to (roughly) group
	// them by sending profile. This lets the mailer re-use the Sender
	// instead of having to re-connect to the SMTP server for every
	// email.
	msg := make(map[int64][]mailer.Mail)
	for _, m := range ms {
		// We cache the campaign here to greatly reduce the time it takes to
		// generate the message (ref #1726)
		c, ok := campaignCache[m.CampaignId]
		if !ok {
			c, err = models.GetCampaignMailContext(m.CampaignId, m.UserId)
			if err != nil {
				return err
			}
			campaignCache[c.Id] = c
		}
		m.CacheCampaign(&c)
		msg[m.CampaignId] = append(msg[m.CampaignId], m)
	}

	// Next, we process each group of maillogs in parallel
	for cid, msc := range msg {
		go func(cid int64, msc []mailer.Mail) {
			c := campaignCache[cid]
			if c.Status == models.CampaignQueued {
				err := c.UpdateStatus(models.CampaignInProgress)
				if err != nil {
					log.Error(err)
					return
				}
			}
			log.WithFields(logrus.Fields{
				"num_emails": len(msc),
				"batch_size": w.batchSize,
			}).Info("Sending emails to mailer for processing")
			w.queueMailBatches(msc)
		}(cid, msc)
	}
	return nil
}

// Start launches the worker to poll the database every minute for any pending maillogs
// that need to be processed.
func (w *DefaultWorker) Start() {
	log.Info("Background Worker Started Successfully - Waiting for Campaigns")
	go w.mailer.Start(context.Background())
	for t := range time.Tick(1 * time.Minute) {
		err := w.processCampaigns(t)
		if err != nil {
			log.Error(err)
			continue
		}
	}
}

// LaunchCampaign starts a campaign
func (w *DefaultWorker) LaunchCampaign(c models.Campaign) {
	ms, err := models.GetMailLogsByCampaign(c.Id)
	if err != nil {
		log.Error(err)
		return
	}
	models.LockMailLogs(ms, true)
	// This is required since you cannot pass a slice of values
	// that implements an interface as a slice of that interface.
	mailEntries := []mailer.Mail{}
	currentTime := time.Now().UTC()
	campaignMailCtx, err := models.GetCampaignMailContext(c.Id, c.UserId)
	if err != nil {
		log.Error(err)
		return
	}
	for _, m := range ms {
		// Only send the emails scheduled to be sent for the past minute to
		// respect the campaign scheduling options
		if m.SendDate.After(currentTime) {
			m.Unlock()
			continue
		}
		err = m.CacheCampaign(&campaignMailCtx)
		if err != nil {
			log.Error(err)
			return
		}
		mailEntries = append(mailEntries, m)
	}
	w.queueMailBatches(mailEntries)
}

// SendTestEmail sends a test email
func (w *DefaultWorker) SendTestEmail(s *models.EmailRequest) error {
	go func() {
		ms := []mailer.Mail{s}
		w.mailer.Queue(ms)
	}()
	return <-s.ErrorChan
}
