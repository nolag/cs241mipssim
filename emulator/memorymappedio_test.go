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

func TestSize(t *testing.T) {
	testhelper.RunSizeTest(t, createMappedForTest)
}

func TestBoundsChceking(t *testing.T) {
	testhelper.RunBoundsTests(t, createMappedForTest)
}

func TestAllignmentIsPowersOfTwoUntilFourBytes(t *testing.T) {
	testhelper.RunPowTwoAllignmentFailures(t, 2, createMappedForTest)
}

func TestReadWriteCannotBeEightOrMoreBytes(t *testing.T) {
	testhelper.RunDisallowedSize(t, 8, createMappedForTest)
	testhelper.RunDisallowedSize(t, 16, createMappedForTest)
}

func TestReadWriteMustMustBePowerOfTwo(t *testing.T) {
	anyNonPowerOfTwoLessThanFour := uint64(3)
	testhelper.RunDisallowedSize(t, anyNonPowerOfTwoLessThanFour, createMappedForTest)
}

func TestReadOneByteReadsFromStdin(t *testing.T) {
	anyVal := byte(10)
	reader := fakeReadable{NextVal: anyVal, Tester: t}
	mem := MemoryMappedIO{Stdin: reader}

	actual, err := mem.ReadOneByte(StdinRead)

	assert.NoError(t, err, "Memory must not return an error when stdin has no error")
	assert.Equal(t, anyVal, actual, "Mismatched memory value returned from stdin")
}

func TestReadOneByteRetunsErrorFromStdin(t *testing.T) {
	anyErr := errors.New("Test")
	reader := fakeReadable{Err: anyErr, Tester: t}
	mem := MemoryMappedIO{Stdin: reader}

	_, err := mem.ReadOneByte(StdinRead)

	assert.Equal(t, anyErr, err, "Error from reading stdin must be propogated")
}

func TestReadOneByteReturnsZeroForRemainingWordPartForStdin(t *testing.T) {
	mem := MemoryMappedIO{}

	for i := 1; i < 4; i++ {
		actual, err := mem.ReadOneByte(StdinRead - uint64(i))
		assert.NoError(t, err, "Memory must not return an error when reading off bytes in stdin word")
		assert.Equal(
			t,
			byte(0),
			actual,
			"To match reading from the word, 0 must be returned when reading off bytes in stdin word")
	}
}

func TestReadRawReadsFromStdin(t *testing.T) {
	anyVal := byte(10)
	reader := fakeReadable{NextVal: anyVal, Tester: t}
	mem := MemoryMappedIO{Stdin: reader}

	for i := uint64(1); i <= 4; i <<= 1 {
		actual, backed, err := mem.ReadRaw(StdinRead-i+1, i)
		expected := make([]byte, i)
		expected[i-1] = anyVal
		assert.ElementsMatch(t, expected, actual, "Read raw must read from stdin with %v bytes", i)
		assert.NoError(t, err, "Memory must not return an error when stdin has no error with %v bytes", i)
		assert.False(t, backed, "Memory from IO is not backed with %v bytes", i)
	}
}

func TestReadRawRetunsErrorFromStdin(t *testing.T) {
	anyErr := errors.New("Test")
	reader := fakeReadable{Err: anyErr, Tester: t}
	mem := MemoryMappedIO{Stdin: reader}

	_, _, err := mem.ReadRaw(StdinRead-3, 4)

	assert.Equal(t, anyErr, err, "Error from reading stdin must be propogated")
}

func TestReadRawReturnsFalseForBackingReadFromStdout(t *testing.T) {
	anyVal := byte(10)
	slice := memory.NewSlice(StdoutWrite + 1)
	slice.WriteOneByte(anyVal, StdoutWrite)
	mem := MemoryMappedIO{BackingMemory: slice}

	actual, backed, err := mem.ReadRaw(StdoutWrite-3, 4)

	assert.ElementsMatch(t, []byte{0, 0, 0, anyVal}, actual, "Read must read from memory")
	assert.NoError(t, err, "Memory must not return an error when reading from stdin boundary")
	assert.False(t, backed, "Memory from IO is not backed")
	prior := actual[3]
	actual[3] = 100
	afterWrite, _ := mem.ReadOneByte(StdoutWrite)
	assert.Equal(t, prior, afterWrite, "Non backing memory must not change the value when written to")
}
func TestWriteOneByteWritesToStdout(t *testing.T) {
	anyVal := byte(10)
	writer := fakeWriteable{Tester: t}
	mem := MemoryMappedIO{Stdout: &writer}

	err := mem.WriteOneByte(anyVal, StdoutWrite)

	assert.NoError(t, err, "Memory must not return an error when stdout has no error")
	assert.Equal(t, anyVal, writer.WrittenVal, "Wrong value written to stdout")
}

func TestWriteOneByteRetunsErrorFromStdout(t *testing.T) {
	anyVal := byte(10)
	anyErr := errors.New("Test")
	writer := fakeWriteable{Err: anyErr, Tester: t}
	mem := MemoryMappedIO{Stdout: &writer}

	err := mem.WriteOneByte(anyVal, StdoutWrite)

	assert.Equal(t, anyErr, err, "Error from reading stdout must be propogated")
}

func TestWriteOneByteReturnsNoErrorForRemainingWordPartForStdin(t *testing.T) {
	anyVal := byte(10)
	mem := MemoryMappedIO{}

	for i := 1; i < 4; i++ {
		err := mem.WriteOneByte(anyVal, StdoutWrite-uint64(i))
		assert.NoError(t, err, "Memory must not return an error when writting off bytes in stdout word")
	}
}
func TestWriteRawWritesToStdout(t *testing.T) {
	anyVal := byte(10)
	writer := fakeWriteable{Tester: t}
	mem := MemoryMappedIO{Stdout: &writer}

	for i := uint64(1); i <= 4; i <<= 1 {
		writer.WrittenVal = 0
		anyVals := make([]byte, i)
		anyVals[i-1] = anyVal

		err := mem.WriteRaw(anyVals, StdoutWrite-i+1)

		assert.NoError(t, err, "Memory must not return an error when stdout has no error for %v bytes", i)
		assert.Equal(t, anyVal, writer.WrittenVal, "wrong value written for %v bytes", i)
	}
}

func TestWriteRawRetunsErrorFromStdout(t *testing.T) {
	anyVals := []byte{10, 2, 21, 3}
	anyErr := errors.New("Test")
	writer := fakeWriteable{Err: anyErr, Tester: t}
	mem := MemoryMappedIO{Stdout: &writer}

	err := mem.WriteRaw(anyVals, StdoutWrite-3)

	assert.Equal(t, anyErr, err, "Error from reading stdout must be propogated")
}

func createMappedForTest(size uint64) memory.Memory {
	return &MemoryMappedIO{BackingMemory: memory.NewSlice(size)}
}

type fakeReadable struct {
	Err     error
	NextVal byte
	Tester  *testing.T
}

func (readable fakeReadable) Read(b []byte) (n int, err error) {
	assert.Equal(readable.Tester, 1, len(b), "This memory must only read one byte")
	b[0] = readable.NextVal
	return 1, readable.Err
}

type fakeWriteable struct {
	Err        error
	Tester     *testing.T
	WrittenVal byte
}

func (writeable *fakeWriteable) Write(b []byte) (n int, err error) {
	assert.Equal(writeable.Tester, 1, len(b), "This memory must only write one byte")
	writeable.WrittenVal = b[0]
	return len(b), writeable.Err
}
