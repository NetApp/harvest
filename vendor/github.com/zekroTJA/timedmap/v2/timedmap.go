package timedmap

import (
	"sync"
	"sync/atomic"
	"time"
)

// Callback is a function which can be called when a key-value-pair has expired.
type Callback[TVal any] func(value TVal)

// TimedMap is a key-value map with lifetimes attached to values.
// Expired values are removed on access or via a cleanup coroutine,
// which can be enabled via the StartCleanerInternal method.
type TimedMap[TKey comparable, TVal any] struct {
	mtx         sync.RWMutex
	container   map[TKey]*Element[TVal]
	elementPool *sync.Pool

	cleanupTickTime time.Duration
	cleanerTicker   *time.Ticker
	cleanerStopChan chan bool
	cleanerRunning  atomic.Bool
}

// Element contains the actual value as interface type,
// the time when the value expires and an array of
// callbacks, which will be executed when the Element
// expires.
type Element[TVal any] struct {
	value   TVal
	expires time.Time
	cbs     []Callback[TVal]
}

// New creates and returns a new instance of TimedMap.
// The passed cleanupTickTime will be passed to the
// cleanup ticker, which iterates through the map and
// deletes expired key-value pairs on each iteration.
//
// Optionally, you can also pass a custom <-chan time.Time,
// which controls the cleanup cycle if you want to use
// a single synchronized timer or if you want to have more
// granular control over the cleanup loop.
//
// When passing 0 as cleanupTickTime and no tickerChan,
// the cleanup loop will not be started. You can call
// StartCleanerInternal or StartCleanerExternal to
// manually start the cleanup loop. These both methods
// can also be used to re-define the specification of
// the cleanup loop when already running if you want to.
func New[TKey comparable, TVal any](cleanupTickTime time.Duration, tickerChan ...<-chan time.Time) *TimedMap[TKey, TVal] {
	return newTimedMap[TKey, TVal](make(map[TKey]*Element[TVal]), cleanupTickTime, tickerChan)
}

// FromMap creates a new TimedMap containing all entries from
// the passed map m. Each entry will get assigned the passed
// expiration duration.
func FromMap[TKey comparable, TVal any](
	m map[TKey]TVal,
	expiration time.Duration,
	cleanupTickTime time.Duration,
	tickerChan ...<-chan time.Time,
) (*TimedMap[TKey, TVal], error) {
	if m == nil {
		return nil, ErrValueNoMap
	}

	exp := time.Now().Add(expiration)
	container := make(map[TKey]*Element[TVal])

	for k, v := range m {
		el := &Element[TVal]{
			value:   v,
			expires: exp,
		}
		container[k] = el
	}

	return newTimedMap(container, cleanupTickTime, tickerChan), nil
}

// Set appends a key-value pair to the map or sets the value of
// a key. expiresAfter sets the expiry time after the key-value pair
// will automatically be removed from the map.
func (tm *TimedMap[TKey, TVal]) Set(key TKey, value TVal, expiresAfter time.Duration, cb ...Callback[TVal]) {
	tm.set(key, value, expiresAfter, cb...)
}

// GetValue returns an interface of the value of a key in the
// map. The returned value is nil if there is no value to the
// passed key or if the value was expired.
func (tm *TimedMap[TKey, TVal]) GetValue(key TKey) (val TVal, ok bool) {
	v := tm.get(key)
	if v == nil {
		return val, false
	}
	tm.mtx.RLock()
	defer tm.mtx.RUnlock()
	return v.value, true
}

// GetExpires returns the expiry time of a key-value pair.
// If the key-value pair does not exist in the map or
// was expired, this will return an error object.
func (tm *TimedMap[TKey, TVal]) GetExpires(key TKey) (time.Time, error) {
	v := tm.get(key)
	if v == nil {
		return time.Time{}, ErrKeyNotFound
	}
	return v.expires, nil
}

// SetExpires sets the expiry time for a key-value
// pair to the passed duration. If there is no value
// to the key passed , this will return an error.
func (tm *TimedMap[TKey, TVal]) SetExpires(key TKey, d time.Duration) error {
	return tm.setExpires(key, d)
}

// Contains returns true, if the key exists in the map.
// false will be returned, if there is no value to the
// key or if the key-value pair was expired.
func (tm *TimedMap[TKey, TVal]) Contains(key TKey) bool {
	return tm.get(key) != nil
}

// Remove deletes a key-value pair in the map.
func (tm *TimedMap[TKey, TVal]) Remove(key TKey) {
	tm.remove(key)
}

// Refresh extends the expiry time for a key-value pair
// about the passed duration. If there is no value to
// the key passed, this will return an error object.
func (tm *TimedMap[TKey, TVal]) Refresh(key TKey, d time.Duration) error {
	return tm.refresh(key, d)
}

// Flush deletes all key-value pairs of the map.
func (tm *TimedMap[TKey, TVal]) Flush() {
	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	for k, v := range tm.container {
		tm.elementPool.Put(v)
		delete(tm.container, k)
	}
}

// Size returns the current number of key-value pairs
// existent in the map.
func (tm *TimedMap[TKey, TVal]) Size() int {
	return len(tm.container)
}

// StartCleanerInternal starts the cleanup loop controlled
// by an internal ticker with the given interval.
//
// If the cleanup loop is already running, it will be
// stopped and restarted using the new specification.
func (tm *TimedMap[TKey, TVal]) StartCleanerInternal(interval time.Duration) {
	if tm.cleanerRunning.Load() {
		tm.StopCleaner()
	}
	tm.cleanerTicker = time.NewTicker(interval)
	go tm.cleanupLoop(tm.cleanerTicker.C)
}

