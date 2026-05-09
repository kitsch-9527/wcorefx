//go:build windows

package evtx

// ByteBuffer 是一个由字节切片支持的可扩展缓冲区。
type ByteBuffer struct {
	buf    []byte
	offset int
}

// NewByteBuffer 创建一个具有指定初始容量的新ByteBuffer。
//   initialSize - 缓冲区的初始容量（字节数）
//   返回 - 新创建的ByteBuffer指针
func NewByteBuffer(initialSize int) *ByteBuffer {
	return &ByteBuffer{buf: make([]byte, initialSize)}
}

// Write 将数据追加到缓冲区，必要时自动扩容。
//   p - 待写入的字节数据
//   返回 - 实际写入的字节数；始终为nil的错误值
func (b *ByteBuffer) Write(p []byte) (int, error) {
	if len(b.buf) < b.offset+len(p) {
		spaceNeeded := len(b.buf) - b.offset + len(p)
		largerBuf := make([]byte, 2*len(b.buf)+spaceNeeded)
		copy(largerBuf, b.buf[:b.offset])
		b.buf = largerBuf
	}
	n := copy(b.buf[b.offset:], p)
	b.offset += n
	return n, nil
}

// Reset 重置缓冲区为空，保留底层存储。
//   返回 - 无
func (b *ByteBuffer) Reset() {
	b.offset = 0
	b.buf = b.buf[:cap(b.buf)]
}

// Bytes 返回已写入数据的切片。
//   返回 - 包含已写入字节数据的切片
func (b *ByteBuffer) Bytes() []byte {
	return b.buf[:b.offset]
}

// Len 返回已写入的字节数。
//   返回 - 缓冲区中已写入的字节数量
func (b *ByteBuffer) Len() int {
	return b.offset
}

// PtrAt 返回指定偏移处字节的指针。
//   offset - 字节偏移量
//   返回 - 指定偏移处的字节指针；越界时返回nil
func (b *ByteBuffer) PtrAt(offset int) *byte {
	if offset > b.offset-1 {
		return nil
	}
	return &b.buf[offset]
}

// Reserve 预留n字节空间，必要时分配新缓冲区。
//   n - 需要预留的字节数
func (b *ByteBuffer) Reserve(n int) {
	b.offset = n
	if n > cap(b.buf) {
		b.buf = make([]byte, n)
	} else {
		b.buf = b.buf[:n]
	}
}
