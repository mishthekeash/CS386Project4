package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// ************* Instructions *************
//
// You should read all of the code from here down when working on Assignment 2:
// Kernel. You will need to write code similar to the code in this file in order
// to add your own instructions, or to add hooks to existing instructions.

// The type of a function used to implement a CPU instruction.
type instrCallback func(c *cpu, args [3]byte) error

// The type of a function used to validate a decoded instruction.
type instrValidate func(args [3]byte) error

// A description of a CPU instruction.
//
// Note that an `instr` describes an instruction, not an *instance* of an
// instruction. A program is a sequence of *instances* of instructions.
type instr struct {
	// The name of the instruction, used by the assembler and for
	// pretty-printing.
	name string
	// The function to call to execute an instance of this instruction.
	//
	// `cb` may assume that its arguments have already been validated with
	// `validate`.
	cb instrCallback
	// The function to call to validate that a decoded instruction is valid.
	//
	// If `validate` is set to `nil`, no validation is performed.
	validate instrValidate
}

// Adds a hook which is called when instances of `i` are executed.
//
// This hook is run before the normal callback, `i.cb`, is run. It returns a
// `bool` which indicates whether to skip further instruction execution (ie,
// skip executing any other installed hooks and skip `i.cb`). It also returns an
// error. If this error is non-nil, it causes the CPU to halt.
//
// This is intended for use by kernel support, which may wish to treat some
// instructions or some invocations of some instructions as privileged.
//
// Multiple hooks may be added, in which case they are run in LIFO order (ie,
// the most-recently-added hook is run first). During execution, if any hook
// returns `true` or an error, all subsequent hooks are skipped.
func (i *instr) addHook(hook func(c *cpu, args [3]byte) (bool, error)) {
	cb := i.cb
	i.cb = func(c *cpu, args [3]byte) error {
		skip, err := hook(c, args)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
		return cb(c, args)
	}
}

// A decoded instance of an instruction.
type decodedInstr struct {
	// The instruction of which this is an instance.
	def *instr
	// The arguments to the instruction.
	args [3]byte
}

func (i *decodedInstr) String() string {
	return fmt.Sprintf("%v %v", i.def.name, i.args)
}

func (i *decodedInstr) Run(c *cpu) error {
	return i.def.cb(c, i.args)
}

// An instruction set.
type instrSet struct {
	// The instructions in this instruction set.
	//
	// Each instruction lives at the same index that is used to code for that
	// instruction. For example, the 0th instruction in this slice is assigned
	// instruction code 0.
	instructions []*instr
}

// Adds new instructions to the set.
func (s *instrSet) add(i ...*instr) {
	s.instructions = append(s.instructions, i...)
}

// Decodes a word as an instruction.
//
// This resolves the instruction code to a particular instruction, and then
// validates the instruction arguments.
func (s *instrSet) decode(w word) (decodedInstr, error) {
	b := w.toBeBytes()
	code := int(b[0])
	args := [3]byte{b[1], b[2], b[3]}

	if code >= len(s.instructions) {
		return decodedInstr{}, fmt.Errorf("invalid instruction code: %v", code)
	}

	def := s.instructions[code]
	if def.validate != nil {
		if err := def.validate(args); err != nil {
			return decodedInstr{}, fmt.Errorf("invalid arguments for instruction %v: %v", def.name, err)
		}
	}

	return decodedInstr{
		def:  def,
		args: args,
	}, nil
}

// Uses `instr.addHook` to add `hook` as a hook to all instructions in this set.
func (i *instrSet) addHookToAll(hook func(c *cpu, args [3]byte) (bool, error)) {
	for _, instr := range i.instructions {
		instr.addHook(hook)
	}
}

// The type of validation to perform on an argument to an instruction.
type argValidation uint8

const (
	// This argument is ignored; do not validate it.
	ignore argValidation = iota
	// This argument is treated as a register address.
	reg
	// This argument is treated as a register address or a literal.
	regOrLit
)

