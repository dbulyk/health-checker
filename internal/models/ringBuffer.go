package models

type RingBuffer struct {
	data  []float64
	size  int
	start int
	end   int
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]float64, size),
		size: size,
	}
}

func (rb *RingBuffer) Add(value float64) {
	rb.data[rb.end] = value
	rb.end = (rb.end + 1) % rb.size
	if rb.end == rb.start {
		rb.start = (rb.start + 1) % rb.size
	}
}

func (rb *RingBuffer) get() []float64 {
	return rb.data
}

func (rb *RingBuffer) GetAverage() float64 {
	sum := 0.0
	values := rb.get()
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
