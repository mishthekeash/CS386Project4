package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
)

// ************* Main program and CPU implementation *************
//
// You will not need to read most of this file to implement your kernel,
// although you may want to be somewhat familiar with the implementation of
// `cpu.step`.

// Prints an error message and exits with status code 1.
func fatal(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func main() {
	// Parse command-line arguments.

	var debugFlag = flag.Bool("debug", false, "run the debugger")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fatal("Expected final arguments to be '<kernel/bootloader> <program>'")
	}

	kernelOrBootloaderFile := args[0]
	progFile := args[1]

	// Load kernel/bootloader and program files.

	kernelOrBootloader, err := os.ReadFile(kernelOrBootloaderFile)
	if err != nil {
		fatal("Could not read kernel/bootloader:", err)
	}
	prog, err := os.ReadFile(progFile)
	if err != nil {
		fatal("Could not read program:", err)
	}

	// Assemble kernel/bootloader and program.

	kernelOrBootloaderInstrs, err := instructionSet.parseInstrSeq(string(kernelOrBootloader))
	if err != nil {
		fatal("Could not parse kernel/bootloader:", err)
	}
	progInstrs, err := instructionSet.parseInstrSeq(string(prog))
	if err != nil {
		fatal("Could not parse program:", err)
	}

	// Load kernel/bootloader into memory.

	memory := make([]word, 2048)
	if len(kernelOrBootloaderInstrs) > 1024 {
		fatal("Kernel/bootloader is larger than 1KB")
	}
	copy(memory[:1024], kernelOrBootloaderInstrs)

	instrBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(instrBytes, uint16(len(progInstrs)))
	for _, instr := range progInstrs {
		instr := instr.toBeBytes()
		instrBytes = append(instrBytes, instr[:]...)
	}

	// Create and run the CPU emulator.

	c := boot(initKernelCpuState, memory)
	if *debugFlag {
		c.read = bytes.NewBuffer(instrBytes)
		debug(&c)
	} else {
		c.read = io.MultiReader(bytes.NewBuffer(instrBytes), os.Stdin)
		for {
			halt, err := c.step()
			if err != nil {
				fmt.Printf("CPU entered a bad state: %v; halting...\n", err)
				return
			}
			if halt {
				return
			}
		}
	}
}

// A machine word.
type word uint32

func (w word) String() string {
	return fmt.Sprintf("%08x", uint32(w))
}

// Convert this word to a byte array in big-endian byte order.
func (w word) toBeBytes() [4]byte {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], uint32(w))
	return bytes
}

// Convert a byte array in big-endian byte order to a word.
func beBytesToWord(bytes [4]byte) word {
	return word(binary.BigEndian.Uint32(bytes[:]))
}

// The state of the CPU.
type cpu struct {
	registers    [8]word
	memory       []word
	kernel       kernelCpuState
	halted       bool
	instructions instrSet

	// The devices used by the `read` and `write` instructions.
	read  io.Reader
	write io.Writer
}

func (c *cpu) String() string {
	var registers []string
	for i, r := range c.registers {
		registers = append(registers, fmt.Sprintf("r%v=%v", i, r))
	}

	return fmt.Sprintf("%v | kernel: %v", registers, c.kernel)
}

// Initializes a new CPU.
func boot(initKernelState kernelCpuState, memory []word) cpu {
	return cpu{
		kernel:       initKernelState,
		memory:       memory,
		instructions: instructionSet,
		read:         os.Stdin,
		write:        os.Stdout,
	}
}

// Executes one instruction.
//
// Returns `true` if the executed instruction performed an explicit halt, and
// `false` otherwise. If an error is returned, the CPU has gotten into a bad
// state, and should halt.
func (c *cpu) step() (bool, error) {
	debugPrintf("[cpu.step] CPU state: %v\n", c)

	if c.halted {
		panic("cannot execute CPU once it has been halted")
	}

	// Give the kernel subsystem a chance to skip this instruction.
	skip, err := c.kernel.preExecuteHook(c)
	if err != nil {
		c.halted = true
		return true, fmt.Errorf("kernel pre-execute hook failed: %v", err)
	}
	if skip {
		debugPrintf("[cpu.step] Kernel pre-execute hook decided to skip this instruction\n")
		return c.halted, nil
	}

	iptr := c.registers[7]

	if int(iptr) >= len(c.memory) {
		c.halted = true
		return true, fmt.Errorf("iptr out of bounds: %v", iptr)
	}

	i, err := c.instructions.decode(c.memory[int(iptr)])
	if err != nil {
		c.halted = true
		return true, fmt.Errorf("failed to decode instruction: %v", err)
	}

	debugPrintf("[cpu.step] decoded instruction: %v\n", &i)

	iptr++
	c.registers[7] = iptr

	err = i.Run(c)
	if err != nil {
		c.halted = true
		return true, fmt.Errorf("failed to execute instruction: %v", err)
	}

	return c.halted, nil
}

// This is enabled by the debugger in `debug.go`.
var debugEnabled bool = false

func debugPrintf(format string, a ...any) (n int, err error) {
	if debugEnabled {
		return fmt.Printf("[DEBUG] "+format, a...)
	}
	return 0, nil
}
