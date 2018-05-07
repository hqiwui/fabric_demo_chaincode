package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
	"time"
)

func QueryDocsByIdxkey(stub shim.ChaincodeStubInterface, docType string, idxKey string, idxKeyvalue string) ([]byte, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"%s\":\"%s\"}}", docType, idxKey, idxKeyvalue)
	return GetQueryResultForQueryString(stub, queryString)
}

func QueryDocsByIdxkeys(stub shim.ChaincodeStubInterface, docType string, idxKey []string, idxKeyvalue []string) ([]byte, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\"", docType)
	i := 0
	for i < len(idxKey) {
		queryString = fmt.Sprintf("%s,\"%s\":\"%s\"", queryString, idxKey[i], idxKeyvalue[i])
		i++
	}
	queryString = fmt.Sprintf("%s}}", queryString)

	return GetQueryResultForQueryString(stub, queryString)
}

// =========================================================================================
// GetQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func GetQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	LogMessage("- getQueryResultForQueryString queryString:" + queryString )

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer

	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	LogMessage("- getQueryResultForQueryString queryResult:" + buffer.String())

	return buffer.Bytes(), nil
}

func GetOnlyOneDocByIdxkeys(stub shim.ChaincodeStubInterface, docType string, idxKey []string, idxKeyvalue []string) ([]byte, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\"", docType)
	i := 0
	for i < len(idxKey) {
		queryString = fmt.Sprintf("%s,\"%s\":\"%s\"", queryString, idxKey[i], idxKeyvalue[i])
		i++
	}
	queryString = fmt.Sprintf("%s}}", queryString)
	return GetOnlyOneForQueryString(stub, queryString)
}

// =========================================================================================
// GetOnlyOneForQueryString executes the passed in query string.
// Doc is built and returned as a byte array containing the JSON results.
// =========================================================================================
func GetOnlyOneForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	LogMessage("- getQueryResultForQueryString queryString:" + queryString )

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	count := 0
	for resultsIterator.HasNext() {
		count ++
		if count > 1 {
			break
		}
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
	}

	if count == 0 { // not found any doc
		return nil,nil
	}else if count == 1{ //find one doc and return
		LogMessage("- getQueryResultForQueryString queryResult:" + buffer.String())
		return buffer.Bytes(), nil
	}else {  // Not the only doc returned.
		err := errors.New("Not the only doc returned")
		return nil,err
	}
}


func GetHistoryForDocWithNamespace(stub shim.ChaincodeStubInterface, ns string, docKey string) ([]byte, error) {
	if len(docKey) < 1 {
		return nil, errors.New("docKey should not be empty")
	}

	LogMessage("- start getHistoryForDocWithNamespace:" + docKey)

	resultsIterator, err := stub.GetHistoryForKey(ns + docKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the asset_securitized_mngt
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"txId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"dataValue\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON asset_securitized_mngt)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"isDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForDocWithNamespace returning:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func GetHistoryForDoc(stub shim.ChaincodeStubInterface, docKey string) ([]byte, error) {
	return GetHistoryForDocWithNamespace(stub, "", docKey)
}

func GetDocWithNamespace(stub shim.ChaincodeStubInterface, ns string, docKey string) ([]byte, error) {
	valAsbytes, err := stub.GetState(ns + docKey)
	return valAsbytes, err
}

func PutDocWithNamespace(stub shim.ChaincodeStubInterface, ns string, docKey string, bytes []byte) error {
	err := stub.PutState(ns+docKey, bytes)
	return err
}

func GetDoc(stub shim.ChaincodeStubInterface, docKey string) ([]byte, error) {
	return GetDocWithNamespace(stub, "", docKey)
}

func PutDoc(stub shim.ChaincodeStubInterface, docKey string, bytes []byte) error {
	return PutDocWithNamespace(stub, "", docKey, bytes)
}

func CreateCKeyWithNamespace(stub shim.ChaincodeStubInterface, ns string, idxName string, idxPair []string) error {
	compositeKey, err := stub.CreateCompositeKey(ns+idxName, idxPair)
	if err != nil {
		return err
	}
	value := []byte{0x00}
	err = PutDoc(stub, compositeKey, value)
	if err != nil {
		return err
	}

	return nil
}

func CreateCKey(stub shim.ChaincodeStubInterface, idxName string, idxPair []string) error {
	return CreateCKeyWithNamespace(stub, "", idxName, idxPair)
}
