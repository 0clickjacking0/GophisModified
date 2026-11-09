package mailer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/textproto"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

type concurrencyTracker struct {
	active     int32
	maxActive  int32
	dialCount  int32
	closeCount int32
	started    chan struct{}
	completed  chan struct{}
	release    <-chan struct{}
}

func (t *concurrencyTracker) enter() {
	active := atomic.AddInt32(&t.active, 1)
	atomic.AddInt32(&t.dialCount, 1)
	for {
		maxActive := atomic.LoadInt32(&t.maxActive)
		if active <= maxActive || atomic.CompareAndSwapInt32(&t.maxActive, maxActive, active) {
			break
		}
	}
	t.started <- struct{}{}
}

type trackedDialer struct {
	tracker *concurrencyTracker
}

func (d *trackedDialer) Dial() (Sender, error) {
	d.tracker.enter()
	<-d.tracker.release
	return &trackedSender{tracker: d.tracker}, nil
}

type trackedSender struct {
	tracker *concurrencyTracker
}

func (s *trackedSender) Send(string, []string, io.WriterTo) error { return nil }
func (s *trackedSender) Reset() error                             { return nil }
func (s *trackedSender) Close() error {
	atomic.AddInt32(&s.tracker.active, -1)
	atomic.AddInt32(&s.tracker.closeCount, 1)
	s.tracker.completed <- struct{}{}
	return nil
}

func waitForSignal(t *testing.T, ch <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func trackedMessage(tracker *concurrencyTracker) *mockMessage {
	message := newMockMessage("from@example.com", []string{"to@example.com"}, bytes.NewBufferString("body"))
	message.setDialer(func() (Dialer, error) {
		return &trackedDialer{tracker: tracker}, nil
	})
	return message
}

func generateMessages(dialer Dialer) []Mail {
	to := []string{"to@example.com"}

	messageContents := []io.WriterTo{
		bytes.NewBuffer([]byte("First email")),
		bytes.NewBuffer([]byte("Second email")),
	}

	m1 := newMockMessage("first@example.com", to, messageContents[0])
	m2 := newMockMessage("second@example.com", to, messageContents[1])

	m1.setDialer(func() (Dialer, error) { return dialer, nil })

	messages := []Mail{m1, m2}
	return messages
}

func newMockErrorSender(err error) *mockSender {
	sender := newMockSender()
	// The sending function will send a temporary error to emulate
	// a backoff.
	sender.setSend(func(mm *mockMessage) error {
		if len(sender.messages) == 1 {
			return err
		}
		sender.messageChan <- mm
		return nil
	})
	return sender
}

func TestDialHost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	md := newMockDialer()
	md.setDial(md.unreachableDial)
	_, err := dialHost(ctx, md)
	if _, ok := err.(*ErrMaxConnectAttempts); !ok {
		t.Fatalf("Didn't receive expected ErrMaxConnectAttempts. Got: %s", err)
	}
	e := err.(*ErrMaxConnectAttempts)
	if e.underlyingError != errHostUnreachable {
		t.Fatalf("Got invalid underlying error. Expected %s Got %s\n", e.underlyingError, errHostUnreachable)
	}
	if md.dialCount != MaxReconnectAttempts {
		t.Fatalf("Unexpected number of reconnect attempts. Expected %d, Got %d", MaxReconnectAttempts, md.dialCount)
	}
	md.setDial(md.defaultDial)
	_, err = dialHost(ctx, md)
	if err != nil {
		t.Fatalf("Unexpected error when dialing the mock host: %s", err)
	}
}

func TestDialHostCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dialer := newMockDialer()
	sender, err := dialHost(ctx, dialer)
	if sender != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected cancelled dial result: sender=%v err=%v", sender, err)
	}
	if dialer.dialCount != 0 {
		t.Fatalf("dialer called after cancellation: got %d calls", dialer.dialCount)
	}
}

func TestSendMailCancellationDuringReconnectPreservesMail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	sender := newMockSender()
	sender.setSend(func(*mockMessage) error {
		cancel()
		return errors.New("connection closed")
	})
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})
	messages := generateMessages(dialer)

	sendMail(ctx, dialer, messages)

	for i, message := range messages {
		mock := message.(*mockMessage)
		if mock.finished || mock.backoffCount != 0 || mock.err != nil {
			t.Fatalf("message %d changed state after cancellation: %#v", i, mock)
		}
	}
}

