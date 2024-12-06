package tokenizer

import (
	"strings"
)

type CyclicBuffer struct {
	buffer       []rune
	size         int
	readPointer  int
	writePointer int
	count        int
}

func NewCyclicBuffer(size int) *CyclicBuffer {
	return &CyclicBuffer{
		buffer:       make([]rune, size),
		size:         size,
		readPointer:  0,
		writePointer: 0,
		count:        0,
	}
}

func (c *CyclicBuffer) String() string {
	content := string(c.buffer[c.readPointer:]) + string(c.buffer[:c.readPointer])
	if len(content) > 10 {
		return content[:10] + "..."
	}
	return content
}

func (c *CyclicBuffer) Put(data ...rune) rune {
	for _, r := range data {
		c.push(r)
	}
	return data[0]
}

func (c *CyclicBuffer) push(data rune) {
	if c.count == c.size {
		c.readPointer = (c.readPointer + 1) % c.size
	} else if c.count < c.size {
		c.count++
	}
	c.buffer[c.writePointer] = data
	c.writePointer = (c.writePointer + 1) % c.size
}

func (c *CyclicBuffer) Has(subs ...string) bool {
	content := string(c.buffer[c.readPointer:]) + string(c.buffer[:c.readPointer])
	for _, sub := range subs {
		if !strings.Contains(content, sub) {
			return false
		}
	}
	return true
}
