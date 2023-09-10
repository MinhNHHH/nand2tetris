// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Put your code here.
(START)
@SCREEN
D=A
@0 // to store Screen Address
M=D

(KBCHECK)
	@KBD
	D=M

	@BLACK
	D;JGT	//JUMP IF ANY KBD KEYS ARE PRESSED

	@WHITE
	D;JEQ	//ELSE JUMP TO WHITEN

	@KBDCHECK
	0;JMP

(BLACK)
	@1 // Set color
	M=-1

	@FILL
	0;JMP


(WHITE)
	@1 // Set color
	M=0
	@FILL
	0;JMP



(FILL)
	@1 // Get color
	D=M

	@0
	A=M // Set address to a the pixel
	M=D // Fill color


	@0
	D=M+1	//INC TO NEXT PIXEL
	@KBD
	D=A-D	//KBD-SCREEN=A

	// Inc to next pixel
	@0
	M=M+1
	A=M
	
	@FILL
	D;JGT	//IF A=0 EXIT AS THE WHOLE SCREEN IS BLACK
	@START
	0;JMP