// Generates the `validate` function for an instruction.
//
// Each argument can be validated as `ignore`, `reg`, or `regOrLit`. `ignore`
// arguments are not validated, `reg` arguments are validated as being a
// register address (i.e., in the range [0, 7]), and `regOrLit` arguments are
// validated as being either a register address or a literal (i.e., in the range
// [0, 7] or in the range [128, 255]).
func genValidate(v0, v1, v2 argValidation) func([3]byte) error {
	validations := []argValidation{v0, v1, v2}
	return func(args [3]byte) error {
		for i, a := range args {
			err := ""
			switch validations[i] {
			case ignore:
			case reg:
				if a >= 128 {
					err = fmt.Sprintf("literal given (%v) where register expected", a-128)
				} else if a >= 8 {
					err = fmt.Sprintf("register out of bounds: %v", a)
				}
			case regOrLit:
				if a >= 8 && a < 128 {
					err = fmt.Sprintf("register out of bounds: %v", a)
				}
			default:
				panic(fmt.Errorf("invalid argValidation: %v", validations[i]))
			}
			if err != "" {
				return fmt.Errorf("argument %v: %v", i, err)
			}
		}
		return nil
	}
}

// Resolves an argument to a value.
//
// If the argument's high bit is set (i.e., if `arg >= 128`), then the remaining
// bits are treated as a literal value, and this literal value is returned.
// Otherwise, the argument is treated as a register address, and the value of
// the addressed register is returned.
func resolveArg(c *cpu, arg byte) word {
	if arg >= 128 {
		return word(arg - 128)
	}
	return c.registers[int(arg)]
}

// Generates the definition of an arithmetic instruction.
//
// The generated instruction has the signature `name <a0> <a1> <a2>`, where <ar0>
// and <a1> are each either a register address or a literal, and <a2> is a
// register address. Execution stores the value `op(<a0>, <a1>)` in <a2>.
func genArithmeticInstr(name string, op func(a, b word) word) *instr {
	return &instr{
		name: name,
		cb: func(c *cpu, args [3]byte) error {
			a0 := resolveArg(c, args[0])
			a1 := resolveArg(c, args[1])
			c.registers[int(args[2])] = op(a0, a1)
			return nil
		},
		validate: genValidate(regOrLit, regOrLit, reg),
	}
}

// Generates the definition of a comparison instruction.
//
// The generated instruction has the signature `name <a0> <a1> <a2>`, where <r0>
// and <a1> are each either a register address or a literal, and <a2> is a
// register address. Execution stores 1 in <a2> if `op(<a1>, <a2>) == true` and
// stores 0 in <a2> otherwise.
func genCmpInstr(name string, op func(a, b word) bool) *instr {
	return genArithmeticInstr(name, func(a, b word) word {
		if op(a, b) {
			return 1
		} else {
			return 0
		}
	})
}

