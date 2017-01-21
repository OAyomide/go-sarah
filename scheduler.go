package sarah

import (
	"fmt"
	"github.com/oklahomer/cron"
	"github.com/oklahomer/go-sarah/log"
	"golang.org/x/net/context"
	"time"
)

type scheduler interface {
	remove(BotType, string) error
	update(BotType, *scheduledTask, func()) error
}

type taskScheduler struct {
	cron         *cron.Cron
	removingTask chan *removingTask
	updatingTask chan *updatingTask
}

func (s *taskScheduler) remove(botType BotType, taskID string) error {
	remove := &removingTask{
		botType: botType,
		taskID:  taskID,
		err:     make(chan error, 1),
	}
	s.removingTask <- remove

	return <-remove.err
}

func (s *taskScheduler) update(botType BotType, task *scheduledTask, fn func()) error {
	add := &updatingTask{
		botType: botType,
		task:    task,
		fn:      fn,
		err:     make(chan error, 1),
	}
	s.updatingTask <- add

	return <-add.err
}

type removingTask struct {
	botType BotType
	taskID  string
	err     chan error
}

type updatingTask struct {
	botType BotType
	task    *scheduledTask
	fn      func()
	err     chan error
}

func runScheduler(ctx context.Context, location *time.Location) scheduler {
	c := cron.NewWithLocation(location)
	// TODO set logger
	//c.ErrorLog = log.New(...)

	c.Start()

	s := &taskScheduler{
		cron:         c,
		removingTask: make(chan *removingTask, 1),
		updatingTask: make(chan *updatingTask, 1),
	}

	go s.receiveEvent(ctx)

	return s
}

func (s *taskScheduler) receiveEvent(ctx context.Context) {
	schedule := make(map[BotType]map[string]cron.EntryID)
	removeFunc := func(botType BotType, taskID string) error {
		botSchedule, ok := schedule[botType]
		if !ok {
			return fmt.Errorf("registered task not found: %s. %s.", botType.String(), taskID)
		}

		storedID, ok := botSchedule[taskID]
		if !ok {
			return fmt.Errorf("registered task not found: %s. %s.", botType.String(), taskID)
		}

		delete(botSchedule, taskID)
		s.cron.Remove(storedID)

		return nil
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("stop cron jobs due to context cancel")
			s.cron.Stop()
			return

		case remove := <-s.removingTask:
			remove.err <- removeFunc(remove.botType, remove.taskID)

		case add := <-s.updatingTask:
			if add.task.config.Schedule() == "" {
				add.err <- fmt.Errorf("empty schedule is given: %s.", add.task.Identifier())
			}

			removeFunc(add.botType, add.task.Identifier())

			id, err := s.cron.AddFunc(add.task.config.Schedule(), add.fn)
			if err != nil {
				add.err <- err
				break
			}

			if _, ok := schedule[add.botType]; !ok {
				schedule[add.botType] = make(map[string]cron.EntryID)
			}
			schedule[add.botType][add.task.Identifier()] = id
			add.err <- nil
		}
	}
}
