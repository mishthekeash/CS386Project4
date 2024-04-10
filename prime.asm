; This program asks the user to enter a decimal number and computes whether or
; not it is prime. It assumes that it is run as a user-mode program, and that
; it is loaded at address 1,024 in memory. For a program that can run embedded
; (without a kernel), see `prime_embedded.asm`.

    ; "Enter a number: "
    loadLiteral 'E' r6
    syscall 1
    loadLiteral 'n' r6
    syscall 1
    loadLiteral 't' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral 'r' r6
    syscall 1
    loadLiteral 32  r6; Space character
    syscall 1
    loadLiteral 'a' r6
    syscall 1
    loadLiteral 32  r6; Space character
    syscall 1
    loadLiteral 'n' r6
    syscall 1
    loadLiteral 'u' r6
    syscall 1
    loadLiteral 'm' r6
    syscall 1
    loadLiteral 'b' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral 'r' r6
    syscall 1
    loadLiteral ':' r6
    syscall 1
    loadLiteral 32  r6; Space character
    syscall 1


    ; We accumulate the user's input one character at a time, storing
    ; the intermediate value in r0. Each subsequent digit that we read
    ; has the effect of implying that all preceding digits represent
    ; numbers 10 times larger than we previously thought. For example,
    ; if we read the digit 1, that might just be the number 1 if the
    ; next character is '\n'. However, if we next read the digit 2,
    ; then we've learned that the input might be '12', in which case the
    ; 1 that we read is actually in the 10's place.
    ;
    ; This suggests a natural algorithm: Each time we read a digit
    ; (instead of reading '\n'), multiply our accumulated value by 10,
    ; and then add the new digit, assuming that it is in the 1's place.
    ; This will ensure that, whenever we finally encounter '\n', the
    ; accumulator (r0) will already store the correct, final value.
    loadLiteral 0 r0 ; Clear r0 to use as a counter
loop:
    ; Read the next character into r3
    syscall 0
    move r6 r3

    ; Check to see whether we read '\n'. If so, break out of the
    ; loop.
    eq r3 10 r4           ; r4 = r3 == '\n'
    loadLiteral 1024 r2
    add r2 .after_loop r2 ; Store 1024 + .after_loop in r2
    cmove r4 r2 r7        ; If r3 == '\n', break out of the loop
    
    ; As described above, since we've read a new digit, we know that
    ; all previous digits we read have values 10 times higher than
    ; we originally thought, so we multiply our accumulator (r0) by
    ; 10. However, it's possible that too many digits have been
    ; provided, and the resulting number cannot fit in a 32-bit
    ; register. If this happens, then 'r0 * 10' will overflow.
    ; By storing the result in r1, we can compare r0 with r1 - if
    ; the multiplication overflowed, then r1 will be less than r0.
    mul r0 10 r1 ; r1 = r0 * 10

    ; Check for overflow. If r0 > r1, this means that "r0 * 10"
    ; overflowed.
    gt r0 r1 r2
    loadLiteral 1024 r4
    loadLiteral .overflow r5
    add r4 r5 r4             ; Store 1024 + .overflow in r4
    cmove r2 r4 r7           ; If "r0 * 10" overflowed, jump to 1024 + .overflow

    ; If we've gotten to this point, the counter did not overflow.
    sub r3 48 r3  ; Convert from ASCII (subtract ASCII '0').

    ; Bounds check that the character we read was an ASCII digit.
    ; Since we've subtracted ASCII '0', r3 will be in the range
    ; [0, 9] if we read an ASCII digit and outside of this range
    ; otherwise.
    gt r3 9 r4
    loadLiteral 1024 r5
    add r5 .not_digit r5 ; Store 1024 + .not_digit in r5
    cmove r4 r5 r7       ; If r3 > 9, jump to 1024 + .not_digit

    add r1 r3 r0  ; Add the new digit into the counter.
    
    loadLiteral 1024 r3
    add r3 .loop r3      ; Store 1024 + .loop in r3
    move r3 r7           ; Continue the loop

after_loop:

    ; Now that we've read the number into r0, we can check for its primality.
    ; We do this using a simple algorithm: starting at 2 and working up to
    ; the number itself, check to see whether each subsequent integer divides
    ; evenly into the number.

    loadLiteral 2 r1 ; Use r1 to store the divisor

