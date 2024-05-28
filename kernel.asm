;=================================
;Our asm kernel
;=================================

 
   ; write '~'

;================
;pass the necessary pointers to the cpu
;=================
    loadLiteral  .trap_handler r0

    loadLiteral  .syscall_number r1
    loadLiteral  .trap_reason r2
    loadLiteral  .return_address r3
    lgdt    

   

;===============================
;load the program into memory
;===============================

; Initialize the register for base address of the program
    loadLiteral 1024 r1

; Read the size of the program in words (assuming two bytes read for size)
    read r0  ; Read first byte (high byte)
    shl r0 8 r0  ; Shift left by 8 bits to make room for the next byte
    read r2  ; Read second byte (low byte)
    or r0 r2 r0  ; Combine the two bytes to form the size

; Initialize the counter for the loop
    move r0 r3  ; Copy the size into r3 to use as a loop counter

; Reading the program into memory starting at address 1024
load_program:
    ; Each loop iteration reads one word
    read r2  ; Read first byte of the word
    shl r2 8 r2
    read r4
    or r2 r4 r2
    shl r2 8 r2
    read r4
    or r2 r4 r2
    shl r2 8 r2
    read r4
    or r2 r4 r2

    store r2 r1  ; Store the word into the memory at address in r1
    add r1 1 r1  ; Increment the memory address
    sub r3 1 r3  ; Decrement the loop counter

    ; Check if the loop should continue
    gt r3 0 r4   ; Check if we still have words to read
    cmove r4 .load_program r7  ; If yes, continue loop

;==============================
;initialize timer counter
;==============================
    store   .timer_counter 0

;==============================
;switch to user mode and pass control to the loaded program
;==============================
 
    loadLiteral    1024 r0
    usermode
   ; loadLiteral  r7

    unreachable  ; Optionally, end bootloader like this for safety

;=================================================================
;data section
;====================

;messages
;"\nProgram has exited\n"
;"\nTimer fired!\n"
;"\nIllegal instruction!\n"
;"\nOut of bounds memory access!\n"
;"\nTimer fired XXXXXXXX times\n"

;If the program attempts to access memory which is not in the address range [1024, 2048),
;the kernel must print \nOut of bounds memory access!\nTimer fired XXXXXXXX times\n to the
;output device and halt the CPU.

;variables
syscall_number:
    nop
    
trap_reason:
    nop
    
return_address:
    nop
    
timer_counter:
    nop

;storage for kernel state with registers
registers:      ;r6 thru r1
    nop
    nop
    nop
    nop
    nop
    nop
    nop
   


;===========================
;kernel mode handlers
;=================================================================

trap_handler:
    ;write   'T'
    ;write   10

;save registers
;seems like r0 will have to be stored by the cpu because we simply can't do it here
;
    loadLiteral .registers  r0
    store   r6 r0  ; Store the word into the memory at address in r0
    add     r0 1 r0  ; Increment the memory address
    store   r5 r0  ; Store the word into the memory at address in r0
    add     r0 1 r0  ; Increment the memory address

    store   r4 r0  ; Store the word into the memory at address in r0
    add     r0 1 r0  ; Increment the memory address

    store   r3 r0  ; Store the word into the memory at address in r0
    add     r0 1 r0  ; Increment the memory address

    store   r2 r0  ; Store the word into the memory at address in r0
    add     r0 1 r0  ; Increment the memory address

    store   r1 r0  ; Store the word into the memory at address in r0


;determine the cause and jump to the appropriate 
    load    .trap_reason r0
    eq      r0 0 r1 
    cmove   r1 .syscall_handler r7
   
    eq      r0 1 r2 
     
    loadLiteral .timer_handler r1
    cmove   r2 r1 r7  ;timer_handler must be here
     
    eq      r0 2 r2 

    loadLiteral .memory_trap_handler r1
    cmove   r2 r1 r7
     
    eq      r0 3 r2 

    loadLiteral  .illegal_instr_handler r1
      
    cmove   r2 r1 r7

    unreachable


trap_handler_wrap:

;restore registers
    loadLiteral .registers  r0
    load    r0 r6  ; Store the word into the memory at address in r1
    add     r0 1 r0  ; Increment the memory address

trap_handler_wrap_nor6:
    load    r0 r5  ; Store the word into the memory at address in r1
    add     r0 1 r0  ; Increment the memory address
    load    r0 r4  ; Store the word into the memory at address in r1
    add     r0 1 r0  ; Increment the memory address
    load    r0 r3  ; Store the word into the memory at address in r1
    add     r0 1 r0  ; Increment the memory address
    load    r0 r2  ; Store the word into the memory at address in r1
    add     r0 1 r0  ; Increment the memory address
    load    r0 r1  ; Store the word into the memory at address in r1
   
    load    .return_address r0
    usermode

    unreachable

