package emulator

import (
	"encoding/binary"

	"github.com/nolag/gocpu/memory"
	"github.com/nolag/gocpu/registers"
	"github.com/nolag/gomips"
)

// Emulator is a CS241 mips emulator, construct using NewTwoInts, NewArrayInts, or NewZeroed
type Emulator struct {
	processor gomips.Processor
}

// NewZeroed creates a new Emulator with all the registers and memory set to zero
func NewZeroed(memorySize uint32) *Emulator {
	hi := registers.RegisterUint32(0)
	low := registers.RegisterUint32(0)
	pc := registers.RegisterUint32(0)
	// TODO this should be a special memory that's backed by a slice that writes and reads at the special locations.
	memory := memory.NewSlice(uint64(memorySize))
	emulator := Emulator{
		gomips.Processor{
			ByteOrder: binary.BigEndian,
			Hi:        &hi,
			Low:       &low,
			Memory:    memory,
			Pc:        &pc,
		}}

	processor := &emulator.processor
	processor.Registers[0] = &registers.ZeroRegister{}
	for i := 1; i < 32; i++ {
		reg := registers.RegisterUint32(0)
		processor.Registers[i] = &reg
	}

	return &emulator
}

// NewTwoInts creates a new Emulator, like mips.twoints defined at https://www.student.cs.uwaterloo.ca/~cs241/a1/.NewTwoInts
// This does not load the program to memory.
func NewTwoInts(memorySize uint32, program []byte, reg1 int32, reg2 int32) *Emulator {
	emulator := NewZeroed(memorySize)
	processor := &emulator.processor
	processor.Registers[1].SetFromUint32(uint32(reg1))
	processor.Registers[2].SetFromUint32(uint32(reg2))
	setupForCs241(processor, program, memorySize)
	return emulator
}

// NewArrayInts creates a new Emulator, like mips.array defined at https://www.student.cs.uwaterloo.ca/~cs241/a1/
// This does not load the program to memory.
func NewArrayInts(memorySize uint32, program []byte, values []int32) *Emulator {
	emulator := NewZeroed(memorySize)
	processor := &emulator.processor
	location := len(program)
	onLocation := location
	for _, value := range values {
		memory.WriteUint32(processor.Memory, processor.ByteOrder, uint32(value), uint64(onLocation))
		onLocation += 4
	}

	processor.Registers[1].SetFromUint32(uint32(location))
	processor.Registers[2].SetFromUint32(uint32(len(values)))
	setupForCs241(processor, program, memorySize)
	return emulator
}

func setupForCs241(processor *gomips.Processor, program []byte, memorySize uint32) {
	processor.Memory.WriteRaw(program, 0)
	processor.Registers[31].SetFromUint32(uint32(0xFFFFFFFF))
	processor.Registers[30].SetFromUint32(memorySize)
}
