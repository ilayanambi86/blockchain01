package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
    "strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type SimpleChaincode struct {
}

type Account struct {
	ID          string  `json:"id"`
	Balance float64 `json:"cashBalance"`
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("Initialization complete")
	return nil, nil
}

func (t *SimpleChaincode) createAccount(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 2 {
        fmt.Println("Error obtaining useraccount & opening balance")
        return nil, errors.New("createAccount accepts useraccount & opening balance as arguments")
    }
    useraccount := args[0]
	openBal, err  := strconv.ParseFloat(args[1], 64)

	if err != nil {
		fmt.Println("Internal Error ")
        return nil, errors.New("Internal Error ")
	}

    var account = Account{ID: useraccount, Balance: openBal}
    accountBytes, err := json.Marshal(&account)
    if err != nil {
        fmt.Println("error creating account" + account.ID)
        return nil, errors.New("Error creating account " + account.ID)
    }

    fmt.Println("Attempting to get state of any existing account for " + account.ID)
    existingBytes, err := stub.GetState(account.ID)
	if err == nil {

        var user Account
        err = json.Unmarshal(existingBytes, &user)
        if err != nil {
            fmt.Println("Error unmarshalling account " + account.ID + "\n--->: " + err.Error())

            if strings.Contains(err.Error(), "unexpected end") {
                fmt.Println("No data means existing account found for " + account.ID + ", initializing account.")
                err = stub.PutState(account.ID, accountBytes)

                if err == nil {
                    fmt.Println("created account" + account.ID)
                    return nil, nil
                } else {
                    fmt.Println("failed to create initialize account for " + account.ID)
                    return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
                }
            } else {
                return nil, errors.New("Error unmarshalling existing account " + account.ID)
            }
        } else {
            fmt.Println("Account already exists for " + account.ID + " " + user.ID)
		    return nil, errors.New("Can't reinitialize existing user " + account.ID)
        }
    } else {

        fmt.Println("No existing account found for " + account.ID + ", initializing account.")
        err = stub.PutState(account.ID, accountBytes)

        if err == nil {
            fmt.Println("created account" + account.ID)
            return nil, nil
        } else {
            fmt.Println("failed to create initialize account for " + account.ID)
            return nil, errors.New("failed to initialize an account for " + account.ID + " => " + err.Error())
        }
    }
}

func (t *SimpleChaincode) depositMoney(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    if len(args) != 2 {
        fmt.Println("Error obtaining useraccount & deposit money")
        return nil, errors.New("depositMoney accepts useraccount & money as arguments")
    }

	useraccount := args[0]
	depositAmt, err  := strconv.ParseFloat(args[1], 64)

	if err != nil {
		fmt.Println("Internal Error ")
		return nil, errors.New("Internal Error ")
	}

	fmt.Println("Getting State on account " + useraccount)
	accountBytes, err := stub.GetState(useraccount)
	if err != nil {
		fmt.Println("Account not found")
		return nil, errors.New("Account not found " + useraccount)
	}

	var account Account
	fmt.Println("Unmarshalling Account " + useraccount)
	err = json.Unmarshal(accountBytes, &account)
	if err != nil {
		fmt.Println("Error unmarshalling account " + useraccount)
		return nil, errors.New("Error unmarshalling account " + useraccount)
	}

	account.Balance += depositAmt

	updatedAccountBytes, err := json.Marshal(&account)
	if err != nil {
		fmt.Println("Error marshalling the account")
		return nil, errors.New("Error marshalling the account")
	}

	err = stub.PutState(account.ID, updatedAccountBytes)

	if err == nil {
		fmt.Println("deposited money to account" + account.ID)
		return nil, nil
	} else {
		fmt.Println("failed to deposit money to account " + account.ID)
		return nil, errors.New("failed to deposit money to account " + account.ID + " => " + err.Error())
   	}
	fmt.Println("Successfully completed deposit")
	return nil, nil
}

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting ......")
	}

	if args[0] == "GetBalance" {
			fmt.Println("Getting particular Balance")
			account, err := GetBalance(args[1], stub)
			if err != nil {
				fmt.Println("Error Getting particular account balance")
				return nil, err
			} else {
				accountBytes, err1 := json.Marshal(&account)
				if err1 != nil {
					fmt.Println("Error marshalling the account")
					return nil, err1
				}
				fmt.Println("All success, returning the account")
				return accountBytes, nil
			}
	}
	return nil, nil
}

func GetBalance(userAccount string, stub *shim.ChaincodeStub) (Account, error){
	var account Account

	accountBytes, err := stub.GetState(userAccount)
	if err != nil {
		fmt.Println("Error retrieving account " + userAccount)
		return account, errors.New("Error retrieving account " + userAccount)
	}

	err = json.Unmarshal(accountBytes, &account)
	if err != nil {
		fmt.Println("Error unmarshalling account " + userAccount)
		return account, errors.New("Error unmarshalling account " + userAccount)
	}

	return account, nil
}

func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	if function == "createAccount" {
        fmt.Println("Firing createAccount")
        return t.createAccount(stub, args)
    } else if function == "depositMoney" {
        fmt.Println("Firing deposit")
        return t.depositMoney(stub, args)
    } else if function == "init" {
        fmt.Println("Firing init")
        return t.Init(stub, "init", args)
    }

	return nil, errors.New("Received unknown function invocation")
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: %s", err)
	}
}

