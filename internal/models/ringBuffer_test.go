package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_RingBuffer_Add(t *testing.T) {
	rb := NewRingBuffer(3)

	for i := 1; i < 5; i++ {
		rb.Add(float64(i))
	}

	values := rb.get()
	assert.Equal(t, []float64{4, 2, 3}, values, "Ожидались значения [4, 2, 3]")
}

func Test_RingBuffer_Get(t *testing.T) {
	rb := NewRingBuffer(3)

	for i := 1; i < 4; i++ {
		rb.Add(float64(i))
	}

	values := rb.get()
	assert.Equal(t, []float64{1.0, 2.0, 3.0}, values, "Ожидались значения [1.0, 2.0, 3.0]")
}

func Test_RingBuffer_Average(t *testing.T) {
	rb := NewRingBuffer(3)

	for i := 1; i < 5; i++ {
		rb.Add(float64(i))
	}

	avg := rb.GetAverage()
	assert.Equal(t, 3.0, avg, "Ожидаемое среднее 3.0")
}

func Test_RingBuffer_AverageWith10Values(t *testing.T) {
	rb := NewRingBuffer(5)

	for i := 1; i <= 10; i++ {
		rb.Add(float64(i))
	}

	avg := rb.GetAverage()
	assert.Equal(t, 8.0, avg, "Ожидаемое среднее 8.0")
}
