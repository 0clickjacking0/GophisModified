package worker

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/mailer"
	"github.com/gophish/gophish/models"
)

type logMailer struct {
	queue chan []mailer.Mail
}

func (m *logMailer) Start(ctx context.Context) {}

func (m *logMailer) Queue(ms []mailer.Mail) {
	m.queue <- ms
}

// testContext is context to cover API related functions
type testContext struct {
	config *config.Config
}

func setupTest(t *testing.T) *testContext {
	conf := &config.Config{
		DBName:         "sqlite3",
		DBPath:         ":memory:",
		MigrationsPath: "../db/db_sqlite3/migrations/",
	}
	err := models.Setup(conf)
	if err != nil {
		t.Fatalf("Failed creating database: %v", err)
	}
	ctx := &testContext{}
	ctx.config = conf
	createTestData(t, ctx)
	return ctx
}

func createTestData(t *testing.T, ctx *testContext) {
	ctx.config.TestFlag = true
	// Add a group
	group := models.Group{Name: "Test Group"}
	for i := 0; i < 10; i++ {
		group.Targets = append(group.Targets, models.Target{
			BaseRecipient: models.BaseRecipient{
				Email:     fmt.Sprintf("test%d@example.com", i),
				FirstName: "First",
				LastName:  "Example"}})
	}
	group.UserId = 1
	models.PostGroup(&group)

	// Add a template
	template := models.Template{Name: "Test Template"}
	template.Subject = "Test subject"
	template.Text = "Text text"
	template.HTML = "<html>Test</html>"
	template.UserId = 1
	models.PostTemplate(&template)

	// Add a landing page
	p := models.Page{Name: "Test Page"}
	p.HTML = "<html>Test</html>"
	p.UserId = 1
	models.PostPage(&p)

	// Add a sending profile
	smtp := models.SMTP{Name: "Test Page"}
	smtp.UserId = 1
	smtp.Host = "example.com"
	smtp.FromAddress = "test@test.com"
	models.PostSMTP(&smtp)
}

func setupCampaign(id int) (*models.Campaign, error) {
	// Setup and "launch" our campaign
	// Set the status such that no emails are attempted
	c := models.Campaign{Name: fmt.Sprintf("Test campaign - %d", id)}
	c.UserId = 1
	template, err := models.GetTemplate(1, 1)
	if err != nil {
		return nil, err
	}
	c.Template = template

	page, err := models.GetPage(1, 1)
	if err != nil {
		return nil, err
	}
	c.Page = page

	smtp, err := models.GetSMTP(1, 1)
	if err != nil {
		return nil, err
	}
	c.SMTP = smtp

	group, err := models.GetGroup(1, 1)
	if err != nil {
		return nil, err
	}
	c.Groups = []models.Group{group}
	err = models.PostCampaign(&c, c.UserId)
	if err != nil {
		return nil, err
	}
	err = c.UpdateStatus(models.CampaignEmailsSent)
	return &c, err
}

func TestMailLogGrouping(t *testing.T) {
	setupTest(t)

	// Create the campaigns and unlock the maillogs so that they're picked up
	// by the worker
	for i := 0; i < 10; i++ {
		campaign, err := setupCampaign(i)
		if err != nil {
			t.Fatalf("error creating campaign: %v", err)
		}
		ms, err := models.GetMailLogsByCampaign(campaign.Id)
		if err != nil {
			t.Fatalf("error getting maillogs for campaign: %v", err)
		}
		for _, m := range ms {
			m.Unlock()
		}
	}

	const batchSize = 3
	lm := &logMailer{queue: make(chan []mailer.Mail, 40)}
	worker := &DefaultWorker{mailer: lm, batchSize: batchSize}

	// Trigger the worker, generating the maillogs and sending them to the
	// mailer
	worker.processCampaigns(time.Now())

	// Verify that each batch contains one campaign, respects the configured
	// size, and that no campaign loses a recipient.
	mailByCampaign := make(map[int64]int)
	for i := 0; i < 40; i++ {
		var ms []mailer.Mail
		select {
		case ms = <-lm.queue:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for campaign mail batches")
		}
		if len(ms) == 0 || len(ms) > batchSize {
			t.Fatalf("unexpected batch size: got %d, want 1-%d", len(ms), batchSize)
		}
		maillog, ok := ms[0].(*models.MailLog)
		if !ok {
			t.Fatalf("unable to cast mail to models.MailLog")
		}
		expected := maillog.CampaignId
		mailByCampaign[expected] += len(ms)
		for _, m := range ms {
			maillog, ok = m.(*models.MailLog)
			if !ok {
				t.Fatalf("unable to cast mail to models.MailLog")
			}
			got := maillog.CampaignId
			if got != expected {
				t.Fatalf("unexpected campaign ID received for maillog: got %d expected %d", got, expected)
			}
		}
	}
	if len(mailByCampaign) != 10 {
		t.Fatalf("unexpected campaign count: got %d expected 10", len(mailByCampaign))
	}
	for campaignID, count := range mailByCampaign {
		if count != 10 {
			t.Fatalf("unexpected mail count for campaign %d: got %d expected 10", campaignID, count)
		}
	}
}

