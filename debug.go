package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ************* Interactive debugger *************
//
// You will not need to read most of this file to implement your kernel,
// although you may find it useful to read the implementation to understand how
// the debugger works in more detail.

// Run this CPU in an interactive debugger.
func debug(c *cpu) {
	breakpoints := make(map[word]bool)

	// Step and print the result of the step if the CPU halts.
	//
	// Returns `true` if the CPU has halted.
	step := func() bool {
		halt, err := c.step()
		if err != nil {
			fmt.Printf("CPU entered a bad state: %v; halting...\n", err)
		}
		return halt
	}

	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !s.Scan() {
			break
		}

		line := s.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		cmd := fields[0]
		args := fields[1:]
		switch cmd {
		case "step":
			if c.halted {
				fmt.Println("CPU is halted; cannot step")
				break
			}

			step()

		case "debug":
			usage := "invalid format; expected 'debug [on | off | print [--memory]]'"

			if len(args) == 0 {
				fmt.Println(usage)
				break
			}

			switch {
			case args[0] == "on" && len(args) == 1:
				debugEnabled = true
			case args[0] == "off" && len(args) == 1:
				debugEnabled = false
			case args[0] == "print" && (len(args) == 1 || (len(args) == 2 && args[1] == "--memory")):
				fmt.Printf("CPU state: %v\n", c)
				if len(args) == 2 {
					fmt.Printf("Memory: %v\n", c.memory)
				}
			default:
				fmt.Println(usage)
				break
			}
		case "continue":
			if len(args) != 0 {
				fmt.Println("invalid format: continue takes no arguments")
				break
			}

			if c.halted {
				fmt.Println("CPU is halted; cannot continue")
				break
			}

			for {
				halted := step()
				if halted {
					break
				}
				iptr := c.registers[7]
				if breakpoints[iptr] {
					fmt.Printf("Encountered a breakpoint at instruction pointer %v\n", iptr)
					break
				}
			}
		case "break":
			if len(args) != 2 || (args[0] != "--set" && args[0] != "--clear") {
				fmt.Println("invalid format; expected 'break [--set | --clear] <addr>' (note that <addr> is encoded as hex)")
				break
			}

			addr, err := strconv.ParseUint(args[1], 16, 32)
			if err != nil {
				fmt.Printf("failed to parse address: %v\n", err)
				break
			}

			if args[0] == "--set" {
				breakpoints[word(addr)] = true
			} else {
				delete(breakpoints, word(addr))
			}
		case "appendInput":
			input := strings.Join(args, " ") + "\n"
			c.read = io.MultiReader(c.read, bytes.NewBuffer([]byte(input)))
		case "help":
			if len(args) != 0 {
				fmt.Println("invalid format: help takes no arguments")
				break
			}

			fmt.Println("Available commands:")
			fmt.Println("    step        : Execute a single instruction")
			fmt.Println("    debug       : Enable or disable debugging, or print the current machine state")
			fmt.Println("    continue    : Continue executing instructions until the next breakpoint")
			fmt.Println("    break       : Set or clear a breakpoint")
			fmt.Println("    appendInput : Add text to the input read by the 'read' instruction")
			fmt.Println("    help        : Print this help text")
		default:
			fmt.Println("Unrecognized command")
		}
	}
}
