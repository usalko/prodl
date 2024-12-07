package tokenizer

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

type BytesBuffer struct {
	buf      []byte
	off      int    // read at &buf[off], write at &buf[len(buf)]
	lastRead readOp // last read operation, so that Unread* can work correctly.
}

// The readOp constants describe the last action performed on
// the buffer, so that UnreadRune and UnreadByte can check for
// invalid usage. opReadRuneX constants are chosen such that
// converted to int they correspond to the rune size that was read.
type readOp int8

// Don't use iota for these, as the values need to correspond with the
// names and comments, which is easier to see when being explicit.
const (
	opRead      readOp = -1 // Any other read operation.
	opInvalid   readOp = 0  // Non-read operation.
	opReadRune1 readOp = 1  // Read rune of size 1.
	opReadRune2 readOp = 2  // Read rune of size 2.
	opReadRune3 readOp = 3  // Read rune of size 3.
	opReadRune4 readOp = 4  // Read rune of size 4.
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var errNegativeRead = errors.New("tokenizer.BytesBuffer: reader returned negative count from Read")

const maxInt = int(^uint(0) >> 1)

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
// The slice is valid for use only until the next buffer modification (that is,
// only until the next call to a method like [Buffer.Read], [Buffer.Write], [Buffer.Reset], or [Buffer.Truncate]).
// The slice aliases the buffer content at least until the next buffer modification,
// so immediate changes to the slice will affect the result of future reads.
func (bytesBuffer *BytesBuffer) Bytes() []byte { return bytesBuffer.buf[bytesBuffer.off:] }

// AvailableBuffer returns an empty buffer with b.Available() capacity.
// This buffer is intended to be appended to and
// passed to an immediately succeeding [Buffer.Write] call.
// The buffer is only valid until the next write operation on b.
func (bytesBuffer *BytesBuffer) AvailableBuffer() []byte {
	return bytesBuffer.buf[len(bytesBuffer.buf):]
}

// String returns the contents of the unread portion of the buffer
// as a string. If the [tokenizer.BytesBuffer] is a nil pointer, it returns "<nil>".
//
// To build strings more efficiently, see the [strings.Builder] type.
func (bytesBuffer *BytesBuffer) String() string {
	if bytesBuffer == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(bytesBuffer.buf[bytesBuffer.off:])
}

// String returns the contents of the buffer
// as a string. If the [tokenizer.BytesBuffer] is a nil pointer, it returns "<nil>".
//
// To build strings more efficiently, see the [strings.Builder] type.
func (bytesBuffer *BytesBuffer) StringAt(start, end int) string {
	if bytesBuffer == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(bytesBuffer.buf[start:end])
}

// empty reports whether the unread portion of the buffer is empty.
func (bytesBuffer *BytesBuffer) empty() bool { return len(bytesBuffer.buf) <= bytesBuffer.off }

// Size returns the number of bytes in the buffer;
func (bytesBuffer *BytesBuffer) Size() int { return len(bytesBuffer.buf) }

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (bytesBuffer *BytesBuffer) Len() int { return len(bytesBuffer.buf) - bytesBuffer.off }

// Cap returns the capacity of the buffer's underlying byte slice, that is, the
// total space allocated for the buffer's data.
func (bytesBuffer *BytesBuffer) Cap() int { return cap(bytesBuffer.buf) }

// Available returns how many bytes are unused in the buffer.
func (bytesBuffer *BytesBuffer) Available() int { return cap(bytesBuffer.buf) - len(bytesBuffer.buf) }

// Truncate discards all but the first n unread bytes from the buffer
// but continues to use the same allocated storage.
// It panics if n is negative or greater than the length of the buffer.
func (bytesBuffer *BytesBuffer) Truncate(n int) {
	if n == 0 {
		bytesBuffer.Reset()
		return
	}
	bytesBuffer.lastRead = opInvalid
	if n < 0 || n > bytesBuffer.Len() {
		panic("bytes.Buffer: truncation out of range")
	}
	bytesBuffer.buf = bytesBuffer.buf[:bytesBuffer.off+n]
}

// Reset resets the buffer to be empty,
// but it retains the underlying storage for use by future writes.
// Reset is the same as [Buffer.Truncate](0).
func (bytesBuffer *BytesBuffer) Reset() {
	bytesBuffer.buf = bytesBuffer.buf[:0]
	bytesBuffer.off = 0
	bytesBuffer.lastRead = opInvalid
}

// tryGrowByReslice is an inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (bytesBuffer *BytesBuffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(bytesBuffer.buf); n <= cap(bytesBuffer.buf)-l {
		bytesBuffer.buf = bytesBuffer.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (bytesBuffer *BytesBuffer) grow(n int) int {
	m := bytesBuffer.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && bytesBuffer.off != 0 {
		bytesBuffer.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := bytesBuffer.tryGrowByReslice(n); ok {
		return i
	}
	if bytesBuffer.buf == nil && n <= smallBufferSize {
		bytesBuffer.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(bytesBuffer.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(bytesBuffer.buf, bytesBuffer.buf[bytesBuffer.off:])
	} else if c > maxInt-c-n {
		panic(bytes.ErrTooLarge)
	} else {
		// Add b.off to account for b.buf[:b.off] being sliced off the front.
		bytesBuffer.buf = growSlice(bytesBuffer.buf[bytesBuffer.off:], bytesBuffer.off+n)
	}
	// Restore b.off and len(b.buf).
	bytesBuffer.off = 0
	bytesBuffer.buf = bytesBuffer.buf[:m+n]
	return m
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with [ErrTooLarge].
func (bytesBuffer *BytesBuffer) Grow(n int) {
	if n < 0 {
		panic("bytes.Buffer.Grow: negative count")
	}
	m := bytesBuffer.grow(n)
	bytesBuffer.buf = bytesBuffer.buf[:m]
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with [ErrTooLarge].
func (bytesBuffer *BytesBuffer) Write(p []byte) (n int, err error) {
	bytesBuffer.lastRead = opInvalid
	m, ok := bytesBuffer.tryGrowByReslice(len(p))
	if !ok {
		m = bytesBuffer.grow(len(p))
	}
	return copy(bytesBuffer.buf[m:], p), nil
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The return value n is the length of s; err is always nil. If the
// buffer becomes too large, WriteString will panic with [ErrTooLarge].
func (bytesBuffer *BytesBuffer) WriteString(s string) (n int, err error) {
	bytesBuffer.lastRead = opInvalid
	m, ok := bytesBuffer.tryGrowByReslice(len(s))
	if !ok {
		m = bytesBuffer.grow(len(s))
	}
	return copy(bytesBuffer.buf[m:], s), nil
}

// MinRead is the minimum slice size passed to a [Buffer.Read] call by
// [Buffer.ReadFrom]. As long as the [tokenizer.BytesBuffer] has at least MinRead bytes beyond
// what is required to hold the contents of r, [Buffer.ReadFrom] will not grow the
// underlying buffer.
const MinRead = 512

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with [ErrTooLarge].
func (bytesBuffer *BytesBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	bytesBuffer.lastRead = opInvalid
	for {
		i := bytesBuffer.grow(MinRead)
		bytesBuffer.buf = bytesBuffer.buf[:i]
		m, e := r.Read(bytesBuffer.buf[i:cap(bytesBuffer.buf)])
		if m < 0 {
			panic(errNegativeRead)
		}

		bytesBuffer.buf = bytesBuffer.buf[:i+m]
		n += int64(m)
		if e == io.EOF {
			return n, nil // e is EOF, so return nil explicitly
		}
		if e != nil {
			return n, e
		}
	}
}

// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
func growSlice(b []byte, n int) []byte {
	defer func() {
		if recover() != nil {
			panic(bytes.ErrTooLarge)
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]byte, n)...)
	// This avoids unnecessary zero-ing of the first len(b) bytes of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest size class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)]
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of bytes written; it always fits into an
// int, but it is int64 to match the [io.WriterTo] interface. Any error
// encountered during the write is also returned.
func (bytesBuffer *BytesBuffer) WriteTo(w io.Writer) (n int64, err error) {
	bytesBuffer.lastRead = opInvalid
	if nBytes := bytesBuffer.Len(); nBytes > 0 {
		m, e := w.Write(bytesBuffer.buf[bytesBuffer.off:])
		if m > nBytes {
			panic("bytes.Buffer.WriteTo: invalid Write count")
		}
		bytesBuffer.off += m
		n = int64(m)
		if e != nil {
			return n, e
		}
		// all bytes should have been written, by definition of
		// Write method in io.Writer
		if m != nBytes {
			return n, io.ErrShortWrite
		}
	}
	// Buffer is now empty; reset.
	bytesBuffer.Reset()
	return n, nil
}

// WriteByte appends the byte c to the buffer, growing the buffer as needed.
// The returned error is always nil, but is included to match [bufio.Writer]'s
// WriteByte. If the buffer becomes too large, WriteByte will panic with
// [ErrTooLarge].
func (bytesBuffer *BytesBuffer) WriteByte(c byte) error {
	bytesBuffer.lastRead = opInvalid
	m, ok := bytesBuffer.tryGrowByReslice(1)
	if !ok {
		m = bytesBuffer.grow(1)
	}
	bytesBuffer.buf[m] = c
	return nil
}

// WriteRune appends the UTF-8 encoding of Unicode code point r to the
// buffer, returning its length and an error, which is always nil but is
// included to match [bufio.Writer]'s WriteRune. The buffer is grown as needed;
// if it becomes too large, WriteRune will panic with [ErrTooLarge].
func (bytesBuffer *BytesBuffer) WriteRune(r rune) (n int, err error) {
	// Compare as uint32 to correctly handle negative runes.
	if uint32(r) < utf8.RuneSelf {
		bytesBuffer.WriteByte(byte(r))
		return 1, nil
	}
	bytesBuffer.lastRead = opInvalid
	m, ok := bytesBuffer.tryGrowByReslice(utf8.UTFMax)
	if !ok {
		m = bytesBuffer.grow(utf8.UTFMax)
	}
	bytesBuffer.buf = utf8.AppendRune(bytesBuffer.buf[:m], r)
	return len(bytesBuffer.buf) - m, nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is [io.EOF] (unless len(p) is zero);
// otherwise it is nil.
func (bytesBuffer *BytesBuffer) Read(p []byte) (n int, err error) {
	bytesBuffer.lastRead = opInvalid
	if bytesBuffer.empty() {
		// Buffer is empty, reset to recover space.
		bytesBuffer.Reset()
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n = copy(p, bytesBuffer.buf[bytesBuffer.off:])
	bytesBuffer.off += n
	if n > 0 {
		bytesBuffer.lastRead = opRead
	}
	return n, nil
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by [Buffer.Read].
// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
func (bytesBuffer *BytesBuffer) Next(n int) []byte {
	bytesBuffer.lastRead = opInvalid
	m := bytesBuffer.Len()
	if n > m {
		n = m
	}
	data := bytesBuffer.buf[bytesBuffer.off : bytesBuffer.off+n]
	bytesBuffer.off += n
	if n > 0 {
		bytesBuffer.lastRead = opRead
	}
	return data
}

// ReadByte reads and returns the next byte from the buffer.
// If no byte is available, it returns error [io.EOF].
func (bytesBuffer *BytesBuffer) ReadByte() (byte, error) {
	if bytesBuffer.empty() {
		// Buffer is empty, reset to recover space.
		bytesBuffer.Reset()
		return 0, io.EOF
	}
	c := bytesBuffer.buf[bytesBuffer.off]
	bytesBuffer.off++
	bytesBuffer.lastRead = opRead
	return c, nil
}

// ReadRune reads and returns the next UTF-8-encoded
// Unicode code point from the buffer.
// If no bytes are available, the error returned is io.EOF.
// If the bytes are an erroneous UTF-8 encoding, it
// consumes one byte and returns U+FFFD, 1.
func (bytesBuffer *BytesBuffer) ReadRune() (r rune, size int, err error) {
	if bytesBuffer.empty() {
		// Buffer is empty, reset to recover space.
		bytesBuffer.Reset()
		return 0, 0, io.EOF
	}
	c := bytesBuffer.buf[bytesBuffer.off]
	if c < utf8.RuneSelf {
		bytesBuffer.off++
		bytesBuffer.lastRead = opReadRune1
		return rune(c), 1, nil
	}
	r, n := utf8.DecodeRune(bytesBuffer.buf[bytesBuffer.off:])
	bytesBuffer.off += n
	bytesBuffer.lastRead = readOp(n)
	return r, n, nil
}

// UnreadRune unreads the last rune returned by [Buffer.ReadRune].
// If the most recent read or write operation on the buffer was
// not a successful [Buffer.ReadRune], UnreadRune returns an error.  (In this regard
// it is stricter than [Buffer.UnreadByte], which will unread the last byte
// from any read operation.)
func (bytesBuffer *BytesBuffer) UnreadRune() error {
	if bytesBuffer.lastRead <= opInvalid {
		return errors.New("bytes.Buffer: UnreadRune: previous operation was not a successful ReadRune")
	}
	if bytesBuffer.off >= int(bytesBuffer.lastRead) {
		bytesBuffer.off -= int(bytesBuffer.lastRead)
	}
	bytesBuffer.lastRead = opInvalid
	return nil
}

// RuneAt reads and returns the next UTF-8-encoded
// Unicode code point from the buffer.
func (bytesBuffer *BytesBuffer) RuneAt(pos int) rune {
	if bytesBuffer.empty() {
		return 0
	}
	c := bytesBuffer.buf[pos]
	if c < utf8.RuneSelf {
		return rune(c)
	}
	r, _ := utf8.DecodeRune(bytesBuffer.buf[pos:])
	return r
}

var errUnreadByte = errors.New("bytes.Buffer: UnreadByte: previous operation was not a successful read")

// UnreadByte unreads the last byte returned by the most recent successful
// read operation that read at least one byte. If a write has happened since
// the last read, if the last read returned an error, or if the read read zero
// bytes, UnreadByte returns an error.
func (bytesBuffer *BytesBuffer) UnreadByte() error {
	if bytesBuffer.lastRead == opInvalid {
		return errUnreadByte
	}
	bytesBuffer.lastRead = opInvalid
	if bytesBuffer.off > 0 {
		bytesBuffer.off--
	}
	return nil
}

// ReadBytes reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and including the delimiter.
// If ReadBytes encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often [io.EOF]).
// ReadBytes returns err != nil if and only if the returned data does not end in
// delim.
func (bytesBuffer *BytesBuffer) ReadBytes(delim byte) (line []byte, err error) {
	slice, err := bytesBuffer.readSlice(delim)
	// return a copy of slice. The buffer's backing array may
	// be overwritten by later calls.
	line = append(line, slice...)
	return line, err
}

// readSlice is like ReadBytes but returns a reference to internal buffer data.
func (bytesBuffer *BytesBuffer) readSlice(delim byte) (line []byte, err error) {
	i := bytes.IndexByte(bytesBuffer.buf[bytesBuffer.off:], delim)
	end := bytesBuffer.off + i + 1
	if i < 0 {
		end = len(bytesBuffer.buf)
		err = io.EOF
	}
	line = bytesBuffer.buf[bytesBuffer.off:end]
	bytesBuffer.off = end
	bytesBuffer.lastRead = opRead
	return line, err
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
// If ReadString encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often [io.EOF]).
// ReadString returns err != nil if and only if the returned data does not end
// in delim.
func (bytesBuffer *BytesBuffer) ReadString(delim byte) (line string, err error) {
	slice, err := bytesBuffer.readSlice(delim)
	return string(slice), err
}

// NewBuffer creates and initializes a new [tokenizer.BytesBuffer] using buf as its
// initial contents. The new [tokenizer.BytesBuffer] takes ownership of buf, and the
// caller should not use buf after this call. NewBuffer is intended to
// prepare a [tokenizer.BytesBuffer] to read existing data. It can also be used to set
// the initial size of the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new([tokenizer.BytesBuffer]) (or just declaring a [tokenizer.BytesBuffer] variable) is
// sufficient to initialize a [tokenizer.BytesBuffer].
func NewBytesBuffer(buf []byte) *BytesBuffer { return &BytesBuffer{buf: buf} }

// NewBufferString creates and initializes a new [tokenizer.BytesBuffer] using string s as its
// initial contents. It is intended to prepare a buffer to read an existing
// string.
//
// In most cases, new([tokenizer.BytesBuffer]) (or just declaring a [tokenizer.BytesBuffer] variable) is
// sufficient to initialize a [tokenizer.BytesBuffer].
func NewBytesBufferString(s string) *BytesBuffer {
	return &BytesBuffer{buf: []byte(s)}
}
