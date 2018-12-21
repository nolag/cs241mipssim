package emulator

import (
	"github.com/nolag/gocpu/memory"
)

type boundaryType int

const (
	crossesBoundary      boundaryType = iota
	outsideBoundry                    = iota
	laterBytesInBoundary              = iota
	firstIsBoundry                    = iota
)

// StdinRead is the word that is mapped to reading stdio
const StdinRead = 0xffff0004

// StdoutWrite is the word mapped to writing to stdout
const StdoutWrite = 0xffff000c

// Readable wraps reading from a file.
type Readable interface {
	Read(b []byte) (n int, err error)
}

// Writable wraps writing to a file.
type Writable interface {
	Write(b []byte) (n int, err error)
}

// MemoryMappedIO maps memory to IO to simulate CS241 mips.
type MemoryMappedIO struct {
	BackingMemory memory.Memory
	Stdin         Readable
	Stdout        Writable
}

// ReadOneByte reads a byte at memory location index. Reads from 0xffff0004 will
// return next byte from standard input, other bytes in that word will return 0.
func (mem *MemoryMappedIO) ReadOneByte(index uint64) (byte, error) {
	if index == StdinRead {
		val := make([]byte, 1)
		_, err := mem.Stdin.Read(val)
		return val[0], err
	}

	if index > StdinRead && index < StdinRead+4 {
		return 0, nil
	}

	return mem.BackingMemory.ReadOneByte(index)
}

// ReadRaw allows reading from memory starting at startIndex and providing numBytes bytes
// data is the bytes read
// backed, when true means that changes made to data will impact the memory stored
// err is any error that occured
// Reads from the word StdinRead will return the next byte from standard output, padded to a word.
// Crossing the StdinRead word boundary will cause an error to be returned
func (mem *MemoryMappedIO) ReadRaw(startIndex uint64, numBytes uint64) (data []byte, backed bool, err error) {
	boundaryCheck := bounderyCheck(startIndex, numBytes, StdinRead)
	switch boundaryCheck {
	case crossesBoundary:
		return nil, false, &memory.AccessViolationError{Location: startIndex, NumBytes: numBytes, WasRead: true}
	case firstIsBoundry:
		read := make([]byte, 1)
		_, err := mem.Stdin.Read(read)
		result := make([]byte, numBytes)
		copy(result, read)
		return result, false, err
	case laterBytesInBoundary:
		return make([]byte, numBytes), false, nil
	}

	val, backed, err := mem.BackingMemory.ReadRaw(startIndex, numBytes)

	if err == nil {
		boundaryCheck = bounderyCheck(startIndex, numBytes, StdoutWrite)
		switch boundaryCheck {
		case crossesBoundary:
			fallthrough
		case firstIsBoundry:
			fallthrough
		case laterBytesInBoundary:
			backed = false
			val2 := make([]byte, numBytes)
			copy(val2, val)
			val = val2
		}
	}

	return val, backed, err
}

func bounderyCheck(startIndex uint64, numBytes uint64, boundaryStart uint64) boundaryType {
	boundaryEnd := boundaryStart + 3
	endIndex := startIndex + numBytes - 1

	if (startIndex < boundaryStart && endIndex >= boundaryStart) ||
		((startIndex >= boundaryStart && startIndex <= boundaryEnd) && endIndex > boundaryEnd) {
		return crossesBoundary
	}

	if startIndex == boundaryStart {
		return firstIsBoundry
	}

	if startIndex > boundaryStart && startIndex <= boundaryEnd {
		return laterBytesInBoundary
	}

	return outsideBoundry
}

// Size in bytes this memory can represent
func (mem *MemoryMappedIO) Size() uint64 {
	return mem.BackingMemory.Size()
}

// WriteOneByte reads a byte at memory location index
func (mem *MemoryMappedIO) WriteOneByte(val byte, index uint64) error {
	return mem.BackingMemory.WriteOneByte(val, index)
}

// WriteRaw writes data back to memory
func (mem *MemoryMappedIO) WriteRaw(data []byte, startIndex uint64) error {
	return mem.BackingMemory.WriteRaw(data, startIndex)
}
