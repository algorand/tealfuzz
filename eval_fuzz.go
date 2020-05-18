package fuzz

import (
	"bytes"
	"fmt"

	aconfig "github.com/algorand/go-algorand/config"
	abasics "github.com/algorand/go-algorand/data/basics"
	atransactions "github.com/algorand/go-algorand/data/transactions"
	alogic "github.com/algorand/go-algorand/data/transactions/logic"
	aprotocol "github.com/algorand/go-algorand/protocol"

	mconfig "github.com/algorand/go-algorand-master/config"
	mbasics "github.com/algorand/go-algorand-master/data/basics"
	mtransactions "github.com/algorand/go-algorand-master/data/transactions"
	mlogic "github.com/algorand/go-algorand-master/data/transactions/logic"
	mprotocol "github.com/algorand/go-algorand-master/protocol"
)

func constructOldParams(program []byte, args [][]byte) mlogic.EvalParams {
	// Construct transaction
	var txn mtransactions.SignedTxn
	txn.Txn.Type = mprotocol.PaymentTx
	txn.Txn.Sender = mbasics.Address{4, 3, 2, 1}
	txn.Lsig.Logic = program
	txn.Lsig.Args = args

	// Set protocol version info
	proto := mconfig.ConsensusParams{
		LogicSigVersion: 5,
		LogicSigMaxCost: 100000,
	}

	// Constuct TxnGroup
	group := []mtransactions.SignedTxn{
		txn,
	}

	// Construct eval params
	return mlogic.EvalParams{
		Txn:        &txn,
		Proto:      &proto,
		TxnGroup:   group,
		GroupIndex: 0,
	}
}

func constructNewParams(program []byte, args [][]byte) alogic.EvalParams {
	// Construct transaction
	var txn atransactions.SignedTxn
	txn.Txn.Type = aprotocol.ApplicationCallTx
	txn.Txn.Sender = abasics.Address{4, 3, 2, 1}
	txn.Lsig.Logic = program
	txn.Lsig.Args = args

	// Set protocol version info
	proto := aconfig.ConsensusParams{
		LogicSigVersion: 5,
		LogicSigMaxCost: 100000,
	}

	// Constuct TxnGroup
	group := []atransactions.SignedTxn{
		txn,
	}

	// Construct eval params
	return alogic.EvalParams{
		Txn:        &txn,
		Proto:      &proto,
		TxnGroup:   group,
		GroupIndex: 0,
	}
}

func disasm(program []byte) (string, error) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic while disassembling program: %v\n", x)
		}
	}()
	return alogic.Disassemble(program)
}

func logFailed(program []byte, args [][]byte) {
	fmt.Printf("crasher bytes: %x\n", program)
	text, err := disasm(program)
	fmt.Printf("crasher disasm (disasm err: %v):\n %s\n", err, text)
	fmt.Printf("crasher args:\n")
	for i, arg := range args {
		fmt.Printf("arg %d: %x\n", i, arg)
	}
}

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

	//fmt.Printf("program: %x\n", program)
	//for i, arg := range args {
	//	fmt.Printf("arg %d: %x\n", i, arg)
	//}

	// Construct eval params for each environment
	ep0 := constructNewParams(program, args)
	ep1 := constructOldParams(program, args)

	// Check program is valid and cost is sufficiently low
	cost0, err0 := alogic.Check(program, ep0)
	cost1, err1 := mlogic.Check(program, ep1)

	// Check for panics
	if pe, ok := err0.(alogic.PanicError); ok {
		logFailed(program, args)
		panic(pe.Error())
	}
	if pe, ok := err1.(mlogic.PanicError); ok {
		logFailed(program, args)
		panic(pe.Error())
	}

	// Check for err nilness equality
	if (err0 == nil) != (err1 == nil) {
		logFailed(program, args)
		panic(fmt.Sprintf("check error nilness not equal! %v != %v", err0, err1))
	}

	// Check for cost equality
	if cost0 != cost1 {
		logFailed(program, args)
		panic(fmt.Sprintf("costs not equal! %d != %d", cost0, cost1))
	}

	// Run the program
	pass0, err0 := alogic.Eval(program, ep0)
	pass1, err1 := mlogic.Eval(program, ep1)

	// Check for panics
	if pe, ok := err0.(alogic.PanicError); ok {
		logFailed(program, args)
		panic(pe.Error())
	}
	if pe, ok := err1.(mlogic.PanicError); ok {
		logFailed(program, args)
		panic(pe.Error())
	}

	// Check for err nilness equality
	if (err0 == nil) != (err1 == nil) {
		logFailed(program, args)
		panic(fmt.Sprintf("eval error nilness not equal! %v != %v", err0, err1))
	}

	// Check for pass equality
	if pass0 != pass1 {
		logFailed(program, args)
		panic(fmt.Sprintf("success not equal! %v != %v", pass0, pass1))
	}

	// If we made it here, the test case is interesting
	return 1
}
