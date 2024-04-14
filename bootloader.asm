; Initialize the register for base address of the program
loadLiteral 1024 r1

; Read the size of the program in words
read r0  ; Read first byte high byte
shl r0 8 r0  ; Shift left by 8 bits to make room for the next byte
read r2  ; Read second byte low byte
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

; Set the instruction pointer to start executing the loaded program
loadLiteral 1024 r7

halt  ; Optionally, end bootloader with a halt for safety
