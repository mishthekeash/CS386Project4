; Your bootloader's job is to:

; Read the program from the input device and store it in memory starting at address 1,024
; Execute the program by jumping to address 1,024

read r0
loadLiteral 1024 r1

; if r1 =! 0, jump to loop
;read every word and store it in memory starting at address 1024
;when r1 = 0, jump to after loop

loop:
    read r3
    load r3 r1
    add 1 r1
    sub 1 r0
    loadLiteral 0 r4
    add r4 .loop r4

    ;leaves loop if r0 = 0
    cmove r0 r4 r7


;if r0 =0 jump to afer loop

after_loop: 
    loadLiteral 1024 r7
     
  
