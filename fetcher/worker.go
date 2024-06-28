package fetcher

import (
	"sync"
	"time"

	"github.com/cloudfoundry/cf_exporter/filters"
	"github.com/cloudfoundry/cf_exporter/models"
	log "github.com/sirupsen/logrus"
)

type WorkHandler func(*SessionExt, *BBSClient, *models.CFObjects) error

type Work struct {
	name    string
	handler WorkHandler
}

type Worker struct {
	filter  *filters.Filter
	group   sync.WaitGroup
	list    chan Work
	errs    chan error
	threads int
}

func NewWorker(threads int, filter *filters.Filter) *Worker {
	return &Worker{
		filter:  filter,
		list:    make(chan Work, 1000),
		errs:    make(chan error, 1000),
		threads: threads,
	}
}

func (c *Worker) Push(name string, handler WorkHandler) {
	c.group.Add(1)
	c.list <- Work{
		name:    name,
		handler: handler,
	}
}

func (c *Worker) PushIf(name string, handler WorkHandler, any ...string) {
	if c.filter.Any(any...) {
		c.Push(name, handler)
	}
}

func (c *Worker) Reset() {
	c.list = make(chan Work, 1000)
	c.errs = make(chan error, 1000)
}

func (c *Worker) Do(session *SessionExt, bbs *BBSClient, result *models.CFObjects) error {
	for i := 0; i < c.threads; i++ {
		go c.run(i, session, bbs, result)
	}
	return c.Wait()
}

func (c *Worker) Wait() error {
	log.Debugf("waiting for work groups to complete")
	c.group.Wait()
	close(c.list)
	close(c.errs)
	for err := range c.errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Worker) run(id int, session *SessionExt, bbs *BBSClient, entry *models.CFObjects) {
	for {
		work, ok := <-c.list
		if !ok {
			break
		}
		log.Debugf("[%2d] %s", id, work.name)
		start := time.Now()
		err := work.handler(session, bbs, entry)
		duration := time.Since(start)
		if err != nil {
			log.Errorf("[%2d] %s error: %s", id, work.name, err)
			c.errs <- err
		}
		log.Debugf("[%2d] %s (done, %.0f sec)", id, work.name, duration.Seconds())
		c.group.Done()
	}
}
