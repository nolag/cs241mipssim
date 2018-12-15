package emulator

import (
	"encoding/binary"
	"testing"

	"github.com/nolag/gocpu/memory"
	"github.com/nolag/gocpu/registers"
	"github.com/nolag/gomips"
	"github.com/stretchr/testify/assert"
)

func TestNewZeroed(t *testing.T) {
	memorySize := uint32(2048)
	emulator := NewZeroed(memorySize)
	processor := &emulator.processor
	VerifyProcessorSetup(t, processor, memorySize)
	for i := 1; i < 32; i++ {
		assert.Equal(t, uint32(0), processor.Registers[i].ValueAsUint32(), "Registers must be zeroed out.")
	}
}

func TestTwoIntsSetup(t *testing.T) {
	memorySize := uint32(2048)
	anyProgram := []byte{123, 134, 255, 21, 10, 10, 21, 30}
	anyVal1 := int32(12352)
	anyVal2 := int32(-3234)
	emulator := NewTwoInts(memorySize, anyProgram, anyVal1, anyVal2)
	processor := &emulator.processor
	VerifyProcessorSetup(t, processor, memorySize)

	assert.Equal(
		t, uint32(anyVal1), processor.Registers[1].ValueAsUint32(), "Register one is not set correctly")
	assert.Equal(
		t, uint32(anyVal2), processor.Registers[2].ValueAsUint32(), "Register two is not set correctly")
	verifyCs241(t, processor, anyProgram, memorySize)
}

func TestArray(t *testing.T) {
	memorySize := uint32(2048)
	anyProgram := []byte{123, 134, 255, 21, 10, 10, 21, 30}
	values := []int32{5321, 33241, -1, -432}
	emulator := NewArrayInts(memorySize, anyProgram, values)
	processor := &emulator.processor
	VerifyProcessorSetup(t, processor, memorySize)
	assert.Equal(t, uint32(len(values)), processor.Registers[2].ValueAsUint32(), "Register two is not set correctly")
	arrayLoc := processor.Registers[1].ValueAsUint32()
	assert.Equal(
		t,
		uint32(len(anyProgram)),
		arrayLoc,
		"Register one is not set correctly, storred array should start after the program")

	for index, value := range values {
		location := uint64(arrayLoc + uint32(4*index))
		actual, err := memory.ReadUint32(processor.Memory, processor.ByteOrder, location)
		assert.Nil(t, err, "No error should occur reading from memory")
		assert.Equal(t, value, int32(actual), "Wrong value stored in the int array")
	}

	verifyCs241(t, processor, anyProgram, memorySize)
}

func VerifyProcessorSetup(t *testing.T, processor *gomips.Processor, memorySize uint32) {
	assert.Equal(t, binary.BigEndian, processor.ByteOrder, "CS241 uses BigEndian byte ordering.")
	for i := 0; i < 4; i++ {
		assert.Nil(t, processor.Coprocessors[i], "Coprocessors are not used in CS241.")
	}

	for i := 0; i < 32; i++ {
		assert.Nil(t, processor.FloatRegisters[i], "Float registers are not used in CS241.")
	}

	for i := 1; i < 32; i++ {
		_, ok := processor.Registers[i].(registers.ZeroRegister)
		assert.False(t, ok, "Non zero registers must not must not be a zero register.")
	}

	verifyZeroValueAndNonZeroRegister(t, processor.Hi, "Hi")
	assert.Equal(t, false, processor.InBranchDelay, "Delay branch is not used in cs241 mips.")
	verifyZeroValueAndNonZeroRegister(t, processor.Low, "Low")
	assert.Equal(t, uint64(memorySize), processor.Memory.Size(), "Wrong memory size payments")
	verifyZeroValueAndNonZeroRegister(t, processor.Pc, "Pc")
	_, ok := processor.Registers[0].(*registers.ZeroRegister)
	assert.True(t, ok, "Register 0 must be a zero register")
}

func verifyZeroValueAndNonZeroRegister(t *testing.T, register registers.IIntRegister32, name string) {
	assert.Equal(t, uint32(0), register.ValueAsUint32(), name+" must be initilized as 0.")
	_, ok := register.(registers.ZeroRegister)
	assert.False(t, ok, name+" must not be a zero register")
}

func verifyCs241(t *testing.T, processor *gomips.Processor, program []byte, memorySize uint32) {
	for i := 3; i < 30; i++ {
		assert.Equal(t, uint32(0), processor.Registers[i].ValueAsUint32(), "Registers must be zeroed out.")
	}
	assert.Equal(
		t,
		uint32(0xFFFFFFFF),
		processor.Registers[31].ValueAsUint32(),
		"Return address must be the last address of memory possible")
	assert.Equal(t, memorySize, processor.Registers[30].ValueAsUint32(), "Stack must be at end of memory")
	data, _, err := processor.Memory.ReadRaw(0, uint64(len(program)))
	assert.Nil(t, err, "The program should be read with no errors")
	assert.ElementsMatch(t, program, data, "Wrong program loaded")
	// TODO verify that the memory has the correct write/read memory map

}
