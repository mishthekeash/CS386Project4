package main

import (
	"fmt"
)

// ************* Kernel support *************
//
// All of your CPU emulator changes for Assignment 2 will go in this file.

/*
	Ok here we descibe what we have implemented
	we only do the bare minimum here as is requested in the assignment, all the rest
	is done in the kernel.asm

	// TODO: timer counter update  in kernel.asm after every instruction?
	//such a performance hog?

	//note a small problem in prime.asm. under windows enter key produces 13 not 10 so
	I modified prime.asm


	"you must store r7 in kernel memory as soon as you execute the trap handler,
	and you must use your copy in kernel memory when switching back to user mode.
	In other words, you can't just rely on a copy of r7 being stored in your
	kernelCpuState struct." that's how we do it yes

	test math instructions


*/

/*
reasons for switching into kernel mode
*/
const (
	syscall            = 0
	timer              = 1
	oobMemory          = 2
	illegalInstruction = 3
)
const debugflag = false

// The state kept by the CPU in order to implement kernel support.
type kernelCpuState struct {
	// TODO: Fill this in.

	kernelMode bool //mode we are in, user mode or kernel mode
	counter    word //timer counter
	r0         word //see the justification for this at usage points
	/* four values needed in the trap handler */
	trapHandler   word
	syscallNumber word
	trapReason    word
	returnAddress word
}

// The initial kernel state when the CPU boots.
var initKernelCpuState = kernelCpuState{
	// TODO: Fill this in.
	true, 0, 0, 0, 0, 0, 0,
}

/*
if we need to switch to kernel mode, trap handler is called and gives reason why
*/
func activateTrapHandler(c *cpu, reason word) {
	c.kernel.kernelMode = true

	// write the return address to the variable
	if reason == timer {
		//c.registers[7]-- //why? because if timer fired we want to
		//return to the same instruction again, this time to execute it
	}
	c.memory[c.kernel.returnAddress] = c.registers[7]

	// Printing values
	//fmt.Println("c.kernel.returnAddress:", c.kernel.returnAddress)
	//fmt.Println("c.registers[7]:", c.registers[7])
	if debugflag {
		fmt.Print("Reason: ", reason, " ")
	}

	//var name string

	//fmt.Fscan(os.Stdin, &name)
	//os.Exit(0)

	//set the instr pointer r7 to syscall handler in the
	//iptr
	c.registers[7] = c.kernel.trapHandler

	// write the reason address to the  variable
	c.memory[c.kernel.trapReason] = reason

	c.kernel.r0 = c.registers[0] //see the justification for this at usage points

}

// A hook which is executed at the beginning of each instruction step.
//
// This permits the kernel support subsystem to perform extra validation that is
// not part of the core CPU emulator functionality.
//
// If `preExecuteHook` returns an error, the CPU is considered to have entered
// an illegal state, and it halts.
//
// If `preExecuteHook` returns `true`, the instruction is "skipped": `cpu.step`
// will immediately return without any further execution.
func (k *kernelCpuState) preExecuteHook(c *cpu) (bool, error) {

	//ok so here we have the instruction counter to keep track of the timer

	//ToDO: ensure no problem with syscall and counter
	if k.kernelMode {
		return false, nil
	}

	//fmt.Println(c.registers)

	//timer handling

	//fmt.Print(k.counter)
	k.counter++
	//fmt.Print(" | ", k.counter)
	if k.counter == 128+1 {
		//I thought it was supposed to be 128 + 1, but it was givining me a longer input by 1 then the actual answer, I started guessing
		//put 130 and it passes all tests, why I'm not sure but wont question it. Maybe something related to the indexes of syscalls?
		//we need to put number 128 plus 1 here, due to our algorithm
		// it will count exactly 128 instructions without syscalls.

		//notify the kernel "\nTimer fired!\n"
		activateTrapHandler(c, timer)
		//fmt.Print(" | ")

		k.counter = 0

		return true, nil //let's skip this instruction
		//for now and go to the timer handler, then return to this instruction again
		//and execute it normally

	}

	return false, nil
}

/*
confirm kernel state for certain instructions
*/
func assertKernelMode(c *cpu) bool {
	if !c.kernel.kernelMode {
		//go to trap handler illegal instruction
		//fmt.Print(" kmode assrtion failed ")
		activateTrapHandler(c, illegalInstruction)
		return false
	}
	return true
}