var (
	// nop
	//
	// No-op; do nothing.
	instrNop = &instr{
		name: "nop",
		cb: func(_ *cpu, _ [3]byte) error {
			return nil
		},
		validate: nil,
	}

	// add <a0> <a1> <a2>
	//
	// Stores the value <a0> + <a1> in the register <a2>.
	instrAdd = genArithmeticInstr("add", func(a, b word) word { return a + b })

	// sub <a0> <a1> <a2>
	//
	// Stores the value <a0> - <a1> in the register <a2>.
	instrSub = genArithmeticInstr("sub", func(a, b word) word { return a - b })

	// mul <a0> <a1> <a2>
	//
	// Stores the value <a0> * <a1> in the register <a2>.
	instrMul = genArithmeticInstr("mul", func(a, b word) word { return a * b })

	// div <a0> <a1> <a2>
	//
	// Stores the value <a0> / <a1> in the register <a2>.
	instrDiv = genArithmeticInstr("div", func(a, b word) word { return a / b })

	// shl <a0> <a1> <a2>
	//
	// Stores the value <a0> << <a1> in the register <a2>.
	instrShl = genArithmeticInstr("shl", func(a, b word) word { return a << b })

	// shr <a0> <a1> <a2>
	//
	// Stores the value <a0> >> <a1> in the register <a2>.
	instrShr = genArithmeticInstr("shr", func(a, b word) word { return a >> b })

	// and <a0> <a1> <a2>
	//
	// Stores the value <a0> & <a1> in the register <a2>.
	instrAnd = genArithmeticInstr("and", func(a, b word) word { return a & b })

	// or <a0> <a1> <a2>
	//
	// Stores the value <a0> | <a1> in the register <a2>.
	instrOr = genArithmeticInstr("or", func(a, b word) word { return a | b })

	// xor <a0> <a1> <a2>
	//
	// Stores the value <a0> ^ <a1> in the register <a2>.
	instrXor = genArithmeticInstr("xor", func(a, b word) word { return a ^ b })

	// not <a0> <a1>
	//
	// Stores the value ^ <a0> in the register <a1>.
	instrNot = &instr{
		name: "not",
		cb: func(c *cpu, args [3]byte) error {
			a0 := resolveArg(c, args[0])
			c.registers[int(args[1])] = ^a0
			return nil
		},
		validate: genValidate(regOrLit, reg, ignore),
	}

	// gt <a0> <a1> <a2>
	//
	// Stores the value <a0> > <a1> in the register <a2>.
	instrGt = genCmpInstr("gt", func(a, b word) bool { return a > b })

	// lt <a0> <a1> <a2>
	//
	// Stores the value <a0> < <a1> in the register <a2>.
	instrLt = genCmpInstr("lt", func(a, b word) bool { return a < b })

	// eq <a0> <a1> <a2>
	//
	// Stores the value <a0> == <a1> in the register <a2>.
	instrEq = genCmpInstr("eq", func(a, b word) bool { return a == b })

	// move <a0> <a1>
	//
	// Copies the value <a0> into the register <a1>.
	instrMove = &instr{
		name: "move",
		cb: func(c *cpu, args [3]byte) error {
			c.registers[int(args[1])] = resolveArg(c, args[0])
			return nil
		},
		validate: genValidate(regOrLit, reg, ignore),
	}

	// cmove <a0> <a1> <a2>
	//
	// If <a0> is non-zero, copies the value <a1> into the register <a2>.
	// Otherwise, does nothing.
	instrCmove = &instr{
		name: "cmove",
		cb: func(c *cpu, args [3]byte) error {
			a0 := resolveArg(c, args[0])
			if a0 != 0 {
				c.registers[int(args[2])] = resolveArg(c, args[1])
			}
			return nil
		},
		validate: genValidate(regOrLit, regOrLit, reg),
	}

	// load <a0> <a1>
	//
	// Loads the value at address <r0> and stores it in the register <r1>.
	instrLoad = &instr{
		name: "load",
		cb: func(c *cpu, args [3]byte) error {
			r := resolveArg(c, args[0])
			addr := int(r)

			if addr >= len(c.memory) {
				return fmt.Errorf("load: address out of bounds: %v", r)
			}

			val := c.memory[addr]
			c.registers[int(args[1])] = val

			return nil
		},
		validate: genValidate(regOrLit, reg, ignore),
	}

	// store <a0> <a1>
	//
	// Stores the value <a0> at address <a1>.
	instrStore = &instr{
		name: "store",
		cb: func(c *cpu, args [3]byte) error {
			val := resolveArg(c, args[0])
			r := resolveArg(c, args[1])
			addr := int(r)

			if addr >= len(c.memory) {
				return fmt.Errorf("store: address out of bounds: %v", r)
			}
			c.memory[int(addr)] = val

			return nil
		},
		validate: genValidate(regOrLit, regOrLit, ignore),
	}

	// loadLiteral <val> <a0>
	//
	// Loads the 16-bit literal value <val> into the least-significant 16 bits
	// of the register <a0>. Note that all other bytes of <a0> are overwritten
	// with zeros.
	instrLoadLiteral = &instr{
		name: "loadLiteral",
		cb: func(c *cpu, args [3]byte) error {
			val := beBytesToWord([4]byte{0, 0, args[0], args[1]})
			c.registers[int(args[2])] = val

			return nil
		},
		validate: genValidate(ignore, ignore, reg),
	}

	// read <a0>
	//
	// Reads a byte from the input device and stores it in the least-significant
	// 8 bits of the register <a0>. Note that all other bytes of <a0> are
	// overwritten with zeros.
	instrRead = &instr{
		name: "read",
		cb: func(c *cpu, args [3]byte) error {
			var buf [1]byte
			_, err := c.read.Read(buf[:])
			if err != nil {
				// We treat the output device as infallible to simplify the
				// programming model. If the output device encounters an error,
				// that's equivalent to a hardware bug, and so we just crash the
				// emulator.
				e := fmt.Errorf("the machine has encountered an internal error: read: %v", err)
				if err == io.EOF {
					e = fmt.Errorf("the machine has encountered an internal error: read: EOF (perhaps you were running the debugger and forgot to populate the input buffer using 'appendInput'?)")
				}
				panic(e)
			}

			c.registers[int(args[0])] = word(buf[0])

			return nil
		},
		validate: genValidate(reg, ignore, ignore),
	}

	// write <a0>
	//
	// Writes the least-sigificant 8 bits of <a0> to the output device.
	instrWrite = &instr{
		name: "write",
		cb: func(c *cpu, args [3]byte) error {
			b := [1]byte{byte(resolveArg(c, args[0]))}
			_, err := c.write.Write(b[:])
			if err != nil {
				// We treat the output device as infallible to simplify the
				// programming model. If the output device encounters an error,
				// that's equivalent to a hardware bug, and so we just crash the
				// emulator.
				panic(fmt.Errorf("the machine has encountered an internal error: write: %v", err))
			}

			return nil
		},
		validate: genValidate(regOrLit, ignore, ignore),
	}

	// halt
	//
	// Halts the execution of the machine.
	instrHalt = &instr{
		name: "halt",
		cb: func(c *cpu, _ [3]byte) error {
			c.halted = true
			return nil
		},
		validate: nil,
	}

	// debug <a0>
	//
	// Causes the emulator to print the CPU's current state to stderr. <a0> is
	// treated as a "verbose" flag: if it is non-zero, `debug` also print the
	// contents of memory.
	instrDebug = &instr{
		name: "debug",
		cb: func(c *cpu, args [3]byte) error {
			fmt.Fprintf(os.Stderr, "[debug instruction] CPU state: %v\n", c)
			if resolveArg(c, args[0]) != 0 {
				fmt.Fprintf(os.Stderr, "[debug instruction][verbose] memory: %v\n", c.memory)
			}
			return nil
		},
		validate: genValidate(regOrLit, ignore, ignore),
	}

	// unreachable
	//
	// An instruction that should never be executed, similar to an assertion in
	// a programming language. If this instruction is executed, the CPU
	// immediately halts with an error.
	instrUnreachable = &instr{
		name: "unreachable",
		cb: func(c *cpu, args [3]byte) error {
			return fmt.Errorf("unreachable instruction reached!")
		},
		validate: nil,
	}

	// The set of instructions used by this CPU emulator. This is used by `main`
	// when constructing the `cpu` object. It is also the set of instructions
	// which code in `kernel.go` will add to.
	instructionSet = instrSet{instructions: []*instr{
		instrNop,
		instrAdd,
		instrSub,
		instrMul,
		instrDiv,
		instrShl,
		instrShr,
		instrAnd,
		instrOr,
		instrXor,
		instrNot,
		instrGt,
		instrLt,
		instrEq,
		instrMove,
		instrCmove,
		instrLoad,
		instrStore,
		instrLoadLiteral,
		instrRead,
		instrWrite,
		instrHalt,
		instrDebug,
		instrUnreachable,
	}}
)

