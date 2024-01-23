package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRingBuffer_Add(t *testing.T) {
	rb := NewRingBuffer(3)

	rb.Add(1.0)
	rb.Add(2.0)
	rb.Add(3.0)
	rb.Add(4.0)

	values := rb.get()
	assert.Equal(t, []float64{4, 2, 3}, values, "Ожидались значения [4, 2, 3]")
}

func TestRingBuffer_Get(t *testing.T) {
	rb := NewRingBuffer(3)

	rb.Add(1.0)
	rb.Add(2.0)
	rb.Add(3.0)

	values := rb.get()
	assert.Equal(t, []float64{1.0, 2.0, 3.0}, values, "Ожидались значения [1.0, 2.0, 3.0]")
}

func TestRingBuffer_Average(t *testing.T) {
	rb := NewRingBuffer(3)

	rb.Add(1.0)
	rb.Add(2.0)
	rb.Add(3.0)
	rb.Add(4.0)

	avg := rb.GetAverage()
	assert.Equal(t, 3.0, avg, "Ожидаемое среднее 3.0")
}