// StartCleanerExternal starts the cleanup loop controlled
// by the given initiator channel. This is useful if you
// want to have more control over the cleanup loop or if
// you want to sync up multiple TimedMaps.
//
// If the cleanup loop is already running, it will be
// stopped and restarted using the new specification.
func (tm *TimedMap[TKey, TVal]) StartCleanerExternal(initiator <-chan time.Time) {
	if tm.cleanerRunning.Load() {
		tm.StopCleaner()
	}
	go tm.cleanupLoop(initiator)
}

// StopCleaner stops the cleaner go routine and timer.
// This should always be called after exiting a scope
// where TimedMap is used that the data can be cleaned
// up correctly.
func (tm *TimedMap[TKey, TVal]) StopCleaner() {
	if !tm.cleanerRunning.Load() {
		return
	}
	tm.cleanerStopChan <- true
	if tm.cleanerTicker != nil {
		tm.cleanerTicker.Stop()
	}
}

// Snapshot returns a new map which represents the
// current key-value state of the internal container.
func (tm *TimedMap[TKey, TVal]) Snapshot() map[TKey]TVal {
	return tm.getSnapshot()
}

// cleanupLoop holds the loop executing the cleanup
// when initiated by tc.
func (tm *TimedMap[TKey, TVal]) cleanupLoop(tc <-chan time.Time) {
	tm.cleanerRunning.Store(true)
	defer func() {
		tm.cleanerRunning.Store(false)
	}()

	for {
		select {
		case <-tc:
			tm.cleanUp()
		case <-tm.cleanerStopChan:
			return
		}
	}
}

// expireElement removes the specified key-value Element
// from the map and executes all defined Callback functions
func (tm *TimedMap[TKey, TVal]) expireElement(key TKey, v *Element[TVal]) {
	for _, cb := range v.cbs {
		cb(v.value)
	}

	tm.elementPool.Put(v)
	delete(tm.container, key)
}

// cleanUp iterates through the map and expires all key-value
// pairs which expire time after the current time
func (tm *TimedMap[TKey, TVal]) cleanUp() {
	now := time.Now()

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	for k, v := range tm.container {
		if now.After(v.expires) {
			tm.expireElement(k, v)
		}
	}
}

// set sets the value for a key and section with the
// given expiration parameters
func (tm *TimedMap[TKey, TVal]) set(key TKey, val TVal, expiresAfter time.Duration, cb ...Callback[TVal]) {
	// re-use Element when existent on this key
	if v := tm.getRaw(key); v != nil {
		tm.mtx.Lock()
		defer tm.mtx.Unlock()
		v.value = val
		v.expires = time.Now().Add(expiresAfter)
		v.cbs = cb
		return
	}

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	v := tm.elementPool.Get().(*Element[TVal])
	v.value = val
	v.expires = time.Now().Add(expiresAfter)
	v.cbs = cb
	tm.container[key] = v
}

// get returns an Element object by key and section
// if the value has not already expired
func (tm *TimedMap[TKey, TVal]) get(key TKey) *Element[TVal] {
	v := tm.getRaw(key)

	if v == nil {
		return nil
	}

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	if time.Now().After(v.expires) {
		tm.expireElement(key, v)
		return nil
	}

	return v
}

// getRaw returns the raw Element object by key,
// not depending on expiration time
func (tm *TimedMap[TKey, TVal]) getRaw(key TKey) *Element[TVal] {
	tm.mtx.RLock()
	v, ok := tm.container[key]
	tm.mtx.RUnlock()

	if !ok {
		return nil
	}

	return v
}

// remove removes an Element from the map by give back the key
func (tm *TimedMap[TKey, TVal]) remove(key TKey) {
	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	v, ok := tm.container[key]
	if !ok {
		return
	}

	tm.elementPool.Put(v)
	delete(tm.container, key)
}

// refresh extends the lifetime of the given key in the
// given section by the duration d.
func (tm *TimedMap[TKey, TVal]) refresh(key TKey, d time.Duration) error {
	v := tm.get(key)
	if v == nil {
		return ErrKeyNotFound
	}
	tm.mtx.Lock()
	v.expires = v.expires.Add(d)
	tm.mtx.Unlock()
	return nil
}

// setExpires sets the lifetime of the given key in the
// given section to the duration d.
func (tm *TimedMap[TKey, TVal]) setExpires(key TKey, d time.Duration) error {
	v := tm.get(key)
	if v == nil {
		return ErrKeyNotFound
	}
	tm.mtx.Lock()
	v.expires = time.Now().Add(d)
	tm.mtx.Unlock()
	return nil
}

func (tm *TimedMap[TKey, TVal]) getSnapshot() (m map[TKey]TVal) {
	m = make(map[TKey]TVal)

	tm.mtx.RLock()
	defer tm.mtx.RUnlock()

	for k, v := range tm.container {
		m[k] = v.value
	}

	return
}

func newTimedMap[TKey comparable, TVal any](
	container map[TKey]*Element[TVal],
	cleanupTickTime time.Duration,
	tickerChan []<-chan time.Time,
) *TimedMap[TKey, TVal] {
	tm := &TimedMap[TKey, TVal]{
		container:       container,
		cleanerStopChan: make(chan bool),
		elementPool: &sync.Pool{
			New: func() any {
				return new(Element[TVal])
			},
		},
	}

	if len(tickerChan) > 0 {
		tm.StartCleanerExternal(tickerChan[0])
	} else if cleanupTickTime > 0 {
		tm.StartCleanerInternal(cleanupTickTime)
	}

	return tm
}
