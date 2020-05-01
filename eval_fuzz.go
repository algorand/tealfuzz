package logic

import (
	"bytes"
	"fmt"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/protocol"
)

type mockAppLedger struct{}

func (ml *mockAppLedger) Balance(addr basics.Address) (basics.MicroAlgos, error) {
	return basics.MicroAlgos{1234}, nil
}

func (ml *mockAppLedger) Round() basics.Round {
	return basics.Round(1234)
}

func (ml *mockAppLedger) AppGlobalState(appIdx basics.AppIndex) (basics.TealKeyValue, error) {
	tkv := make(basics.TealKeyValue)
	tkv["A"] = basics.TealValue{
		Type:  basics.TealBytesType,
		Bytes: "hello",
	}
	tkv["A"] = basics.TealValue{
		Type: basics.TealUintType,
		Uint: 1234,
	}
	return tkv, nil
}

func (ml *mockAppLedger) AppLocalState(addr basics.Address, appIdx basics.AppIndex) (basics.TealKeyValue, error) {
	return ml.AppGlobalState(appIdx)
}

func (ml *mockAppLedger) AssetHolding(addr basics.Address, assetIdx basics.AssetIndex) (holding basics.AssetHolding, err error) {
	return basics.AssetHolding{
		Frozen: true,
		Amount: 1234,
	}, nil
}

func (al *mockAppLedger) AssetParams(addr basics.Address, assetIdx basics.AssetIndex) (params basics.AssetParams, err error) {
	return basics.AssetParams{}, nil
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

	// Construct transaction
	var txn transactions.SignedTxn

	txn.Txn.Type = protocol.ApplicationCallTx
	txn.Txn.ApplicationArgs = args
	txn.Txn.Sender = basics.Address{4, 3, 2, 1}
	txn.Txn.Accounts = []basics.Address{
		basics.Address{1, 2, 3, 4},
	}

	txn.Lsig.Logic = program
	txn.Lsig.Args = args

	// Set protocol version info
	proto := config.ConsensusParams{
		LogicSigVersion: 5,
		LogicSigMaxCost: 100000,
	}

	// Constuct TxnGroup
	group := []transactions.SignedTxn{
		txn,
	}

	// Construct eval params
	ep := EvalParams{
		Txn:        &txn,
		Proto:      &proto,
		TxnGroup:   group,
		GroupIndex: 0,
		Ledger:     &mockAppLedger{},
	}

	fmt.Printf("program: %x\n", program)
	for i, arg := range args {
		fmt.Printf("arg %d: %x\n", i, arg)
	}

	// Check program is valid and cost is sufficiently low
	_, err = CheckStateful(program, ep)

	// Check if err was panic (since we recover)
	if pe, ok := err.(PanicError); ok {
		panic(pe.Error())
	}

	// Run the program
	_, _, err = EvalStateful(program, ep)
	fmt.Printf("err: %v\n", err)

	// Check if err was panic (since we recover)
	if pe, ok := err.(PanicError); ok {
		panic(pe.Error())
	}

	// If we made it here, the test case is interesting
	return 1
}
