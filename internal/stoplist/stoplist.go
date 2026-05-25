package stoplist

import (
	"log/slog"
	"sync"
)

type Stoplist struct {
	mu     *sync.RWMutex
	words  map[string]struct{}
	logger *slog.Logger
}

func NewStoplist(stoplist []string, logger *slog.Logger) *Stoplist {
	mu := &sync.RWMutex{}
	words := make(map[string]struct{})
	for _, word := range stoplist {
		words[word] = struct{}{}
	}

	return &Stoplist{
		mu:     mu,
		words:  words,
		logger: logger,
	}
}

func (sl *Stoplist) Add(word string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	sl.words[word] = struct{}{}
	if sl.logger != nil {
		sl.logger.Info("Word was successfully added to the stoplist", "word", word)
	}
}

func (sl *Stoplist) Remove(word string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	delete(sl.words, word)
	if sl.logger != nil {
		sl.logger.Info("Word was successfully removed from the stoplist", "word", word)
	}
}

func (sl *Stoplist) Contains(word string) bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	_, ok := sl.words[word]
	return ok
}

func (sl *Stoplist) GetAll() []string {
	words := make([]string, 0, len(sl.words))
	for word := range sl.words {
		words = append(words, word)
	}
	return words
}
