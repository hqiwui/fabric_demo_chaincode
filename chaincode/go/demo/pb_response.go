package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
)

type PbResponse struct {
	Code    string       `json:"code"`
	Data    interface{}  `json:"data"`
	Error   string       `json:"error"`
}

func SuccessPbResponse(data []byte) pb.Response {
	var err error
	response := PbResponse{RESP_CODE_SUCESS,"",""}
	if data != nil {
		err = json.Unmarshal(data, &response.Data)
		if err != nil {
			return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
		}
	}

	responseJSONasbytes, err := StructToJSONBytes(response)
	if err != nil {
		return ErrorPbResponse(RESP_CODE_SYSTEM_ERROR, err.Error())
	}
	return shim.Success(responseJSONasbytes)
}

func ErrorPbResponse(errCode string, errMsg string) pb.Response {
	LogMessage("errCode[" + errCode + "] ErrorInfo:" + errMsg)
	response := PbResponse{errCode, nil, errMsg }
	responseJSONasbytes, err := StructToJSONBytes(response)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(responseJSONasbytes)
}