func TestMailWorkerConcurrencyLimit(t *testing.T) {
	const (
		maxConcurrent = 3
		batchCount    = 8
	)
	release := make(chan struct{})
	tracker := &concurrencyTracker{
		started:   make(chan struct{}, batchCount),
		completed: make(chan struct{}, batchCount),
		release:   release,
	}
	ctx, cancel := context.WithCancel(context.Background())
	mw := NewMailWorkerWithConcurrency(maxConcurrent)
	startDone := make(chan struct{})
	go func() {
		mw.Start(ctx)
		close(startDone)
	}()

	queueDone := make(chan struct{})
	messages := make([]*mockMessage, batchCount)
	go func() {
		for i := 0; i < batchCount; i++ {
			messages[i] = trackedMessage(tracker)
			mw.Queue([]Mail{messages[i]})
		}
		close(queueDone)
	}()

	for i := 0; i < maxConcurrent; i++ {
		waitForSignal(t, tracker.started, "concurrent batch to start")
	}
	select {
	case <-tracker.started:
		t.Fatalf("more than %d mail batches started concurrently", maxConcurrent)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	for i := 0; i < batchCount; i++ {
		waitForSignal(t, tracker.completed, "mail batch to complete")
	}
	waitForSignal(t, queueDone, "all mail batches to be queued")

	if got := atomic.LoadInt32(&tracker.maxActive); got != maxConcurrent {
		t.Fatalf("unexpected peak concurrency: got %d expected %d", got, maxConcurrent)
	}
	if got := atomic.LoadInt32(&tracker.dialCount); got != batchCount {
		t.Fatalf("unexpected dial count: got %d expected %d", got, batchCount)
	}
	if got := atomic.LoadInt32(&tracker.closeCount); got != batchCount {
		t.Fatalf("unexpected close count: got %d expected %d", got, batchCount)
	}
	for i, message := range messages {
		if message.successCount != 1 || message.errorCount != 0 || message.backoffCount != 0 {
			t.Fatalf("mail batch %d had unexpected status callbacks: success=%d error=%d backoff=%d", i, message.successCount, message.errorCount, message.backoffCount)
		}
	}

	cancel()
	waitForSignal(t, startDone, "mail worker to stop")
}

func TestMailWorkerCancellationWhileAtLimit(t *testing.T) {
	release := make(chan struct{})
	tracker := &concurrencyTracker{
		started:   make(chan struct{}, 1),
		completed: make(chan struct{}, 1),
		release:   release,
	}
	ctx, cancel := context.WithCancel(context.Background())
	mw := NewMailWorkerWithConcurrency(1)
	startDone := make(chan struct{})
	go func() {
		mw.Start(ctx)
		close(startDone)
	}()

	mw.Queue([]Mail{trackedMessage(tracker)})
	waitForSignal(t, tracker.started, "first mail batch to start")

	cancel()
	waitForSignal(t, startDone, "mail worker to stop while at its limit")
	close(release)
	waitForSignal(t, tracker.completed, "active mail batch to complete")

	select {
	case <-tracker.started:
		t.Fatal("mail batch started after worker cancellation")
	default:
	}
}

func TestMailWorkerSkipsEmptyBatch(t *testing.T) {
	release := make(chan struct{})
	close(release)
	tracker := &concurrencyTracker{
		started:   make(chan struct{}, 1),
		completed: make(chan struct{}, 1),
		release:   release,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mw := NewMailWorkerWithConcurrency(1)
	go mw.Start(ctx)

	mw.Queue(nil)
	mw.Queue([]Mail{trackedMessage(tracker)})
	waitForSignal(t, tracker.completed, "non-empty mail batch to complete")
	if got := atomic.LoadInt32(&tracker.dialCount); got != 1 {
		t.Fatalf("unexpected dial count after empty batch: got %d expected 1", got)
	}
}

func TestMailWorkerGetDialerErrorReleasesSlot(t *testing.T) {
	expectedError := errors.New("invalid sending profile")
	failed := newMockMessage("from@example.com", []string{"to@example.com"}, bytes.NewBufferString("body"))
	failed.setDialer(func() (Dialer, error) {
		return nil, expectedError
	})
	release := make(chan struct{})
	close(release)
	tracker := &concurrencyTracker{
		started:   make(chan struct{}, 1),
		completed: make(chan struct{}, 1),
		release:   release,
	}
	ctx, cancel := context.WithCancel(context.Background())
	mw := NewMailWorkerWithConcurrency(1)
	startDone := make(chan struct{})
	go func() {
		mw.Start(ctx)
		close(startDone)
	}()

	queueDone := make(chan struct{})
	go func() {
		mw.Queue([]Mail{failed})
		mw.Queue([]Mail{trackedMessage(tracker)})
		close(queueDone)
	}()
	waitForSignal(t, tracker.completed, "batch after dialer error to complete")
	waitForSignal(t, queueDone, "batches around dialer error to be queued")
	if failed.errorCount != 1 || !errors.Is(failed.err, expectedError) {
		t.Fatalf("unexpected dialer error status: count=%d err=%v", failed.errorCount, failed.err)
	}

	cancel()
	waitForSignal(t, startDone, "mail worker to stop")
}

func TestMailWorkerStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	sender := newMockSender()
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	idx := 0
	for message := range sender.messageChan {
		got = append(got, message)
		original := messages[idx].(*mockMessage)
		if original.from != message.from {
			t.Fatalf("Invalid message received. Expected %s, Got %s", original.from, message.from)
		}
		idx++
	}
	if len(got) != len(messages) {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), len(messages))
	}
}

func TestBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	expectedError := &textproto.Error{
		Code: 400,
		Msg:  "Temporary error",
	}

	sender := newMockErrorSender(expectedError)
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	for message := range sender.messageChan {
		got = append(got, message)
	}
	// Check that we only sent one message
	expectedCount := 1
	if len(got) != expectedCount {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), expectedCount)
	}

	// Check that it's the correct message
	originalFrom := messages[1].(*mockMessage).from
	if got[0].from != originalFrom {
		t.Fatalf("Invalid message received. Expected %s, Got %s", originalFrom, got[0].from)
	}

	// Check that the first message performed a backoff
	backoffCount := messages[0].(*mockMessage).backoffCount
	if backoffCount != expectedCount {
		t.Fatalf("Did not receive expected backoff. Got backoffCount %d, Expected %d", backoffCount, expectedCount)
	}

	// Check that there was a reset performed on the sender
	if sender.resetCount != expectedCount {
		t.Fatalf("Did not receive expected reset. Got resetCount %d, expected %d", sender.resetCount, expectedCount)
	}
}

func TestPermError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	expectedError := &textproto.Error{
		Code: 500,
		Msg:  "Permanent error",
	}

	sender := newMockErrorSender(expectedError)
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		return sender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	for message := range sender.messageChan {
		got = append(got, message)
	}
	// Check that we only sent one message
	expectedCount := 1
	if len(got) != expectedCount {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), expectedCount)
	}

	// Check that it's the correct message
	originalFrom := messages[1].(*mockMessage).from
	if got[0].from != originalFrom {
		t.Fatalf("Invalid message received. Expected %s, Got %s", originalFrom, got[0].from)
	}

	message := messages[0].(*mockMessage)

	// Check that the first message did not perform a backoff
	expectedBackoffCount := 0
	backoffCount := message.backoffCount
	if backoffCount != expectedBackoffCount {
		t.Fatalf("Did not receive expected backoff. Got backoffCount %d, Expected %d", backoffCount, expectedCount)
	}

	// Check that there was a reset performed on the sender
	if sender.resetCount != expectedCount {
		t.Fatalf("Did not receive expected reset. Got resetCount %d, expected %d", sender.resetCount, expectedCount)
	}

	// Check that the email errored out appropriately
	if !reflect.DeepEqual(message.err, expectedError) {
		t.Fatalf("Did not received expected error. Got %#v\nExpected %#v", message.err, expectedError)
	}
}

func TestUnknownError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := NewMailWorker()
	go func(ctx context.Context) {
		mw.Start(ctx)
	}(ctx)

	expectedError := errors.New("Unexpected error")

	failedSender := newMockErrorSender(expectedError)
	reconnectedSender := newMockSender()
	dialer := newMockDialer()
	dialer.setDial(func() (Sender, error) {
		if dialer.dialCount == 1 {
			return failedSender, nil
		}
		return reconnectedSender, nil
	})

	messages := generateMessages(dialer)

	// Send the campaign
	mw.Queue(messages)

	got := []*mockMessage{}

	for message := range reconnectedSender.messageChan {
		got = append(got, message)
	}
	// Check that we only sent one message
	expectedCount := 1
	if len(got) != expectedCount {
		t.Fatalf("Unexpected number of messages received. Expected %d Got %d", len(got), expectedCount)
	}

	// Check that it's the correct message
	originalFrom := messages[1].(*mockMessage).from
	if got[0].from != originalFrom {
		t.Fatalf("Invalid message received. Expected %s, Got %s", originalFrom, got[0].from)
	}

	message := messages[0].(*mockMessage)

	// If we get an unexpected error, this means that it's likely the
	// underlying connection dropped. When this happens, we expect the
	// connection to be re-established (see #997).
	// In this case, we're successfully reestablishing the connection
	// so we expect the backoff to occur.
	expectedBackoffCount := 1
	backoffCount := message.backoffCount
	if backoffCount != expectedBackoffCount {
		t.Fatalf("Did not receive expected backoff. Got backoffCount %d, Expected %d", backoffCount, expectedBackoffCount)
	}

	// Check that the underlying connection was reestablished
	expectedDialCount := 2
	if dialer.dialCount != expectedDialCount {
		t.Fatalf("Did not receive expected dial count. Got %d expected %d", dialer.dialCount, expectedDialCount)
	}
	if failedSender.status != "closed" || reconnectedSender.status != "closed" {
		t.Fatalf("senders were not closed after reconnect: first=%s second=%s", failedSender.status, reconnectedSender.status)
	}

	// Check that the email errored out appropriately
	if !reflect.DeepEqual(message.err, expectedError) {
		t.Fatalf("Did not received expected error. Got %#v\nExpected %#v", message.err, expectedError)
	}
}
