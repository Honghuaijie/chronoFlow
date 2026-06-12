package process

type LogBuffer struct {
	maxBytes  int
	content   []byte
	truncated bool
}

func NewLogBuffer(maxBytes int64) *LogBuffer {
	if maxBytes <= 0 {
		maxBytes = 5 * 1024 * 1024
	}
	return &LogBuffer{maxBytes: int(maxBytes)}
}

func (b *LogBuffer) Write(p []byte) (int, error) {
	b.content = append(b.content, p...)
	if len(b.content) > b.maxBytes {
		b.truncated = true
		head := b.maxBytes / 2
		tail := b.maxBytes - head
		next := make([]byte, 0, b.maxBytes)
		next = append(next, b.content[:head]...)
		next = append(next, b.content[len(b.content)-tail:]...)
		b.content = next
	}
	return len(p), nil
}

func (b *LogBuffer) String() string {
	return string(b.content)
}

func (b *LogBuffer) Truncated() bool {
	return b.truncated
}
