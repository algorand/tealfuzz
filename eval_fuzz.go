package logic

import (
	"bytes"
	"fmt"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
)

func Fuzz(data []byte) int {
	// Ensure input isn't too long
	if len(data) > 50000 {
		return 0
	}

	buf := bytes.NewBuffer(data)

	/*
	 * First, let's read off some bytes as arguments
	 */

	// First byte determines how many arguments
	numArgs, err := buf.ReadByte()
	if err != nil {
		return 0
	}

	// Read in each argument
	var args [][]byte
	for i := 0; i < int(numArgs); i++ {
		// Next two bytes are length of arg
		var argLen uint16
		highOrder, err := buf.ReadByte()
		if err != nil {
			return 0
		}

		lowOrder, err := buf.ReadByte()
		if err != nil {
			return 0
		}

		argLen = (uint16(highOrder) << 8) | uint16(lowOrder)

		// Read in the arg
		arg := buf.Next(int(argLen))
		if len(arg) != int(argLen) {
			// Not enough bytes
			return 0
		}

		// Append it
		args = append(args, arg)
	}

	// Rest of bytes is the program
	program := buf.Bytes()

	// Construct transaction
	var txn transactions.SignedTxn
	txn.Lsig.Logic = program
	txn.Lsig.Args = args

	// Set protocol version info
	proto := config.ConsensusParams{
		LogicSigVersion: EvalMaxVersion - 1,
	}

	// Constuct GroupSenders
	gsnd := make([]basics.BalanceRecord, 4)

	// Construct eval params
	ep := EvalParams{Txn: &txn, Proto: &proto, GroupSenders: gsnd}

	fmt.Printf("program: %x\n", program)
	for i, arg := range args {
		fmt.Printf("arg %d: %x\n", i, arg)
	}

	// Check program is valid and cost is sufficiently low
	_, err = Check(program, ep)

	// Check if err was panic (since we recover)
	if pe, ok := err.(PanicError); ok {
		panic(pe.Error())
	}

	// Run the program
	_, err = Eval(program, ep)

	// Check if err was panic (since we recover)
	if pe, ok := err.(PanicError); ok {
		panic(pe.Error())
	}

	// If we made it here, the test case is interesting
	return 1
}
