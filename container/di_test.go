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
	LowLevelComponent ILowLevel `dep:""`
}

func (tlc TopLevelComponent) Constructor() *TopLevelComponent {
	return &TopLevelComponent{tlc.LowLevelComponent}
}

func TestExample(t *testing.T) {
	dic := NewDIContainer()

	tlc := new(TopLevelComponent)
	dic.Register(new(LowLevelComponent))
	dic.Register(tlc)
	dic.Populate()

	tlc = tlc.Constructor()
	assert.Equal(t, 12, tlc.LowLevelComponent.Counter())

	// use Get
	tlcPtr, _ := dic.Get("github.com/pot-code/go-di-pattern/container/TopLevelComponent")
	tlc = tlcPtr.(*TopLevelComponent)
	assert.Equal(t, 12, tlc.LowLevelComponent.Counter())
}
