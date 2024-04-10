# CS386Project4
In this project, you will be provided with a basic CPU emulator. You will extend the CPU emulator with support for instructions that can be used implement a kernel, you will implement a (very simple) kernel using these instructions, and you will attempt to find and exploit vulnerabilities in your classmates' kernels.




Project 4: Kernel

In this project, you will be provided with a basic CPU emulator. You will extend the CPU emulator with support for instructions that can be used implement a kernel, you will implement a (very simple) kernel using these instructions, and you will attempt to find and exploit vulnerabilities in your classmates' kernels.
Architecture

This section describes the CPU architecture and the instructions which are already implemented.
Words

The CPU is a 32-bit, word-oriented CPU. This means that, unlike computers which you've worked with before, each memory address uniquely addresses a single 32-bit word rather than a single byte. If this analogy helps, you can think of this CPU as having 32-bit bytes.
Memory

The memory available to the CPU is a sequence of words which are addressed starting at 0. Memory can be accessed using the load and store instructions, which are described below.
Registers

The CPU has 8 registers which are numbered 0 through 7. Register 7 is also the instruction pointer. Each register stores a single word.

In text, we denote these registers as r0 through r7.
I/O

The CPU is attached to one input device and one output device. Each device is byte-oriented (this makes it easier to interact with the terminal). The instructions read and write read and write a single byte directly from/to the input/output devices.

Since the CPU is word-oriented, we need to convert between 32-bit words and 8-bit bytes in order to interact with the I/O devices. We do this by only considering the least-significant 8 bits of a word and discarding the rest. For example, consider the following instructions:

read r3
write r5

This is the format that you will write your kernel and any other programs in. We provide a basic "assembler" which can read this format and compile it into the equivalent sequence of instruction words. The assembler format is documented below.

read r3 reads one byte from the input device. It overwrites the contents of r3 with a 32-bit word whose lowest-order 8 bits are set to the value of the input, and whose remaining (higher-order) bits are set to zeros.

Similarly, write r5 writes one byte to the output device. The byte that it writes is taken from the lowest-order 8 bits of r5.
Halting

It is possible for the CPU to halt. This can happen either due to an explicit invocation of the halt instruction or because an instruction was executed which gets the CPU into an illegal state. Examples of such instructions are invalid instructions (e.g., instructions with an invalid instruction code), instructions which attempt to read from a memory address which is out-of-bounds of the CPU's memory, instructions which refer to a register which doesn't exist (e.g., r8), etc.
Execution

