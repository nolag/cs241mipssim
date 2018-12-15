package emulator

import (
	"errors"
	"testing"

	"github.com/nolag/gocpu/memory"
	"github.com/nolag/gocpu/memory/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestRunReadWriteTest(t *testing.T) {
	testhelper.RunReadWriteTest(t, true, createMappedForTest)
}

func TestNewMeorySliceCreatesMemoryWithCorrectSize(t *testing.T) {
	testhelper.RunSizeTest(t, createMappedForTest)
}

func TestSliceBoundsChceking(t *testing.T) {
	testhelper.RunBoundsTests(t, createMappedForTest)
}

func TestReadOneByteReadsFromStdin(t *testing.T) {
	anyVal := byte(10)
	reader := fakeReadable{NextVal: anyVal}
	mem := MemoryMappedIO{Stdin: reader}

	actual, err := mem.ReadOneByte(StdinRead)

	assert.NoError(t, err, "Memory must not return an error when stdin has no error")
	assert.Equal(t, anyVal, actual, "Mismatched memory value returned from stdin")
}

func TestReadOneByteRetunsErrorFromStdin(t *testing.T) {
	anyErr := errors.New("Test")
	reader := fakeReadable{Err: anyErr}
	mem := MemoryMappedIO{Stdin: reader}

	_, err := mem.ReadOneByte(StdinRead)

	assert.Equal(t, anyErr, err, "Error from reading stdin must be propogated")
}

func TestReadOneByteReturnsZeroForRemainingWordPartForStdin(t *testing.T) {
	mem := MemoryMappedIO{}

	for i := 1; i < 4; i++ {
		actual, err := mem.ReadOneByte(StdinRead + uint64(i))
		assert.NoError(t, err, "Memory must not return an error when reading off bytes in stdin word")
		assert.Equal(
			t,
			byte(0),
			actual,
			"To match reading from the word, 0 must be returned when reading off bytes in stdin word")
	}
}

// TODO test raw read from location and crossing boundary
func TestReadRawReadsFromStdin(t *testing.T) {
	anyVal := byte(10)
	reader := fakeReadable{NextVal: anyVal}
	mem := MemoryMappedIO{Stdin: reader}

	actual, backed, err := mem.ReadRaw(StdinRead, 4)

	assert.ElementsMatch(t, []byte{anyVal, 0, 0, 0}, actual, "Read raw must read from stdin")
	assert.NoError(t, err, "Memory must not return an error when stdin has no error")
	assert.False(t, backed, "Memory from IO is not backed")
}

func TestReadRawRetunsErrorFromStdin(t *testing.T) {
	anyErr := errors.New("Test")
	reader := fakeReadable{Err: anyErr}
	mem := MemoryMappedIO{Stdin: reader}

	_, _, err := mem.ReadRaw(StdinRead, 4)

	assert.Equal(t, anyErr, err, "Error from reading stdin must be propogated")
}

func TestReadRawReturnsZeroForRemainingWordPartForStdin(t *testing.T) {
	mem := MemoryMappedIO{}

	for i := 1; i < 4; i++ {
		actual, _, err := mem.ReadRaw(StdinRead+uint64(i), 4-uint64(i))
		assert.NoError(t, err, "Memory must not return an error when reading off bytes in stdin word")
		assert.Equal(
			t,
			make([]byte, 4-uint64(i)),
			actual,
			"To match reading from the word, 0 must be returned when reading off bytes in stdin word")
	}
}

func TestReadRawReturnsAccessErrorIfCrossingBoundaryForMemoryRead(t *testing.T) {
	mem := MemoryMappedIO{}
	readLoc := StdinRead + uint64(1)
	readLoc2 := uint64(StdinRead)
	readLoc3 := StdinRead - uint64(1)
	_, _, err := mem.ReadRaw(readLoc, 4)
	_, _, err2 := mem.ReadRaw(readLoc2, 5)
	_, _, err3 := mem.ReadRaw(readLoc3, 2)

	expected := memory.AccessViolationError{Location: readLoc, NumBytes: 4, WasRead: true}
	expected2 := memory.AccessViolationError{Location: readLoc2, NumBytes: 5, WasRead: true}
	expected3 := memory.AccessViolationError{Location: readLoc3, NumBytes: 2, WasRead: true}
	assert.Equal(t, &expected, err)
	assert.Equal(t, &expected2, err2)
	assert.Equal(t, &expected3, err3)
}

// TODO test single byte writes
// TODO test raw writes from location and crossing boundary

func createMappedForTest(size uint64) memory.Memory {
	return &MemoryMappedIO{BackingMemory: memory.NewSlice(size)}
}

type fakeReadable struct {
	NextVal byte
	Err     error
}

func (readable fakeReadable) Read(b []byte) (n int, err error) {
	b[0] = readable.NextVal
	return 1, readable.Err
}
