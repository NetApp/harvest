package timedmap

import (
	"sync"
	"time"
)

type callback func(value interface{})

// TimedMap contains a map with all key-value pairs,
// and a timer, which cleans the map in the set
// tick durations from expired keys.
type TimedMap struct {
	mtx         sync.RWMutex
	container   map[keyWrap]*element
	elementPool *sync.Pool

	cleanupTickTime time.Duration
	cleanerTicker   *time.Ticker
	cleanerStopChan chan bool
	cleanerRunning  bool
}

type keyWrap struct {
	sec int
	key interface{}
}

// element contains the actual value as interface type,
// the thime when the value expires and an array of
// callbacks, which will be executed when the element
// expires.
type element struct {
	value   interface{}
	expires time.Time
	cbs     []callback
}

// New creates and returns a new instance of TimedMap.
// The passed cleanupTickTime will be passed to the
// cleanup Timer, which iterates through the map and
// deletes expired key-value pairs.
//
// Optionally, you can also pass a custom <-chan time.Time
// which controls the cleanup cycle if you want to use
// a single syncronyzed timer or something like that.
func New(cleanupTickTime time.Duration, tickerChan ...<-chan time.Time) *TimedMap {
	tm := &TimedMap{
		container:       make(map[keyWrap]*element),
		cleanerStopChan: make(chan bool),
		elementPool: &sync.Pool{
			New: func() interface{} {
				return new(element)
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

// Section returns a sectioned subset of
// the timed map with the given section
// identifier i.
func (tm *TimedMap) Section(i int) Section {
	if i == 0 {
		return tm
	}
	return newSection(tm, i)
}

// Ident returns the current sections ident.
// In the case of the root object TimedMap,
// this is always 0.
func (tm *TimedMap) Ident() int {
	return 0
}

// Set appends a key-value pair to the map or sets the value of
// a key. expiresAfter sets the expire time after the key-value pair
// will automatically be removed from the map.
func (tm *TimedMap) Set(key, value interface{}, expiresAfter time.Duration, cb ...callback) {
	tm.set(key, 0, value, expiresAfter, cb...)
}

// GetValue returns an interface of the value of a key in the
// map. The returned value is nil if there is no value to the
// passed key or if the value was expired.
func (tm *TimedMap) GetValue(key interface{}) interface{} {
	v := tm.get(key, 0)
	if v == nil {
		return nil
	}
	return v.value
}

// GetExpires returns the expire time of a key-value pair.
// If the key-value pair does not exist in the map or
// was expired, this will return an error object.
func (tm *TimedMap) GetExpires(key interface{}) (time.Time, error) {
	v := tm.get(key, 0)
	if v == nil {
		return time.Time{}, ErrKeyNotFound
	}
	return v.expires, nil
}

// SetExpire is deprecated.
// Please use SetExpires instead.
func (tm *TimedMap) SetExpire(key interface{}, d time.Duration) error {
	return tm.SetExpires(key, d)
}

// SetExpires sets the expire time for a key-value
// pair to the passed duration. If there is no value
// to the key passed , this will return an error.
func (tm *TimedMap) SetExpires(key interface{}, d time.Duration) error {
	return tm.setExpires(key, 0, d)
}

// Contains returns true, if the key exists in the map.
// false will be returned, if there is no value to the
// key or if the key-value pair was expired.
func (tm *TimedMap) Contains(key interface{}) bool {
	return tm.get(key, 0) != nil
}

// Remove deletes a key-value pair in the map.
func (tm *TimedMap) Remove(key interface{}) {
	tm.remove(key, 0)
}

// Refresh extends the expire time for a key-value pair
// about the passed duration. If there is no value to
// the key passed, this will return an error object.
func (tm *TimedMap) Refresh(key interface{}, d time.Duration) error {
	return tm.refresh(key, 0, d)
}

// Flush deletes all key-value pairs of the map.
func (tm *TimedMap) Flush() {
	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	for k, v := range tm.container {
		tm.elementPool.Put(v)
		delete(tm.container, k)
	}
}

// Size returns the current number of key-value pairs
// existent in the map.
func (tm *TimedMap) Size() int {
	return len(tm.container)
}

// StartCleanerInternal starts the cleanup loop controlled
// by an internal ticker with the given interval.
//
// If the cleanup loop is already running, it will be
// stopped and restarted using the new specification.
func (tm *TimedMap) StartCleanerInternal(interval time.Duration) {
	if tm.cleanerRunning {
		tm.StopCleaner()
	}
	tm.cleanerTicker = time.NewTicker(interval)
	go tm.cleanupLoop(tm.cleanerTicker.C)
}

// StartCleanerExternal starts the cleanup loop controlled
// by the given initiator channel. This is useful if you
// want to have more control over the cleanup loop or if
// you want to sync up multiple timedmaps.
//
// If the cleanup loop is already running, it will be
// stopped and restarted using the new specification.
func (tm *TimedMap) StartCleanerExternal(initiator <-chan time.Time) {
	if tm.cleanerRunning {
		tm.StopCleaner()
	}
	go tm.cleanupLoop(initiator)
}

// StopCleaner stops the cleaner go routine and timer.
// This should always be called after exiting a scope
// where TimedMap is used that the data can be cleaned
// up correctly.
func (tm *TimedMap) StopCleaner() {
	if !tm.cleanerRunning {
		return
	}
	tm.cleanerStopChan <- true
	if tm.cleanerTicker != nil {
		tm.cleanerTicker.Stop()
	}
}

// Snapshot returns a new map which represents the
// current key-value state of the internal container.
func (tm *TimedMap) Snapshot() map[interface{}]interface{} {
	return tm.getSnapshot(0)
}

// cleanupLoop holds the loop executing the cleanup
// when initiated by tc.
func (tm *TimedMap) cleanupLoop(tc <-chan time.Time) {
	tm.cleanerRunning = true
	defer func() {
		tm.cleanerRunning = false
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

// expireElement removes the specified key-value element
// from the map and executes all defined callback functions
func (tm *TimedMap) expireElement(key interface{}, sec int, v *element) {
	for _, cb := range v.cbs {
		cb(v.value)
	}

	k := keyWrap{
		sec: sec,
		key: key,
	}

	tm.elementPool.Put(v)
	delete(tm.container, k)
}

// cleanUp iterates trhough the map and expires all key-value
// pairs which expire time after the current time
func (tm *TimedMap) cleanUp() {
	now := time.Now()

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	for k, v := range tm.container {
		if now.After(v.expires) {
			tm.expireElement(k.key, k.sec, v)
		}
	}
}

// set sets the value for a key and section with the
// given expiration parameters
func (tm *TimedMap) set(key interface{}, sec int, val interface{}, expiresAfter time.Duration, cb ...callback) {
	// re-use element when existent on this key
	if v := tm.getRaw(key, sec); v != nil {
		v.value = val
		v.expires = time.Now().Add(expiresAfter)
		v.cbs = cb
		return
	}

	k := keyWrap{
		sec: sec,
		key: key,
	}

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	v := tm.elementPool.Get().(*element)
	v.value = val
	v.expires = time.Now().Add(expiresAfter)
	v.cbs = cb
	tm.container[k] = v
}

// get returns an element object by key and section
// if the value has not already expired
func (tm *TimedMap) get(key interface{}, sec int) *element {
	v := tm.getRaw(key, sec)

	if v == nil {
		return nil
	}

	if time.Now().After(v.expires) {
		tm.mtx.Lock()
		defer tm.mtx.Unlock()
		tm.expireElement(key, sec, v)
		return nil
	}

	return v
}

// getRaw returns the raw element object by key,
// not depending on expiration time
func (tm *TimedMap) getRaw(key interface{}, sec int) *element {
	k := keyWrap{
		sec: sec,
		key: key,
	}

	tm.mtx.RLock()
	v, ok := tm.container[k]
	tm.mtx.RUnlock()

	if !ok {
		return nil
	}

	return v
}

// remove removes an element from the map by giveb
// key and section
func (tm *TimedMap) remove(key interface{}, sec int) {
	k := keyWrap{
		sec: sec,
		key: key,
	}

	tm.mtx.Lock()
	defer tm.mtx.Unlock()

	v, ok := tm.container[k]
	if !ok {
		return
	}

	tm.elementPool.Put(v)
	delete(tm.container, k)
}

// refresh extends the lifetime of the given key in the
// given section by the duration d.
func (tm *TimedMap) refresh(key interface{}, sec int, d time.Duration) error {
	v := tm.get(key, sec)
	if v == nil {
		return ErrKeyNotFound
	}
	v.expires = v.expires.Add(d)
	return nil
}

// setExpires sets the lifetime of the given key in the
// given section to the duration d.
func (tm *TimedMap) setExpires(key interface{}, sec int, d time.Duration) error {
	v := tm.get(key, sec)
	if v == nil {
		return ErrKeyNotFound
	}
	v.expires = time.Now().Add(d)
	return nil
}

func (tm *TimedMap) getSnapshot(sec int) (m map[interface{}]interface{}) {
	m = make(map[interface{}]interface{})

	tm.mtx.RLock()
	defer tm.mtx.RUnlock()

	for k, v := range tm.container {
		if k.sec == sec {
			m[k.key] = v.value
		}
	}

	return
}
