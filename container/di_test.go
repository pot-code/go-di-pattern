package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type LowLevelComponent struct {
	Counter int
}

func (llc LowLevelComponent) Constructor() *LowLevelComponent {
	return &LowLevelComponent{12}
}

type TopLevelComponent struct {
	LowLevelComponent `dep:"LowLevelComponent"`
}

func (tlc TopLevelComponent) Constructor() *TopLevelComponent {
	return &TopLevelComponent{tlc.LowLevelComponent}
}

func TestGet(t *testing.T) {
	dic := NewDIContainer()

	dic.Register(new(LowLevelComponent))
	dic.Register(new(TopLevelComponent))

	tlc := dic.Get("TopLevelComponent").(*TopLevelComponent)
	assert.Equal(t, 12, tlc.Counter)
}