check_prime_loop:
    ; Check to see whether we've reached the number itself (ie, r0 == r1).
    ; If so, then we found no divisors, and so the number is prime.
    eq r0 r1 r2
    loadLiteral 1024 r3
    add r3 .prime r3    ; Store 1024 + .prime in r3
    cmove r2 r3 r7      ; If r0 == r1 (it's prime), jump to 1024 + .prime

    ; Check to see whether r1 evenly divides r0. Since we're performing integer
    ; division, if r1 does not evenly divide r0, '(r0 / r1) * r1' will not be
    ; equal to r0.
    div r0 r1 r2 ; r2 = r0 / r1
    mul r2 r1 r2 ; r2 = (r0 / r1) * r1
    eq r0 r2 r2  ; r2 = r0 == (r0 / r1) * r1
    loadLiteral 1024 r3
    add r3 .not_prime r3 ; Store 1024 + .not_prime in r3
    cmove r2 r3 r7       ; If it's not prime, jump to 1024 + .not_prime

    ; Increment the divisor (r1) and continue the loop
    add r1 1 r1
    loadLiteral 1024 r3
    add r3 .check_prime_loop r3 ; Store 1024 + .check_prime_loop in r3
    move r3 r7                  ; Jump to 1024 + .check_prime_loop

prime:
    ; "Prime!\n"
    loadLiteral 'P' r6
    syscall 1
    loadLiteral 'r' r6
    syscall 1
    loadLiteral 'i' r6
    syscall 1
    loadLiteral 'm' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral '!' r6
    syscall 1
    loadLiteral 10  r6 ; Newline character
    syscall 1
    syscall 2 ; Exit process

not_prime:
    ; "Not prime!\n"
    loadLiteral 'N' r6
    syscall 1
    loadLiteral 'o' r6
    syscall 1
    loadLiteral 't' r6
    syscall 1
    loadLiteral 32  r6 ; Space character
    syscall 1
    loadLiteral 'p' r6
    syscall 1
    loadLiteral 'r' r6
    syscall 1
    loadLiteral 'i' r6
    syscall 1
    loadLiteral 'm' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral '!' r6
    syscall 1
    loadLiteral 10  r6 ; Newline character
    syscall 1
    syscall 2 ; Exit process

not_digit:
    ; "Read non-digit character!\n"
    loadLiteral 'R' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral 'a' r6
    syscall 1
    loadLiteral 'd' r6
    syscall 1
    loadLiteral 32  r6 ; Space character
    syscall 1
    loadLiteral 'n' r6
    syscall 1
    loadLiteral 'o' r6
    syscall 1
    loadLiteral 'n' r6
    syscall 1
    loadLiteral '-' r6
    syscall 1
    loadLiteral 'd' r6
    syscall 1
    loadLiteral 'i' r6
    syscall 1
    loadLiteral 'g' r6
    syscall 1
    loadLiteral 'i' r6
    syscall 1
    loadLiteral 't' r6
    syscall 1
    loadLiteral 32  r6 ; Space character
    syscall 1
    loadLiteral 'c' r6
    syscall 1
    loadLiteral 'h' r6
    syscall 1
    loadLiteral 'a' r6
    syscall 1
    loadLiteral 'r' r6
    syscall 1
    loadLiteral 'a' r6
    syscall 1
    loadLiteral 'c' r6
    syscall 1
    loadLiteral 't' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral 'r' r6
    syscall 1
    loadLiteral '!' r6
    syscall 1
    loadLiteral 10  r6 ; Newline character
    syscall 1
    syscall 2 ; Exit process

overflow:
    ; "Overflowed!\n"
    loadLiteral 'O' r6
    syscall 1
    loadLiteral 'v' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral 'r' r6
    syscall 1
    loadLiteral 'f' r6
    syscall 1
    loadLiteral 'l' r6
    syscall 1
    loadLiteral 'o' r6
    syscall 1
    loadLiteral 'w' r6
    syscall 1
    loadLiteral 'e' r6
    syscall 1
    loadLiteral 'd' r6
    syscall 1
    loadLiteral '!' r6
    syscall 1
    loadLiteral 10  r6 ; Newline character
    syscall 1
    syscall 2 ; Exit process