// ************* Assembler *************
//
// You don't need to read past here - you will not need to modify this code or
// understand how it works.

// Parses an instruction instance.
//
// See the project handout for a description of the supported format.
func (s *instrSet) parseInstr(labels *map[string]uint32, str string) (word, error) {
	parts := strings.Fields(str)
	if len(parts) == 0 {
		return 0, fmt.Errorf("missing instruction name")
	}

	name := parts[0]

	type argInner struct {
		val uint16
		// true: literal
		// false: register address
		literal bool
	}

	var args []argInner

	// First, do a pass which computes the values of all arguments as
	// `argInner`s, which can store 16-bit integers and can distinguish literal
	// values from register addresses.
	for i, arg := range parts[1:] {
		if strings.HasPrefix(arg, ".") {
			label := arg[1:]
			addr, ok := (*labels)[label]
			if !ok {
				return 0, fmt.Errorf("unrecognized label: %v", label)
			}

			if addr < 0x10000 {
				args = append(args, argInner{val: uint16(addr), literal: true})
			} else {
				return 0, fmt.Errorf("invalid label (\"%v\", value: %v): overflows 16-bit integer", label, addr)
			}
		} else if strings.HasPrefix(arg, "'") {
			if !strings.HasSuffix(arg, "'") || len(arg) != 3 {
				return 0, fmt.Errorf("invalid argument format: %v", arg)
			}
			args = append(args, argInner{val: uint16(arg[1]), literal: true})
		} else {
			literal := !strings.HasPrefix(arg, "r")
			if !literal {
				arg = arg[1:]
			}

			n, err := strconv.ParseUint(arg, 0, 16)
			if err != nil {
				return 0, fmt.Errorf("invalid argument %v (\"%v\"): %v", i, arg, err)
			}

			args = append(args, argInner{val: uint16(n), literal: literal})
		}
	}

	// Second, encode the sequence of arguments in a machine word, performing
	// necessary bounds checks.

	if (name == "loadLiteral" && len(args) != 2) || len(args) > 3 {
		return 0, fmt.Errorf("wrong number of arguments")
	}

	ret := [4]byte{0, 0, 0, 0}

	argToByte := func(a argInner) (byte, error) {
		if a.val > 127 || (!a.literal && a.val > 7) {
			return 0, fmt.Errorf("out of bounds")
		}
		val := a.val
		if a.literal {
			val += 128
		}
		return byte(val), nil
	}

	if name == "loadLiteral" {
		val := args[0].val
		ret[1] = byte(val >> 8)
		ret[2] = byte(val)

		v, err := argToByte(args[1])
		if err != nil {
			return 0, err
		}
		ret[3] = v
	} else {
		for i, a := range args {
			v, err := argToByte(a)
			if err != nil {
				return 0, fmt.Errorf("invalid argument \"%v\": %v", parts[i+1], err)
			}
			ret[i+1] = v
		}
	}

	// Find the instruction with the given name.
	found := false
	for code, i := range s.instructions {
		if i.name == name {
			ret[0] = byte(code)

			var args [3]byte
			copy(args[:], ret[1:])
			var err error
			if i.validate != nil {
				err = i.validate(args)
			}
			if err != nil {
				return 0, fmt.Errorf("validation failed: %v", err)
			}

			found = true
			break
		}
	}

	if !found {
		return 0, fmt.Errorf("no such instruction: %v", name)
	}

	return beBytesToWord(ret), nil
}

