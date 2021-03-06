package sarah

import (
	"testing"
	"time"
)

func Test_status_start(t *testing.T) {
	s := &status{}
	s.start()

	if s.finished == nil {
		t.Error("A channel to judge running status must be set.")
	}
}

func Test_status_stop(t *testing.T) {
	s := &status{
		finished: make(chan struct{}),
	}

	s.stop()

	select {
	case <-s.finished:
		// O.K. Channel is closed.

	case <-time.NewTimer(100 * time.Millisecond).C:
		t.Error("A channel is not closed on status.stop.")

	}

	s.stop() // Multiple call to this method should not panic.
}

func Test_status_addBot(t *testing.T) {
	botType := BotType("dummy")
	bot := &DummyBot{BotTypeValue: botType}
	s := &status{}
	s.addBot(bot)

	botStatuses := s.bots
	if len(botStatuses) != 1 {
		t.Fatal("Status for one and only one Bot should be set.")
	}

	bs := botStatuses[0]

	if bs.botType != botType {
		t.Errorf("Expected BotType is not set: %s.", bs.botType)
	}

	if !bs.running() {
		t.Error("Bot status must be running at this point.")
	}
}

func Test_status_stopBot(t *testing.T) {
	botType := BotType("dummy")
	bs := &botStatus{
		botType:  botType,
		finished: make(chan struct{}),
	}
	s := &status{
		bots: []*botStatus{bs},
	}

	bot := &DummyBot{BotTypeValue: botType}
	s.stopBot(bot)

	botStatuses := s.bots
	if len(botStatuses) != 1 {
		t.Fatal("Status for one and only one Bot should be set.")
	}

	stored := botStatuses[0]

	if stored.botType != botType {
		t.Errorf("Expected BotType is not set: %s.", bs.botType)
	}

	if stored.running() {
		t.Error("Bot status must not be running at this point.")
	}
}

func Test_status_snapshot(t *testing.T) {
	botType := BotType("dummy")
	bs := &botStatus{
		botType:  botType,
		finished: make(chan struct{}),
	}
	s := &status{
		bots:     []*botStatus{bs},
		finished: make(chan struct{}),
	}

	snapshot := s.snapshot()
	if !snapshot.Running {
		t.Error("Status.Running should be true at this point.")
	}

	if len(snapshot.Bots) != 1 {
		t.Errorf("The number of registered Bot should be one, but was %d.", len(snapshot.Bots))
	}

	if !snapshot.Bots[0].Running {
		t.Error("BotStatus.Running should be true at this point.")
	}

	close(bs.finished)
	close(s.finished)

	snapshot = s.snapshot()

	if snapshot.Running {
		t.Error("Status.Running should be false at this point.")
	}

	if snapshot.Bots[0].Running {
		t.Error("BotStatus.Running should be false at this point.")
	}
}

func Test_botStatus_running(t *testing.T) {
	bs := &botStatus{
		botType:  "dummy",
		finished: make(chan struct{}),
	}

	if !bs.running() {
		t.Error("botStatus.running() should be true at this point.")
	}

	close(bs.finished)

	if bs.running() {
		t.Error("botStatus.running() should be false at this point.")
	}
}

func Test_botStatus_stop(t *testing.T) {
	bs := &botStatus{
		finished: make(chan struct{}),
	}

	bs.stop()

	select {
	case <-bs.finished:
		// O.K. Channel is closed.

	case <-time.NewTimer(100 * time.Millisecond).C:
		t.Error("A channel is not closed on botStatus.stop.")

	}

	bs.stop() // Multiple call to this method should not panic.
}
