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

	type Bank struct {
		BankCode	string	`json:"bankcode"`
		Amount		float64 `json:"amount"`
	}

	type Account struct {
		No          string  `json:"no"`
		Name        string  `json:"name"`
		Balance 		float64 `json:"cashBalance"`
		Banks    		[]Bank 	`json:"banks"`
	}

	const cap float64 = 250000.0

	func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
		fmt.Println("Initializing account keys")
		var blank []string
		blankBytes, _ := json.Marshal(&blank)
		err := stub.PutState("AccountKeys", blankBytes)
		if err != nil {
		    fmt.Println("Failed to initialize account key")
		}
		fmt.Println("Initialization complete")
		return nil, nil
	}

	func createBankDeposits(name string, amount float64) (Bank, error) {
		var bank Bank
		bank.BankCode = name
		bank.Amount = amount
		return bank, nil
	}

	func updateBankDeposits(bank Bank, amount float64) (Bank) {
		bank.Amount += amount
		return bank
	}

	func getBankSplit(banksnaming string, amount float64) ([]Bank, error) {
		var banks []Bank
		banknames := strings.Split(banksnaming, ":")

		if(len(banks) == 0) {
			fmt.Println("Bank names are mising.")
			return nil, errors.New("Bank names are mising.")
		}

		for _, value := range banknames {
			if(amount >= cap) {
				bank, err  := createBankDeposits(value, cap)
				if err != nil {
					fmt.Println("Error createBankDeposits ")
			    return nil, err
				}
				banks = append(banks, bank)
				amount -= cap
			} else if (amount < cap) {
				bank, err  := createBankDeposits(value, amount)
				if err != nil {
					fmt.Println("Error createBankDeposits ")
					return nil, err
				}
				banks = append(banks, bank)
			}
		}

		if(amount > 0){
			updateBankDeposits(banks[0], amount)
		}

		return banks, nil
	}

	func (t *SimpleChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	  if len(args) != 4 {
	      fmt.Println("Error obtaining request details. Missing arguments.")
	      return nil, errors.New("Error obtaining request details. Missing arguments.")
	  }

	  accountno := args[0]
		name := args[1]

		balance, err  := strconv.ParseFloat(args[2], 64)
		if err != nil {
			fmt.Println("Internal error while reading balance from request")
	    return nil, errors.New("Internal error while reading balance from request")
		}

		banks, err := getBankSplit(args[3], balance)
		if err != nil {
			fmt.Println("Internal error while spliting amount")
	    return nil, err
		}
		/*var banks []Bank
		banknames := strings.Split(args[3], ":")
		for _, value := range banknames {
			var bank Bank
			bank.BankCode = value
			bank.Amount = 10
			banks = append(banks, bank)
		}*/

	  var account = Account{No: accountno, Name: name, Balance: balance, Banks: banks}
	  accountBytes, err := json.Marshal(&account)
	  if err != nil {
	      fmt.Println("error creating account" + account.No)
	      return nil, errors.New("Error creating account " + account.No)
	  }

	  fmt.Println("Attempting to get state of any existing account for " + account.No)
	  existingBytes, err := stub.GetState(account.No)
		if err == nil {
	        var userAccount Account
	        err = json.Unmarshal(existingBytes, &userAccount)
	        if err != nil {
	            fmt.Println("Error unmarshalling account " + account.No + "\n--->: " + err.Error())

	            if strings.Contains(err.Error(), "unexpected end") {
	                fmt.Println("No data means existing account found for " + account.No + ", initializing account.")
	                err = stub.PutState(account.No, accountBytes)

	                if err == nil {
	                    fmt.Println("created account" + account.No)
	                    return nil, nil
	                } else {
	                    fmt.Println("failed to create initialize account for " + account.No)
	                    return nil, errors.New("failed to initialize an account for " + account.No + " => " + err.Error())
	                }
	            } else {
	                return nil, errors.New("Error unmarshalling existing account " + account.No)
	            }
	        } else {
	            fmt.Println("Account already exists for " + account.No + " " + userAccount.No)
			    		return nil, errors.New("Can't reinitialize existing user " + account.No)
	        }
	    } else {

	        fmt.Println("No existing account found for " + account.No + ", initializing account.")
	        err = stub.PutState(account.No, accountBytes)

	        if err == nil {
	            fmt.Println("created account" + account.No)
	            return nil, nil
	        } else {
	            fmt.Println("failed to create initialize account for " + account.No)
	            return nil, errors.New("failed to initialize an account for " + account.No + " => " + err.Error())
	        }
	    }
	}

	func (t *SimpleChaincode) depositMoney(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
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

		err = stub.PutState(account.No, updatedAccountBytes)

		if err == nil {
			fmt.Println("deposited money to account" + account.No)
			return nil, nil
		} else {
			fmt.Println("failed to deposit money to account " + account.No)
			return nil, errors.New("failed to deposit money to account " + account.No + " => " + err.Error())
	   	}
		fmt.Println("Successfully completed deposit")
		return nil, nil
	}

	func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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

	func GetBalance(userAccount string, stub shim.ChaincodeStubInterface) (Account, error){
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

	func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
		fmt.Println("run is running " + function)
		return t.Invoke(stub, function, args)
	}

	func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
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
