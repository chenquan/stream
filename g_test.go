package stream

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRecover(t *testing.T) {
	i := 0
	Recover(func() {
		i++
	})
	assert.Equal(t, 1, i)

}
func TestNewGoroutine(t *testing.T) {
	i := 0
	defer func() {
		assert.Equal(t, 1, i)
	}()
	ch := make(chan struct{})
	go NewGoroutine(func() {
		defer func() {
			ch <- struct{}{}
		}()

		panic("panic")
	})
	<-ch
	i++
}
