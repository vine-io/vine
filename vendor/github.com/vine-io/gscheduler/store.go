package gscheduler

import (
	"fmt"

	"github.com/vine-io/gscheduler/rbtree"
)

type Store interface {
	GetJobs() ([]*Job, error)
	GetByName(string) (*Job, error)
	GetById(string) (*Job, error)
	Count() uint
	Put(*Job) error
	Del(*Job) error
	Min() (*Job, error)
}

type jobStore struct {
	store *rbtree.Rbtree
}

func newJobStore() *jobStore {
	return &jobStore{
		store: rbtree.New(),
	}
}

func (s *jobStore) GetJobs() ([]*Job, error) {
	jobs := make([]*Job, 0)
	s.store.Ascend(s.store.Min(), func(item rbtree.Item) bool {
		i, ok := item.(*Job)
		if ok {
			jobs = append(jobs, i)
		}

		return true
	})
	return jobs, nil
}

func (s *jobStore) GetByName(name string) (*Job, error) {
	var job *Job
	s.store.Ascend(s.store.Min(), func(item rbtree.Item) bool {
		i, ok := item.(*Job)
		if !ok {
			return false
		}
		if i.name == name {
			job = i
			return true
		}
		return true
	})
	if job == nil {
		return nil, fmt.Errorf("no job")
	}

	return job, nil
}

func (s *jobStore) GetById(id string) (*Job, error) {
	var job *Job
	s.store.Ascend(s.store.Min(), func(item rbtree.Item) bool {
		i, ok := item.(*Job)
		if !ok {
			return false
		}
		if i.id == id {
			job = i
			return true
		}
		return true
	})
	if job == nil {
		return nil, fmt.Errorf("no job")
	}
	return job, nil
}

func (s *jobStore) Count() uint {
	return s.store.Len()
}

func (s *jobStore) Put(job *Job) error {
	s.store.Insert(job)
	return nil
}

func (s *jobStore) Del(job *Job) error {
	s.store.Delete(job)
	return nil
}

func (s *jobStore) Min() (*Job, error) {
	item := s.store.Min()
	if item == nil {
		return nil, fmt.Errorf("store empty")
	}
	job, ok := item.(*Job)
	if !ok {
		return nil, fmt.Errorf("invalid resource")
	}

	return job, nil
}