func TestSplitMailBatches(t *testing.T) {
	tests := []struct {
		name       string
		mailCount  int
		batchSize  int
		batchSizes []int
	}{
		{name: "empty", mailCount: 0, batchSize: 3, batchSizes: nil},
		{name: "less than batch", mailCount: 2, batchSize: 3, batchSizes: []int{2}},
		{name: "exact batch", mailCount: 3, batchSize: 3, batchSizes: []int{3}},
		{name: "partial final batch", mailCount: 7, batchSize: 3, batchSizes: []int{3, 3, 1}},
		{name: "default batch", mailCount: 101, batchSize: 0, batchSizes: []int{100, 1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mails := make([]mailer.Mail, test.mailCount)
			for i := range mails {
				mails[i] = &models.EmailRequest{}
			}
			batches := splitMailBatches(mails, test.batchSize)
			var gotSizes []int
			flattened := make([]mailer.Mail, 0, len(mails))
			for _, batch := range batches {
				if len(batch) == 0 {
					t.Fatal("splitMailBatches returned an empty batch")
				}
				gotSizes = append(gotSizes, len(batch))
				flattened = append(flattened, batch...)
			}
			if !reflect.DeepEqual(gotSizes, test.batchSizes) {
				t.Fatalf("unexpected batch sizes: got %v expected %v", gotSizes, test.batchSizes)
			}
			if !reflect.DeepEqual(flattened, mails) {
				t.Fatal("splitMailBatches changed mail order or contents")
			}
		})
	}
}

func TestLaunchCampaignBatchesMail(t *testing.T) {
	setupTest(t)
	campaign, err := setupCampaign(1)
	if err != nil {
		t.Fatalf("error creating campaign: %v", err)
	}

	lm := &logMailer{queue: make(chan []mailer.Mail, 4)}
	worker := &DefaultWorker{mailer: lm, batchSize: 3}
	worker.LaunchCampaign(*campaign)

	gotSizes := []int{}
	gotMail := 0
	for len(lm.queue) > 0 {
		batch := <-lm.queue
		gotSizes = append(gotSizes, len(batch))
		gotMail += len(batch)
		for _, message := range batch {
			maillog := message.(*models.MailLog)
			if maillog.CampaignId != campaign.Id {
				t.Fatalf("unexpected campaign ID: got %d expected %d", maillog.CampaignId, campaign.Id)
			}
		}
	}
	if !reflect.DeepEqual(gotSizes, []int{3, 3, 3, 1}) {
		t.Fatalf("unexpected batch sizes: got %v expected [3 3 3 1]", gotSizes)
	}
	if gotMail != 10 {
		t.Fatalf("unexpected queued mail count: got %d expected 10", gotMail)
	}
}

func TestWorkerOptions(t *testing.T) {
	configured, err := New(WithBatchSize(25), WithMaxConcurrent(3))
	if err != nil {
		t.Fatalf("unexpected configuration error: %v", err)
	}
	w := configured.(*DefaultWorker)
	if w.batchSize != 25 || w.maxConcurrent != 3 {
		t.Fatalf("worker options were not applied: %#v", w)
	}
	defaults, err := New(WithBatchSize(0), WithMaxConcurrent(0))
	if err != nil {
		t.Fatalf("unexpected default configuration error: %v", err)
	}
	defaultWorker := defaults.(*DefaultWorker)
	if defaultWorker.batchSize != DefaultMailBatchSize || defaultWorker.maxConcurrent != mailer.DefaultMaxConcurrent {
		t.Fatalf("unexpected worker defaults: %#v", defaultWorker)
	}

	for _, option := range []func(Worker) error{WithBatchSize(-1), WithMaxConcurrent(-1)} {
		if _, err := New(option); err == nil {
			t.Fatal("expected negative mailer configuration to fail")
		}
	}
}