In each instruction cycle, the CPU executes the following steps in this order:

    Executes the "kernel pre-execute hook" (more on this later), which may cause the CPU to halt
    Reads the value of r7 (let's refer to this value as iptr)
    Loads the word at address iptr in memory, or halts if the address is out-of-bounds
    Decodes this word as an instruction, or halts if decoding fails
    Increments r7
    Executes the instruction, halting if execution fails or if the instruction was an explicit halt instruction

Instruction encoding

Each instruction is encoded as a single word, consisting of one code byte followed by three argument bytes:

MSB                          LSB
v                              v
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
|------||------||------||------|
  code     a0      a1      a2

This diagram reads from most-significant bit (MSB) on the left to least-significant bit (LSB) on the right.

The code identifies which instruction is encoded. The meaning of the arguments a0, a1, and a2 differ by instruction. Also, not all instructions make use of all arguments - for example, the halt instruction takes no arguments, and so when the halt instruction is executed, a0, a1, and a2 are ignored.

Despite arguments being interpreted differently by different instructions, there are some common patterns.

First, arguments often contain the addresses of registers. For example, consider the following instruction:

add r0 r1 r2

This instruction adds the value in r0 to the value in r1 and stores the result in r2.

Second, arguments may also contain literal values (i.e., values which are hard-coded into the instruction itself). This raises a question: How do we differentiate whether an argument contains a register address or a literal value? In most cases (the only exception is the loadLiteral instruction - see below for more details), the most-significant bit of the argument is used to distinguish between register addresses and literal values. For example, consider the following argument values:

MSB  LSB
v      v
00000111
10000111

The first argument - whose MSB is 0 - encodes the address of r7. The second argument - whose MSB is 1 - encodes the literal value 7.

Usually, you will not have to think too hard about how arguments are encoded. The assembler (described below) will take care of interpreting a value like r7 as 00000111 and a value like 7 as 10000111. However, knowing about this encoding may be useful for debugging.
Instruction reference

This section lists the instructions which are provided by default. Part of your assignment will be to add new instructions to this list.
Arithmetic instructions

Most arithmetic instructions are of the form:

<instr> <a0> <a1> <a2>

Each arithmetic instruction treats <a0> and <a1> as either register addresses or literal values. It loads the register or literal values, computes the instruction's arithmetic operation, and stores the result in the register addressed by <a2>.
Instruction 	Operation (using Go syntax)
add 	<a0> + <a1>
sub 	<a0> - <a1>
mul 	<a0> * <a1>
div 	<a0> / <a1>
shl 	<a0> << <a1>
shr 	<a0> >> <a1>
and 	<a0> & <a1>
or 	<a0> | <a1>
xor 	<a0> ^ <a1>

The not instruction is similar, but operates on only one operand:

not <a0> <a1>

It performs the bit-wise negation of the value of the register or literal <a0> and stores the result in the register addressed by <a1>.
Comparison instructions

All comparison instructions are of the form:

<instr> <a0> <a1> <a2>

Each comparison instruction treats <a0> and <a1> as either register addresses or literal values. It loads the register or literal values, computes the instruction's comparison operation, and stores the result in the register addressed by <a2>. If the comparison is true, the value stored in <a2> is 1, and otherwise it is 0.
Instruction 	Operation (using Go syntax)
gt 	<a0> > <a1>
lt 	<a0> < <a1>
eq 	<a0> == <a1>
Other instructions
Move

move <a0> <a1>

Copies the register or literal <a0> into the register <a1>.
Conditional move

cmove <a0> <a1> <a2>

If the register or literal <a0> is non-zero, copies the register or literal <a1> into the register <a2>. Otherwise, does nothing.

This is the core instruction for branching. Executing cmove <a0> <a1> r7 will have the effect of conditionally overwriting r7, which is the instruction pointer, depending on the value of <a0>. See the provided prime.asm for an example of how to use this construct to implement a loop.
Load

load <a0> <a1>

Treats the register or literal <a0> as a memory address. Loads the word at this address and stores it in the register <a1>.
Store

load <a0> <a1>

Treats the register or literal <a1> as a memory address. Stores the register or literal <a0> at this address.
Load literal

loadLiteral <a0> <a1>

loadLiteral is the one instruction which deviates from the normal instruction encoding scheme. A loadLiteral instruction is encoded as follows:

MSB                          LSB
v                              v
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
|------||--------------||------|
  code         a0          a1

loadLiteral stores the 16-bit value a0 in the least-significant 16 bits of the register a1.
Read

read <a0>

Reads a byte from the input device and stores it in the least-significant bits of the register <a0>.

When executing a read instruction, the CPU will block until an input byte is available. If reading from the input device fails, the CPU will halt with an error.
Write

write <a0>

Writes the least-significant 8 bits of the register or literal <a0> to the output device.

When executing a write instruction, the CPU will block until the output byte is written. If writing to the output device fails, the CPU will halt with an error.
Halt

halt

Halts the execution of the CPU.
Debug

debug <a0>

Causes the emulator to print the CPU's current state to stderr. If the register or literal <a0> is non-zero, then the contents of memory are also printed.
Unreachable

unreachable

An instruction that should never be executed, similar to an assertion in a programming language. If unreachable is executed, the CPU immediately halts with an error.
No-op

nop

Do nothing. Note that nop is specifically chosen to have the instruction code 0, which means that any memory regions which are initialized to zero are implicitly filled with nop instructions. This is useful for debugging since it means that, if memory locations are executed erroneously, they are less likely to clobber state that may be important for debugging.
Assembler

We've provided a basic assembler which you can use to write your programs in a human-readable format rather than having to write instructions directly in their binary encoding.

We'll use this simple program to illustrate the assembler's features:

    ; This program prints the numbers 0 through 9.
    
    loadLiteral 0 r0 ; Use r0 as the loop counter
    
loop:
    add r0 '0' r1 ; Store the ASCII encoding of r0 in r1
    write r1      ; Write the ASCII encoding to the output device
    
    lt r0 9 r1  ; Store 'r0 < 10' in r1. Since we've already
                ; written to the output device, we don't need
                ; the contents of r1 anymore, and so we can
                ; re-use it to store the result of this comparison.
    add r0 1 r0 ; Increment the loop counter

    cmove r1 .loop r7 ; If 'r0 < 9' (before incrementing), then we
                      ; need to loop at least one more time. We
                      ; implement this by moving the address of the
                      ; first instruction in the loop to r7, which
                      ; is the instruction pointer.

Comments

Comments are delimited using the ; character.
Literals

Consider this instruction:

lt r0 9 r1

In this instruction, r0 and r1 are registers, while 9 is the literal value 9. Literals are parsed using Go's strconv.ParseUint, and may be written in any format supported by that function (binary, octal, hex, etc).

In most instructions, literals can have values in the range [0, 127] (127=27−1). The one exception is loadLiteral: its literal argument can have values in the range [0, 65535] (65535=216−1).
ASCII

As a special-case, the assembler supports ASCII literals. Consider this instruction:

add r0 '0' r1

In this instruction, r0 and r1 are registers, while '0' is a literal with the value 48 (the ASCII code for the character 0). This is equivalent writing 48.
Labels

It's often important to know the address of a particular instruction. Labels such as loop: can be used. The assembler will automatically compute the offset of the instruction following the label. Elsewhere in the program, the label can be used as a literal argument to other instructions (e.g., .loop).

Note that the assembler is only able to compute the label's offset relative to the program it's assembling. If the program is loaded at a location in memory other than address 0, then the label values cannot be jumped to directly, but must instead be offset by the appropriate amount. For example, consider this snippet from prime.asm, which uses this technique:

; Store 1024 + .after_loop in r2
loadLiteral 1024 r2
add r2 .after_loop r2

; If r4 != 0, break out of the loop
cmove r4 r2 r7

Note that this has been lightly edited from how it is written in prime.asm.

As explained below, prime.asm expects to be loaded at offset 1024, and so it must add 1024 to all labels before jumping.

Style note: The recommended style is to indent all instructions and comments, and to leave labels as the only un-indented items. This aids in readability. That said, the assembler is agnostic to indentation; using other indentation will not affect the behavior of your programs.
Debugger

We've provided a basic debugger which supports step-by-step execution and breakpoints. To run the CPU emulator in the debugger, use the --debug flag (e.g., go run *.go --debug bootloader.asm prime_embedded.asm).

To see more about the functionality offered by the debugger, launch the debugger and then run the help command.
Assignment 1: Bootloader

Due date: Wednesday, April 17th at 11:59pm

Your first assignment is to implement a very simple bootloader. This will help you gain familiarity with programming in this assembly language, and it will form the basis of the kernel that you will hand in for Assignment 2.

We have provided a number of .go files that implement the CPU emulator. You can run them like so:

go run *.go bootloader.asm prime_embedded.asm

The emulator takes two arguments: a bootloader and a program.

First, the emulator allocates 2,048 words of memory and initializes them to zero. It assembles the bootloader, and copies the resulting instruction sequence into memory starting at word 0. The bootloader's instruction sequence may be at most 1,024 words long.

Next, the emulator assembles the program, whose instruction sequence may also be at most 1,024 words long. The emulator fills the input device's buffer with two objects:

    First, it calculates the program's length (in words) as an unsigned 16-bit integer. It stores this integer in big-endian byte order as the first two bytes of the input buffer.
    Second, it converts each of the program's words to a 4-byte value (also in big-endian byte order). It adds these bytes to the input buffer.

Finally, the emulator begins executing, initializing all registers (including the instruction pointer, r7) to 0.

Your bootloader's job is to:

    Read the program from the input device and store it in memory starting at address 1,024
    Execute the program by jumping to address 1,024

We provide an example program, prime_embedded.asm, that you can use to test your bootloader.

Assignment 1 should not require modifying the emulator (i.e., the .go files) in any way. Our auto-grader will ignore any .go files that you submit.
Interacting with the Autograder

To interact with the autograder for this assignment, submit your implementation of bootloader.asm to Gradescope. The autograder for this assignment will use your bootloader to run a set of testing programs (including prime_embedded.asm) and check that their output is as expected.
Assignment 2: Kernel

Due date: Wednesday, May 1st at 11:59pm

More details coming soon…
Assignment 3: Hacking

Due date: TBD

More details coming soon…