// Parses a sequence of instructions.
//
// See the project handout for a description of the supported format.
func (s *instrSet) parseInstrSeq(str string) ([]word, error) {
	stripComments := func(instr string) string {
		parts := strings.Split(instr, ";")
		return strings.TrimSpace(parts[0])
	}

	instrs := strings.Split(str, "\n")

	// First, do a pass to resolve all labels.
	labels := make(map[string]uint32)
	curInstrAddr := uint32(0)
	for i, instr := range instrs {
		instr = stripComments(instr)
		if strings.HasSuffix(instr, ":") {
			label := instr[:len(instr)-1]
			if strings.ContainsFunc(label, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' }) {
				return nil, fmt.Errorf("line %v (\"%v\"): invalid label (\"%v\"): contains illegal characters", i, instr, label)
			}

			if _, ok := labels[label]; ok {
				return nil, fmt.Errorf("line %v (\"%v\"): cannot overwrite existing label %v", i, instr, label)
			}

			labels[label] = curInstrAddr
		} else if instr != "" {
			curInstrAddr++
		}
	}

	// Second, parse the actual instructions now that all labels have been
	// resolved.
	var words []word
	for i, instr := range instrs {
		instr = stripComments(instr)
		if instr == "" {
			continue
		}

		if !strings.HasSuffix(instr, ":") {
			w, err := s.parseInstr(&labels, instr)
			if err != nil {
				return nil, fmt.Errorf("line %v (\"%v\"): %v", i, instr, err)
			}

			words = append(words, w)
		}
	}
	return words, nil
}
