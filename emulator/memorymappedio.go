package emulator

import (
	"github.com/nolag/gocpu/memory"
)

// StdinRead is the word that is mapped to reading stdio
const StdinRead = 0xffff0007

// StdoutWrite is the word mapped to writing to stdout
const StdoutWrite = 0xffff000F

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
		readVal := make([]byte, 1)
		_, err := mem.Stdin.Read(readVal)
		return readVal[0], err
	} else if index < StdinRead && index > StdinRead-4 {
		return 0, nil
	}

	return mem.BackingMemory.ReadOneByte(index)
}

// ReadRaw allows reading from memory starting at startIndex and providing numBytes bytes
// data is the bytes read
// backed, when true means that changes made to data will impact the memory stored
// err is any error that occured
// Reads from the word StdinRead will return the next byte from standard output, padded to a word.
func (mem *MemoryMappedIO) ReadRaw(startIndex uint64, numBytes uint64) (data []byte, backed bool, err error) {
	if !validMemoryLoc(startIndex, numBytes) {
		return nil, false, &memory.AccessViolationError{Location: startIndex, NumBytes: numBytes, WasRead: true}
	}

	endIndex := startIndex + numBytes - 1

	if endIndex == StdinRead {
		readVal := make([]byte, 1)
		_, err := mem.Stdin.Read(readVal)
		retVal := make([]byte, numBytes)
		retVal[0] = readVal[0]
		return retVal, false, err
	}

	data, backed, err = mem.BackingMemory.ReadRaw(startIndex, numBytes)
	if endIndex == StdoutWrite {
		backed = false
		temp := data
		data = make([]byte, numBytes)
		copy(data, temp)
	}

	return data, backed, err
}

// Size in bytes this memory can represent
func (mem *MemoryMappedIO) Size() uint64 {
	return mem.BackingMemory.Size()
}

// WriteOneByte reads a byte at memory location index
func (mem *MemoryMappedIO) WriteOneByte(val byte, index uint64) error {
	if index == StdoutWrite {
		output := make([]byte, 1)
		output[0] = val
		_, err := mem.Stdout.Write(output)
		return err
	} else if index < StdoutWrite && index > StdoutWrite-4 {
		return nil
	}
	return mem.BackingMemory.WriteOneByte(val, index)
}

// WriteRaw writes data back to memory
func (mem *MemoryMappedIO) WriteRaw(data []byte, startIndex uint64) error {
	numBytes := uint64(len(data))
	if !validMemoryLoc(startIndex, uint64(numBytes)) {
		return &memory.AccessViolationError{Location: startIndex, NumBytes: numBytes, WasRead: false}
	}

	if startIndex+uint64(len(data))-1 == StdoutWrite {
		output := make([]byte, 1)
		output[0] = data[numBytes-1]
		_, err := mem.Stdout.Write(output)
		return err
	}

	return mem.BackingMemory.WriteRaw(data, startIndex)
}

func validMemoryLoc(startIndex uint64, numBytes uint64) bool {
	return numBytes == 1 || numBytes == 2 && startIndex&1 == 0 || numBytes == 4 && startIndex&3 == 0
}
