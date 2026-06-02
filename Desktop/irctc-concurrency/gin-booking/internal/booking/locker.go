package booking

import "sync"

type SeatLocker struct {
	mu    sync.Mutex
	locks map[string]struct{}
}

func NewSeatLocker() *SeatLocker {
	return &SeatLocker{locks: make(map[string]struct{})}
}

func (sl *SeatLocker) TryLock(key string) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	if _, held := sl.locks[key]; held {
		return false
	}
	sl.locks[key] = struct{}{}
	return true
}

func (sl *SeatLocker) Unlock(key string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	delete(sl.locks, key)
}

func (sl *SeatLocker) IsLocked(key string) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	_, held := sl.locks[key]
	return held
}
