package stats

import (
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Buffer is the implementation of a measure handler which uses a Serializer to
// serialize the metric into a memory buffer and write them once the buffer has
// reached a target size.
type Buffer struct {
	// Target size of the memory buffer where metrics are serialized.
	//
	// If left to zero, a size of 1024 bytes is used as default (this is low,
	// you should set this value).
	//
	// Note that if the buffer size is small, the program may generate metrics
	// that don't fit into the configured buffer size. In that case the buffer
	// will still pass the serialized byte slice to its Serializer to leave the
	// decision of accepting or rejecting the metrics.
	BufferSize int

	// Size of the internal buffer pool, this controls how well the buffer
	// performs in highly concurrent environments. If unset, 2 x GOMAXPROCS
	// is used as a default value.
	BufferPoolSize int

	// The Serializer used to write the measures.
	//
	// This field cannot be nil.
	Serializer Serializer

	once    sync.Once
	offset  uint64
	buffers []buffer
}

// HandleMeasures satisfies the Handler interface.
func (b *Buffer) HandleMeasures(time time.Time, measures ...Measure) {
	if len(measures) == 0 {
		return
	}

	size := b.bufferSize()
	b.prepare(size)

	buffer := b.acquireBuffer()
	length := buffer.len()
	buffer.append(b.Serializer, time, measures...)

	if buffer.len() >= size {
		if length == 0 {
			// When there were no data in the buffer prior to writing the set of
			// measures we unfortunately have to overflow the configured buffer
			// size, but the Serializer documentation mentions that this corner
			// case has to be handled.
			length = buffer.len()
		}
		buffer.flush(b.Serializer, length)
	}

	buffer.release()
}

// Flush satisfies the Flusher interface.
func (b *Buffer) Flush() {
	b.prepare(b.bufferSize())

	for i := range b.buffers {
		if buffer := &b.buffers[i]; buffer.acquire() {
			buffer.flush(b.Serializer, buffer.len())
			buffer.release()
		}
	}
}

func (b *Buffer) prepare(bufferSize int) {
	b.once.Do(func() {
		b.buffers = make([]buffer, b.bufferPoolSize())
		for i := range b.buffers {
			b.buffers[i].init(bufferSize)
		}
	})
}

func (b *Buffer) bufferSize() int {
	if b.BufferSize != 0 {
		return b.BufferSize
	}
	return 1024
}

func (b *Buffer) bufferPoolSize() int {
	if b.BufferPoolSize != 0 {
		return b.BufferPoolSize
	}
	return 2 * runtime.GOMAXPROCS(0)
}

func (b *Buffer) acquireBuffer() *buffer {
	i := uint64(0)
	n := uint64(len(b.buffers))

	for {
		offset := atomic.AddUint64(&b.offset, 1) % n
		buffer := &b.buffers[offset]

		if buffer.acquire() {
			return buffer
		}

		if i++; (i % n) == 0 {
			runtime.Gosched()
		}
	}
}

// The Serializer interface is used to abstract the logic of serializing
// measures.
type Serializer interface {
	io.Writer

	// Appends the serialized representation of the given measures into b.
	//
	// The method must not retain any of the arguments.
	AppendMeasures(b []byte, time time.Time, measures ...Measure) []byte
}

type buffer struct {
	lock uint64
	data []byte
	pad  [32]byte // padding to avoid false sharing between threads
}

func (b *buffer) acquire() bool {
	return atomic.CompareAndSwapUint64(&b.lock, 0, 1)
}

func (b *buffer) release() {
	atomic.StoreUint64(&b.lock, 0)
}

func (b *buffer) init(size int) {
	b.data = make([]byte, 0, size+(size/4))
}

func (b *buffer) append(s Serializer, t time.Time, m ...Measure) {
	b.data = s.AppendMeasures(b.data, t, m...)
}

func (b *buffer) len() int {
	return len(b.data)
}

func (b *buffer) cap() int {
	return cap(b.data)
}

func (b *buffer) flush(w io.Writer, n int) {
	w.Write(b.data[:n])
	n = copy(b.data, b.data[n:])
	b.data = b.data[:n]
}
