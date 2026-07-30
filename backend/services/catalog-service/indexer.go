package main

import (
	"context"

	"github.com/zippyra/backend/shared/logger"
)

type IndexerQueue struct {
	searchEngine SearchEngine
	jobs         chan *Product
}

func NewIndexerQueue(searchEngine SearchEngine, bufferSize int) *IndexerQueue {
	if bufferSize <= 0 {
		bufferSize = 1000
	}
	iq := &IndexerQueue{
		searchEngine: searchEngine,
		jobs:         make(chan *Product, bufferSize),
	}
	go iq.startWorker()
	return iq
}

func (iq *IndexerQueue) Enqueue(p *Product) {
	select {
	case iq.jobs <- p:
	default:
		logger.Warn("Indexer queue buffer full, dropping ES async indexing for product %s", p.ID)
	}
}

func (iq *IndexerQueue) startWorker() {
	for p := range iq.jobs {
		if p == nil {
			continue
		}
		if err := iq.searchEngine.IndexProduct(context.Background(), p); err != nil {
			logger.Error("Async ES indexing failed for product %s (%s): %v", p.ID, p.Barcode, err)
		}
	}
}
