package bpool

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	maxPoolSize = 27
	// maxBufferSize value to limit maximum memory for the buffer.
	maxBufferSize = (int64(1) << 34) - 1

	// DefaultInitialInterval duration for waiting in the queue due to system memory surge operations
	DefaultInitialInterval = 500 * time.Millisecond
	// DefaultRandomizationFactor sets factor to backoff when buffer pool reaches target size
	DefaultRandomizationFactor = 0.5
	// DefaultMaxElapsedTime sets maximum elapsed time to wait during backoff
	DefaultMaxElapsedTime   = 15 * time.Second
	DefaultBackoffThreshold = 0.7
)

var timerPool sync.Pool

// Buffer managed buffer by BufferPool for optimum memory usage
type Buffer struct {
	cap          *Capacity
	internal     buffer
	sync.RWMutex // Read Write mutex, guards access to internal buffer.
}

// NewBuffer returns buffer and initializes it using buf as its initial content.
func (pool *BufferPool) NewBuffer(buf []byte) *Buffer {
	return &Buffer{cap: pool.cap, internal: buffer{buf: buf, size: int64(len(buf))}}
}

// Get returns buffer if any in the pool or creates a new buffer
func (pool *BufferPool) Get() (buf *Buffer) {
	t := pool.cap.NewTicker()
	select {
	case buf = <-pool.buf:
	case <-t.C:
		timerPool.Put(t)
		buf = &Buffer{cap: pool.cap}
	}
	return
}

// Put resets the buffer and put it to the pool
func (pool *BufferPool) Put(buf *Buffer) {
	buf.Reset()
	if buf.Size() > pool.maxSize {
		return
	}
	if pool.Capacity() < 1 {
		pool.cap.Reset()
	}
	select {
	case pool.buf <- buf:
	default:
	}
}

// Extend allocates size to the buffer.
// Must extend buffer before calling buffer.Internal() method
func (buf *Buffer) Extend(size int64) (int64, error) {
	buf.Lock()
	defer buf.Unlock()
	off, err := buf.internal.allocate(uint32(size))
	if err != nil {
		return 0, err
	}
	buf.cap.Lock()
	defer buf.cap.Unlock()
	buf.cap.size += size
	return off, nil
}

// Internal returns underline internal buffer.
// This method is useful when you need an underline byte slice of a buffer
// for example while reading chunk from os.File i.e file.ReatAt() method.
// It is not safe to call buffer.Write() method once you get underline byte slice.
// You need to perform an external locking if calling buffer.Internal() and buffer.Write() methods.
func (buf *Buffer) Internal() []byte {
	return buf.internal.buf
}

// Write writes to the buffer
func (buf *Buffer) Write(p []byte) (int, error) {
	buf.Lock()
	defer buf.Unlock()
	if buf.cap.WriteBackOff {
		t := buf.cap.NewTicker()
		select {
		case <-t.C:
			timerPool.Put(t)
		}
	}
	off, err := buf.internal.allocate(uint32(len(p)))
	if err != nil {
		return 0, err
	}
	if _, err := buf.internal.writeAt(p, off); err != nil {
		return 0, err
	}
	buf.cap.Lock()
	defer buf.cap.Unlock()
	buf.cap.size += int64(len(p))
	return len(p), nil
}

