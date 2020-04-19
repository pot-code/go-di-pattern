package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type ILowLevel interface {
	Counter() int
}

type LowLevelComponent struct {
	counter int
}

func (llc LowLevelComponent) Constructor() *LowLevelComponent {
	return &LowLevelComponent{12}
}

func (llc *LowLevelComponent) Counter() int {
	return llc.counter
}

type TopLevelComponent struct {
	ILowLevel `dep:"LowLevelComponent"`
}

func (tlc TopLevelComponent) Constructor() *TopLevelComponent {
	return &TopLevelComponent{tlc.ILowLevel}
}

func TestGet(t *testing.T) {
	dic := NewDIContainer()

	dic.Register(new(LowLevelComponent))
	dic.Register(new(TopLevelComponent))

	tlc := dic.Get("TopLevelComponent").(*TopLevelComponent)
	assert.Equal(t, 12, tlc.Counter())
}