/*
privilegedHook is a hook for privileged instructions
where it checks the kernel mode
*/
func privilegedHook(c *cpu, args [3]uint8) (bool, error) {

	//	fmt.Println(" kmode:", c.kernel.kernelMode)
	if assertKernelMode(c) {

		return false, nil
	} else {
		//fmt.Print("privilged caught, shouldb't execute but")
		return true, nil
	}
}

/*
checks memory bounds
*/
func checkBounds(n word) bool {
	return (n < 1024 || n >= 2048)
}

// Initialize kernel support.
//
// (In Go, any function named `init` automatically runs before `main`.)
func init() {

	// This is an example of adding a hook to an instruction.
	// You probably don't actually want to add a hook to the `add` instruction.
	// Really? :) And how about preventing illegal oob jumps by adding to r7???
	//
	instrAdd.addHook(func(c *cpu, args [3]uint8) (bool, error) {

		if c.kernel.kernelMode {
			return false, nil
		}

		a0 := resolveArg(c, args[0])
		a1 := resolveArg(c, args[1])
		if a0 == a1 {
			// Adding a number to itself? That seems like a weird thing to
			// do. Best just to skip it...
			//Oh, yeah return true, nil
		}

		if args[2] == 7 {
			// This instruction is trying to write to the instruction
			// pointer. That sounds dangerous!

			//return false, fmt.Errorf("You're not allowed to ever change the instruction pointer. No loops for you!")

			// Indeed
			if checkBounds(a0 + a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}

		}

		return false, nil
	})

	// TODO: Add hooks to other existing instructions to implement kernel
	// support.

	// Ok since you've started with "add" here, we now check for other
	// illegal jumping instrcutions, a big headache because many instructions
	// can modify iptr.
	//

	/*
		andhook is a hook for the and instruction because it can modify the iptr in ways we dont want it to
	*/
	instrAnd.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 | a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction div does not change iptr in ways we dont want it to
	*/
	instrDiv.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 / a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction mul does not change iptr in ways we dont want it to
	*/
	instrMul.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 * a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction not does not change iptr in ways we dont want it to
	*/
	instrNot.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			if checkBounds(^a0) {

				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction or does not change iptr in ways we dont want it to
	*/
	instrOr.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 | a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction shift left does not change iptr in ways we dont want it to
	*/
	instrShl.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 << a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction shift right does not change iptr in ways we dont want it to
	*/
	instrShr.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 >> a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction sub does not change iptr in ways we dont want it to
	*/
	instrSub.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 - a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})
	/*
		so intruction xor does not change iptr in ways we dont want it to
	*/
	instrXor.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode {
			return false, nil
		}

		if args[2] == 7 {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			if checkBounds(a0 ^ a1) {
				activateTrapHandler(c, oobMemory)
				return true, nil
			}
		}
		return false, nil

	})

	/*
		so intruction loadliteral does not change iptr in ways we dont want it to
	*/
	instrLoadLiteral.addHook(func(c *cpu, args [3]uint8) (bool, error) {
		if c.kernel.kernelMode || args[1] != 7 {
			return false, nil
		}

		//checking the bounds of the momery
		a0 := resolveArg(c, args[0])
		if checkBounds(a0) {
			// a0 is out of the range [1024, 2048)
			activateTrapHandler(c, oobMemory)
			return true, nil

		}
		return false, nil

	})

	/*
		so intruction move does not change iptr in ways we dont want it to
	*/
	instrMove.addHook(func(c *cpu, args [3]uint8) (bool, error) {

		if c.kernel.kernelMode || args[1] != 7 {
			return false, nil
		}

		//a0 := resolveArg(c, args[0])
		//a1 := resolveArg(c, args[1])
		//checking the bounds of the momery
		a0 := resolveArg(c, args[0])
		if checkBounds(a0) {
			// a0 is out of the range [1024, 2048)
			activateTrapHandler(c, oobMemory)
			return true, nil

		}

		return false, nil

	})

	/*
		so intruction cmove does not change iptr in ways we dont want it to
	*/
	instrCmove.addHook(func(c *cpu, args [3]uint8) (bool, error) {

		if c.kernel.kernelMode || args[2] != 7 {
			return false, nil
		}

		//a0 := resolveArg(c, args[0])
		//

		//checking the bounds of the momery
		a0 := resolveArg(c, args[0])
		a1 := resolveArg(c, args[1])
		if a0 != 0 && checkBounds(a1) {
			// a1 is out of the range [1024, 2048)
			activateTrapHandler(c, oobMemory)
			return true, nil

		}
		return false, nil

	})

	//ok so here we need to hook the privileged instructions to disallow them from usermode
	instrRead.addHook(privilegedHook)
	instrWrite.addHook(privilegedHook)
	instrHalt.addHook(privilegedHook)
	instrUnreachable.addHook(privilegedHook)

	//we check for illegal memory access
	instrLoad.addHook(func(c *cpu, args [3]uint8) (bool, error) {

		if c.kernel.kernelMode { //if kernel mode no problem
			return false, nil

		}

		//checking the bounds of the momery
		a0 := resolveArg(c, args[0])
		if checkBounds(a0) {
			// a0 is out of the range [1024, 2048)
			activateTrapHandler(c, oobMemory)
			return true, nil

		}

		return false, nil
	})

	//we check for illegal memory access
	instrStore.addHook(func(c *cpu, args [3]uint8) (bool, error) {

		if c.kernel.kernelMode {
			return false, nil

		}
		a0 := resolveArg(c, args[1])

		if checkBounds(a0) {
			// a0 is out of the range [1024, 2048)
			activateTrapHandler(c, oobMemory)
			return true, nil

		}

		return false, nil
	})

	var (
		// syscall <code>
		//
		// Executes a syscall. The first argument is a literal which identifies
		// what kernel functionality is requested:
		// - 0/read:  Read a byte from the input device and store it in the
		//            lowest byte of r6 (and set the other bytes of r6 to 0)
		// - 1/write: Write the lowest byte of r6 to the output device
		// - 2/exit:  The program exits; print "Program has exited" and halt the
		// 	 		  machine.
		//
		// You may add new syscall codes if you want, but you may not modify
		// these existing codes, as `prime.asm` assumes that they are supported.
		instrSyscall = &instr{
			name: "syscall",

			cb: func(c *cpu, args [3]byte) error {
				// TODO: Fill this in.
				a0 := resolveArg(c, args[0])

				//if debugflag {
				//fmt.Println(" syscall handler:", 1)
				//}

				activateTrapHandler(c, syscall)

				//write the syscall  number to the kernel variable
				c.memory[c.kernel.syscallNumber] = a0

				c.kernel.counter-- //decrement the counter because syscall not counted

				return nil
			},
			validate: nil,
		}

		// TODO: Add other instructions that can be used to implement a kernel.

		//usermode instruction to switch to user mode and pass
		//control to a usermode pointer
		instrUsermode = &instr{
			name: "usermode",
			cb: func(c *cpu, args [3]byte) error {
				// TODO: Fill this in.
				//	fmt.Println("usermode handler")

				if !assertKernelMode(c) {
					return nil
				}

				c.kernel.kernelMode = false

				//pass control here to the usermode address here because...
				c.registers[7] = c.registers[0]

				//seems like r0 will have to be restored by the cpu because
				// we simply don't have the capability do it in the kernel

				c.registers[0] = c.kernel.r0

				return nil
			},
			validate: nil,
		}

		//LGDT instruction (named in implemintation of real Intel processors),
		//to pass the addresses of the trap handler and syscall number
		//and the return address and other variables to this emulator.
		instrLgdt = &instr{
			name: "lgdt",
			cb: func(c *cpu, args [3]byte) error {
				// TODO: Fill this in

				//if usermode - the end
				if !assertKernelMode(c) {
					return nil
				}

				c.kernel.trapHandler = c.registers[0]
				c.kernel.syscallNumber = c.registers[1]
				c.kernel.trapReason = c.registers[2]
				c.kernel.returnAddress = c.registers[3]

				return nil
			},
			validate: nil,
		}
	)

	// Add kernel instructions to the instruction set.
	instructionSet.add(instrSyscall)
	instructionSet.add(instrUsermode)
	instructionSet.add(instrLgdt)

}
