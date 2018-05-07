package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	DT_USER_INFO string                 = "userInfo"
	NS_USER_INFO string                 = DT_USER_INFO + "_"
	PK_FD_USER_INFO string              = "userEmail"
	IDX_FD_USER_STATUS string           = "userStatus"
	IDX_UERS_STATUS_2_USER_EMAIL string = IDX_FD_USER_STATUS + "_2_" + PK_FD_USER_INFO
)

type UserMng struct {}

type UserInfo struct {
	//docType is used to distinguish the various types of objects in state database
	DocType         string `json:"docType"`         //user_info
	UserEmail       string `json:"userEmail"`       //邮箱
	UserNickname    string `json:"userNickname"`    //昵称
	UserPwdHash     string `json:"userPwdHash"`     //密码hash值
	UserStatus      string `json:"userStatus"`      //当前状态：00-init 99-作废
}


// ============================================================
// initUserInfo - create a new userInfo, store into chaincode state
// ============================================================
func (t *UserMng) InitUserInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// "user_email",   "user_nickname", "user_pwd_hash"
	if len(args) != 3 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR,"Incorrect number of arguments. Expecting 3")
	}

	// ==== Input sanitation ====
	LogMessage("- start init user")
	if len(args[0]) <= 0 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR,"1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR,"2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR,"3rd argument must be a non-empty string")
	}

	email 		:= args[0]
	nickname 	:= args[1]
	pwdHash 	:= args[2]

	// ==== Check if user_info already exists ====
	userInfoAsBytes, err := GetDocWithNamespace(stub, NS_USER_INFO, email)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, "Failed to get UserInfo: " + err.Error())
	} else if userInfoAsBytes != nil {
		return ErrorPbResponse(RESP_CODE_DATA_ALREADY_EXIST, "This UserInfo already exists: " + email)
	}

	// ==== Create user_info object and marshal to JSON ====
	userInfo := UserInfo{DT_USER_INFO, email, nickname, pwdHash, ST_COMM_INIT}
	userInfoJSONasBytes, err := json.Marshal(userInfo)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}

	// === Save user_info to state ===
	err = PutDocWithNamespace(stub, NS_USER_INFO, email, userInfoJSONasBytes)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}

	//create IDX_uers_status_2_user_email
	err = CreateCKeyWithNamespace(stub, NS_USER_INFO, IDX_UERS_STATUS_2_USER_EMAIL, []string{userInfo.UserStatus,userInfo.UserEmail})
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}
	
	// ==== user_info saved and indexed. Return success ====
	LogMessage("- end init user_info")
	return SuccessPbResponse(nil)
}

// ===============================================
// readUserInfo - read a user_info from chaincode state
// ===============================================
func (t *UserMng) ReadUserInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var email string
	var err error

	if len(args) != 1 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR,"Incorrect number of arguments. Expecting UserEmail")
	}

	email = args[0]
	valAsbytes, err := GetDocWithNamespace(stub, NS_USER_INFO, email) //get the user_info from chaincode state
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	} else if valAsbytes == nil {
		return ErrorPbResponse(RESP_CODE_DATA_NOT_EXISTED, "UserInfo does not exist: " + email )
	}

	LogMessage("- end read UserInfo")
	return SuccessPbResponse(valAsbytes)
}


// ==================================================
// delete - delete a user_info key/value pair from state
// ==================================================
func (t *UserMng) DeleteUserinfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// "UserEmail"
	if len(args) != 1 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	if len(args[0]) <= 0 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "1st argument must be a non-empty string")
	}

	email := args[0]

	LogMessage("- start DeleteUserinfo: UserEmail " + email )

	ValAsbytes, err := GetDocWithNamespace(stub, NS_USER_INFO, email) //get the UserInfo from chaincode state
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, "Failed to get doc for " + NS_USER_INFO + email + ":" + err.Error())
	} else if ValAsbytes == nil {
		return ErrorPbResponse(RESP_CODE_DATA_NOT_EXISTED, "user_info does not exist: " + email + ":" + err.Error())
	}

	userInfoToUpdate := UserInfo{}
	err = json.Unmarshal(ValAsbytes, &userInfoToUpdate)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}

	if userInfoToUpdate.UserStatus == ST_COMM_NILED {
		LogMessage("- end delete user_info (success) " +  email + "s UserInfo was already deleted!")
	}else {
		userInfoToUpdate.UserStatus = ST_COMM_NILED

		userInfoJSONasBytes, err := json.Marshal(userInfoToUpdate)
		if err != nil {
			return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
		}

		err = PutDocWithNamespace(stub, NS_USER_INFO, email, userInfoJSONasBytes)
		if err != nil {
			return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
		}
	}
	LogMessage("- end DeleteUserinfo (success)")
	return SuccessPbResponse(nil)
}

// ==================================================
// Change UserInfo key/value pair from state
// ==================================================
func (t *UserMng) ChangeUserInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// "UserEmail",   "user_nickname", "user_pwd_hash"
	if len(args) != 3 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "Incorrect number of arguments. Expecting 3")
	}

	// ==== Input sanitation ====
	if len(args[0]) <= 0 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "3rd argument must be a non-empty string")
	}

	email := args[0]
	nickname := args[1]
	pwdHash := args[2]

	LogMessage("- start ChangeUserInfo: UserEmail " + email + " , UserNickname " + nickname + " , UserPwdHash " + pwdHash)
		
	ValAsbytes, err := GetDocWithNamespace(stub, NS_USER_INFO, email) //get the UserInfo from chaincode state
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, "Failed to get doc for " + NS_USER_INFO + email + ":" + err.Error())
	} else if ValAsbytes == nil {
		return ErrorPbResponse(RESP_CODE_DATA_NOT_EXISTED, "user_info does not exist: " + email)
	}
	
	userInfoToUpdate := UserInfo{}
	err = json.Unmarshal(ValAsbytes, &userInfoToUpdate)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}

	var isChanged bool
	isChanged = false
	
	if userInfoToUpdate.UserNickname != nickname {
		userInfoToUpdate.UserNickname = nickname
		isChanged = true
	}
	if userInfoToUpdate.UserPwdHash != pwdHash {
		userInfoToUpdate.UserPwdHash = pwdHash
		isChanged = true
	}
	
	if !isChanged {
		LogMessage("- end changeUserInfo (no change no commit)")
		return SuccessPbResponse(nil)
	}
	
	userInfoJSONasBytes, err := json.Marshal(userInfoToUpdate)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}
		
	err = PutDocWithNamespace(stub, NS_USER_INFO, email, userInfoJSONasBytes)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}

	LogMessage("- end ChangeUserInfo (success)")
	return SuccessPbResponse(nil)
}

// ===============================================
// queryUserInfoByStatus - read a user_info from chaincode state
// ===============================================
func (t *UserMng) QueryUserInfoByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0
	// "status_01"
	if len(args) != 1 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR, "Incorrect number of arguments. Expecting owner_status")
	}

	userStatus := args[0]

	queryResults, err := QueryDocsByIdxkey(stub, DT_USER_INFO, IDX_FD_USER_STATUS, userStatus)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}
	return SuccessPbResponse(queryResults)
}

func (t *UserMng) GetHistoryForUserInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return ErrorPbResponse(RESP_CODE_ARGUMENTS_ERROR,"Incorrect number of arguments. Expecting userEmail" )
	}

	email := args[0]
	LogMessage("- start getHistoryForAssetOwner: " + email)

	historyUserInfoBytes, err :=  GetHistoryForDocWithNamespace(stub, NS_USER_INFO, email)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR,err.Error())
	}

	return SuccessPbResponse(historyUserInfoBytes)
}