;
syscall_handler:
;
;check the syscallNumber
   
    load    .syscall_number r0
 
    ;debug
    ;halt
    eq      r0 0 r1
    cmove   r1 .read_handler r7
    eq      r0 1 r1

    cmove   r1 .write_handler r7
    eq      r0 2 r1

    cmove   r1 .exit_handler r7
    unreachable

;// - 0/read:  Read a byte from the input device and store it in the
;//            lowest byte of r6 (and set the other bytes of r6 to 0)
;// - 1/write: Write the lowest byte of r6 to the output device
;// - 2/exit:  The program exits; print "Program has exited" and halt the
;// 	 		  machine.
read_handler:
    move    0 r6
    read    r6

    ;a solution to avoid restoring r6 since it obviously is needed for return value here
    loadLiteral .registers  r0
    add     r0 1 r0

    loadLiteral .trap_handler_wrap_nor6 r7
    unreachable

write_handler:

    write   r6
    loadLiteral .trap_handler_wrap r7
    unreachable

exit_handler:
    ; "\nProgram has exited\n"
    write   10
    write 'P'
    write 'r'
    write 'o'
    write 'g'
    write 'r'
    write 'a'
    
    write 'm'
    write 32
    write 'h'
    write 'a'
    write 's'
    write 32
    write 'e'
    write 'x'
    write 'i'
    write 't'
    write 'e'
    write 'd'
   ; write 10  ; Newline character

    ; ";\"\nTimer fired!\n\""
print_timer_counter:
     write 10

    write 'T'
    write 'i'
    write 'm'
    write 'e'
    write 'r'
    write 32
    write 'f'
    write 'i'
    write 'r'
    write 'e'
    write 'd'
    write 32
  
  ;print how many times timer fired
  ;it's difficult
    load    .timer_counter r0
  
    shr     r0 28 r1
    and     r1 15 r1

    move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1

    shr     r0 24 r1
    and     r1 15 r1
    move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1

    shr     r0 20 r1
    and     r1 15 r1
    move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1

    shr     r0 16 r1
    and     r1 15 r1
      move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1

    shr     r0 12 r1
    and     r1 15 r1
       move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1

    shr     r0 8 r1
    and     r1 15 r1
       move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1

    shr     r0 4 r1
    and     r1 15 r1
       move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1

    and     r0 15 r1

    ;debug

    move    '0' r3
    gt      r1 9 r2
    cmove   r2 87 r3
    add     r1 r3 r1
    write   r1
   

    write 32
    write 't'
    write 'i'
    write 'm'
    write 'e'
    write 's'
    write 10  ; Newline
    halt


;timer handler
timer_handler:
  ;loadLiteral .hop_over r7
    ; ";\"\nTimer fired!\n\""
     write 10
    write 'T'
    write 'i'
    write 'm'
    write 'e'
    write 'r'
    write 32
    write 'f'
    write 'i'
    write 'r'
    write 'e'
    write 'd'
    write '!'
    write 10  ; Newline
  
hop_over:
    ;increase timer counter
    load    .timer_counter  r0
    add     r0 1 r0
    store   r0 .timer_counter

    loadLiteral .trap_handler_wrap r7
    unreachable


illegal_instr_handler:
    
    ; ";\"\nIllegal instruction!\n\""
    write 10
    write 'I'
    write 'l'
    write 'l'
    write 'e'
    write 'g'
    write 'a'
    write 'l'
    write 32
    write 'i'
    write 'n'
    write 's'
    write 't'
    write 'r'
    write 'u'
    write 'c'
    write 't'
    write 'i'
    write 'o'
    write 'n'
    write '!'
    ;write 10
  
    loadLiteral .print_timer_counter r7

    unreachable


memory_trap_handler:
  
    ; ";\"\nOut of bounds memory access!\n\""
  
    write 10
    write 'O'
    write 'u'
    write 't'
    write 32
    write 'o'
    write 'f'
    write 32
    write 'b'
    write 'o'
    write 'u'
    write 'n'
    write 'd'
    write 's' 
    write 32
    write 'm'
    write 'e'
    write 'm'
    write 'o'
    write 'r'
    write 'y'
    write 32
    write 'a'
    write 'c'
    write 'c'
    write 'e'
    write 's'
    write 's'
    write '!'
    ;write 10
    
    loadLiteral .print_timer_counter r7
    unreachable
