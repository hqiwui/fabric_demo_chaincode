package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

const (
	ST_COMM_INIT      string = "00"    // init
	ST_COMM_APPROVING string = "01"    // 01-正在审核
	ST_COMM_APPROVED  string = "02"    // 02-审核通过
	ST_COMM_REJECTED  string = "03"    // 03-审核不通过
	ST_COMM_NILED     string = "99"    // 99-作废
)

const (
	RESP_CODE_SUCESS                       string = "1000"   // 成功
	RESP_CODE_ARGUMENTS_ERROR              string = "2000"   // 2000-参数错误
	RESP_CODE_DATA_ALREADY_EXIST           string = "2010"   // 2001-数据已经存在
	RESP_CODE_DATA_NOT_EXISTED             string = "2020"   // 2011-数据不存在
	RESP_CODE_SYSTEM_ERROR                 string = "9999"   // 系统错误
)

type DomoChaincode struct {
	UserMng *UserMng
}

func main() {
	err := shim.Start(new(DomoChaincode))
	if err != nil {
		LogMessage("Error starting Simple chaincode:" + err.Error())
	}
}

/**
 * Init initializes chaincode
 */
func (t *DomoChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	LogMessage("demo chaincode Is Starting Up")
	_, args := stub.GetFunctionAndParameters()
	var Aval int
	var err error

	if len(args) != 1 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "Incorrect number of arguments. Expecting 1")
	}

	// convert numeric string to integer
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "Expecting a numeric string argument to Init()")
	}

	// store compaitible demo application version
	err = stub.PutState("demo_ui", []byte("1.0"))
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}

	// this is a very simple dumb test.  let's write to the ledger and error on any errors
	//making a test var "selftest", its handy to read this right away to test the network
	err = stub.PutState("selftest", []byte(strconv.Itoa(Aval)))
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error()) //self-test fail
	}

	// init modules
	t.UserMng = new(UserMng)

	LogMessage(" - ready for action") //self-test pass
	return SuccessPbResponse(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *DomoChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	LogMessage("invoke is running " + function)

	// Handle different functions
	if function == "Init" { //init the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "Read" { 	//selftest
		return t.Read(stub, args)
	} else if function == "InitUserInfo" { 				//create a new user_info
		return t.UserMng.InitUserInfo(stub, args)
	} else if function == "ReadUserInfo" { 				//read a user_info
		return t.UserMng.ReadUserInfo(stub, args)
	} else if function == "ChangeUserInfo" { 			//changeUserInfo
		return t.UserMng.ChangeUserInfo(stub, args)
	} else if function == "DeleteUserInfo" { 			//delete user_info
		return t.UserMng.DeleteUserinfo(stub, args)
	} else if function == "QueryUserInfoByStatus" { 	//query UserInfo By Status
		return t.UserMng.QueryUserInfoByStatus(stub, args)
	} else if function == "GetHistoryForUserInfo"{
		return t.UserMng.GetHistoryForUserInfo(stub, args)
	}

	LogMessage("invoke did not find func: " + function) //error
	return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "Received unknown function invocation")
}

// ============================================================================================================================
// Read - read a generic variable from ledger
//
// Shows Off GetState() - reading a key/value from the ledger
//
// Inputs - Array of strings
//  0
//  key
//  "abc"
//
// Returns - string
// ============================================================================================================================
func (t *DomoChaincode) Read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key string
	var err error
	LogMessage("starting read")

	if len(args) != 1 {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, "Incorrect number of arguments. Expecting key of the var to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key) //get the var from ledger
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, "Failed to get state for " + key + ":" + err.Error())
	}

	LogMessage("- end read")
	return SuccessPbResponse(valAsbytes) //send it onward
}