// WriteAt writes to the buffer at the given offset
func (buf *Buffer) WriteAt(p []byte, off int64) (int, error) {
	buf.Lock()
	defer buf.Unlock()
	if buf.cap.WriteBackOff {
		t := buf.cap.NewTicker()
		select {
		case <-t.C:
			timerPool.Put(t)
		}
	}
	if _, err := buf.internal.writeAt(p, off); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Bytes gets data from internal buffer
func (buf *Buffer) Bytes() []byte {
	buf.RLock()
	defer buf.RUnlock()
	data, _ := buf.internal.bytes()
	return data
}

// Slice provide the data for start and end offset
func (buf *Buffer) Slice(start int64, end int64) ([]byte, error) {
	p := make([]byte, end-start)
	_, err := buf.internal.readAt(p, start)
	return p, err
}

// ReadAt read byte of size at the given offset from internal buffer
func (buf *Buffer) ReadAt(p []byte, off int64) (int, error) {
	buf.RLock()
	defer buf.RUnlock()
	return buf.internal.readAt(p, off)
}

// Read reads specific number of bytes from internal buffer
func (buf *Buffer) Read(p []byte) (int, error) {
	buf.RLock()
	defer buf.RUnlock()
	return buf.internal.read(p)
}

// Reset resets the buffer
func (buf *Buffer) Reset() (ok bool) {
	buf.Lock()
	defer buf.Unlock()

	buf.cap.Lock()
	defer buf.cap.Unlock()
	buf.cap.size -= buf.internal.size
	return buf.internal.reset()
}

// Size internal buffer size
func (buf *Buffer) Size() int64 {
	buf.RLock()
	defer buf.RUnlock()
	return buf.internal.size
}

type (
	// Capacity manages the BufferPool capacity to limit excess memory usage.
	Capacity struct {
		sync.RWMutex

		size       int64
		targetSize int64

		InitialInterval     time.Duration
		RandomizationFactor float64
		currentInterval     time.Duration
		MaxElapsedTime      time.Duration

		WriteBackOff bool
	}
	// BufferPool represents the thread safe buffer pool.
	// All BufferPool methods are safe for concurrent use by multiple goroutines.
	BufferPool struct {
		buf chan *Buffer

		// Capacity
		maxSize int64
		cap     *Capacity

		// close
		closeC chan struct{}
	}
)

// NewBufferPool creates a new buffer pool.
func NewBufferPool(size int64, opts *Options) *BufferPool {
	opts = opts.copyWithDefaults()
	if size > maxBufferSize {
		size = maxBufferSize
	}

	cap := &Capacity{
		targetSize:          size,
		InitialInterval:     opts.InitialInterval,
		RandomizationFactor: opts.RandomizationFactor,
		MaxElapsedTime:      opts.MaxElapsedTime,

		WriteBackOff: opts.WriteBackOff,
	}
	cap.Reset()

	pool := &BufferPool{
		buf: make(chan *Buffer, opts.MaxPoolSize),

		// Capacity
		maxSize: int64(size / int64(opts.MaxPoolSize)),
		cap:     cap,
		// close
		closeC: make(chan struct{}, 1),
	}

	go pool.drain()

	return pool
}

// Capacity return the buffer pool capacity in proportion to target size.
func (pool *BufferPool) Capacity() float64 {
	pool.cap.RLock()
	defer pool.cap.RUnlock()
	return float64(pool.cap.size) / float64(pool.cap.targetSize)
}

// Reset the interval back to the initial interval.
// Reset must be called before using pool.
func (cap *Capacity) Reset() {
	cap.Lock()
	defer cap.Unlock()
	cap.currentInterval = cap.InitialInterval
}

// NextBackOff calculates the next backoff interval using the formula:
// 	Randomized interval = RetryInterval * (1 ± RandomizationFactor)
func (cap *Capacity) NextBackOff(multiplier float64) time.Duration {
	defer cap.incrementCurrentInterval(multiplier)
	return getRandomValueFromInterval(cap.RandomizationFactor, rand.Float64(), cap.currentInterval)
}

// Increments the current interval by multiplying it with the multiplier.
func (cap *Capacity) incrementCurrentInterval(multiplier float64) {
	cap.Lock()
	defer cap.Unlock()
	cap.currentInterval = time.Duration(float64(cap.currentInterval) * multiplier)
	if cap.currentInterval > cap.MaxElapsedTime {
		cap.currentInterval = cap.MaxElapsedTime
	}
}

// Decrements the current interval by multiplying it with factor.
func (cap *Capacity) decrementCurrentInterval(factor float64) {
	cap.currentInterval = time.Duration(float64(cap.currentInterval) * factor)
}

// Returns a random value from the following interval:
// [currentInterval - randomizationFactor * currentInterval, currentInterval + randomizationFactor * currentInterval].
func getRandomValueFromInterval(randomizationFactor, random float64, currentInterval time.Duration) time.Duration {
	var delta = randomizationFactor * float64(currentInterval)
	var minInterval = float64(currentInterval) - delta
	var maxInterval = float64(currentInterval) + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	return time.Duration(minInterval + (random * (maxInterval - minInterval + 1)))
}

// NewTicker creates or get ticker from timer pool. It uses backoff duration of the pool for the timer.
func (cap *Capacity) NewTicker() *time.Timer {
	cap.RLock()
	factor := float64(cap.size) / float64(cap.targetSize)
	cap.RUnlock()
	d := time.Duration(time.Duration(factor) * time.Millisecond)
	if d > 1 {
		d = cap.NextBackOff(factor)
	}

	if v := timerPool.Get(); v != nil {
		t := v.(*time.Timer)
		if t.Reset(d) {
			panic(fmt.Sprintf("pool.NewTicker: active timer trapped to the pool"))
		}
		return t
	}
	return time.NewTimer(d)
}

// Backoff backs off buffer pool if currentInterval is greater than Backoff threshold.
func (pool *BufferPool) Backoff() {
	t := pool.cap.NewTicker()
	select {
	case <-t.C:
		timerPool.Put(t)
	}
}

// Done closes the buffer pool and stops the drain goroutine.
func (pool *BufferPool) Done() {
	close(pool.closeC)
}

func (pool *BufferPool) drain() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			select {
			case <-pool.closeC:
				return
			case <-pool.buf:
			default:
			}
		}
	}
}
