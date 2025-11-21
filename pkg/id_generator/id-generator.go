package id_generator

import (
	"fmt"
	"github.com/google/wire"
	"sync"
	"time"
)

var ProviderSet = wire.NewSet(NewTaskIDGenerator)

type TaskIDGenerator interface {
	Generate(prefix string) string
}

func NewTaskIDGenerator() TaskIDGenerator {
	return &singleNodeGenerator{}
}

type singleNodeGenerator struct {
	mu sync.Mutex
}

func (t *singleNodeGenerator) Generate(prefix string) string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return prefix + fmt.Sprintf("%v", time.Now().UnixMilli())
}
