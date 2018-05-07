#!/bin/bash
# Copyright London Stock Exchange Group All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
echo
echo " ____    _____      _      ____    _____           _____   ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|         | ____| |___ \  | ____|"
echo "\___ \    | |     / _ \   | |_) |   | |    _____  |  _|     __) | |  _|  "
echo " ___) |   | |    / ___ \  |  _ <    | |   |_____| | |___   / __/  | |___ "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|           |_____| |_____| |_____|"
echo

CHANNEL_NAME="$1"
: ${CHANNEL_NAME:="mychannel"}
: ${TIMEOUT:="60"}
COUNTER=1
MAX_RETRY=5
CORE_ORDERER_ADDRESS=orderer.example.com:7050
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

echo "Channel name : "$CHANNEL_NAME

verifyResult () {
	if [ $1 -ne 0 ] ; then
		echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
                echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
		echo
   		exit 1
	fi
}

setGlobals () {

	if [ $1 -eq 0 -o $1 -eq 1 ] ; then
		CORE_PEER_LOCALMSPID="Org1MSP"
		CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
		CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
		if [ $1 -eq 0 ]; then
			CORE_PEER_ADDRESS=peer0.org1.example.com:7051
		else
			CORE_PEER_ADDRESS=peer1.org1.example.com:7051
		fi
	else
		CORE_PEER_LOCALMSPID="Org2MSP"
		CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
		CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
		if [ $1 -eq 2 ]; then
			CORE_PEER_ADDRESS=peer0.org2.example.com:7051
		else
			CORE_PEER_ADDRESS=peer1.org2.example.com:7051
		fi
	fi

	env |grep CORE
}

createChannel() {
	setGlobals 0

    if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer channel create -o $CORE_ORDERER_ADDRESS -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx >&log.txt
	else
		peer channel create -o $CORE_ORDERER_ADDRESS -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel \"$CHANNEL_NAME\" is created successfully ===================== "
	echo
}

updateAnchorPeers() {
    PEER=$1
    setGlobals $PEER

    if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer channel update -o $CORE_ORDERER_ADDRESS -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx >&log.txt
	else
		peer channel update -o $CORE_ORDERER_ADDRESS -c $CHANNEL_NAME -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Anchor peer update failed"
	echo "===================== Anchor peers for org \"$CORE_PEER_LOCALMSPID\" on \"$CHANNEL_NAME\" is updated successfully ===================== "
	sleep 5
	echo
}

## Sometimes Join takes time hence RETRY atleast for 5 times
joinWithRetry () {
	peer channel join -b $CHANNEL_NAME.block  >&log.txt
	res=$?
	cat log.txt
	if [ $res -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
		COUNTER=` expr $COUNTER + 1`
		echo "PEER$1 failed to join the channel, Retry after 2 seconds"
		sleep 2
		joinWithRetry $1
	else
		COUNTER=1
	fi
        verifyResult $res "After $MAX_RETRY attempts, PEER$ch has failed to Join the Channel"
}

joinChannel () {
	for ch in 0 1 2 3; do
		setGlobals $ch
		joinWithRetry $ch
		echo "===================== PEER$ch joined on the channel \"$CHANNEL_NAME\" ===================== "
		sleep 2
		echo
	done
}

installChaincode () {
	PEER=$1
    CHAINCODE=$2
    CHAINCODE_ID=$3
    CHAINCODE_VER=$4
    setGlobals $PEER
	peer chaincode install -n $CHAINCODE_ID -v $CHAINCODE_VER -p github.com/hyperledger/fabric/examples/chaincode/go/$CHAINCODE >&log.txt 
    res=$?
	cat log.txt
        verifyResult $res "Chaincode installation on remote peer PEER$PEER has Failed"
	echo "===================== Chaincode is installed on remote peer PEER$PEER ===================== "
	echo
}

instantiateChaincode () {
	PEER=$1
    CCINIT_ARGS=$2
    CHAINCODE_ID=$3
    CHAINCODE_VER=$4
	setGlobals $PEER
    if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode instantiate -o $CORE_ORDERER_ADDRESS -C $CHANNEL_NAME -n $CHAINCODE_ID -v $CHAINCODE_VER -c $CCINIT_ARGS -P "OR('Org1MSP.member','Org2MSP.member')" >&log.txt
	else
		peer chaincode instantiate -o $CORE_ORDERER_ADDRESS --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CHAINCODE_ID -v $CHAINCODE_VER -c $CCINIT_ARGS -P "OR('Org1MSP.member','Org2MSP.member')" >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Chaincode instantiation on PEER$PEER on channel '$CHANNEL_NAME' failed"
	echo "===================== Chaincode Instantiation on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}

chaincodeQuery () {
  PEER=$1
  CCQUERY_ARGS=$2 
  CHAINCODE_ID=$3
  EXPECTED_RSLT=$4
  echo "===================== Querying on PEER$PEER on channel '$CHANNEL_NAME'... ===================== "
  setGlobals $PEER
  local rc=1
  local starttime=$(date +%s)

  # continue to poll
  # we either get a successful response, or reach TIMEOUT
  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep 3
     echo "Attempting to Query PEER$PEER ...$(($(date +%s)-starttime)) secs"
     peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_ID -c $CCQUERY_ARGS >&log.txt
     test $? -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
     test "$VALUE" = "$EXPECTED_RSLT" && let rc=0
  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on PEER$PEER is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
  fi
}

chaincodeQueryPrintResult () {
  PEER=$1
  CCQUERY_ARGS=$2 
  CHAINCODE_ID=$3
  
  echo "===================== Querying on PEER$PEER on channel '$CHANNEL_NAME'... ===================== "
  setGlobals $PEER
  local rc=1
  local starttime=$(date +%s)

  # continue to poll
  # we either get a successful response, or reach TIMEOUT
  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep 3
     echo "Attempting to Query PEER$PEER ...$(($(date +%s)-starttime)) secs"
     echo "peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_ID -c $CCQUERY_ARGS >&log.txt "
     peer chaincode query -C $CHANNEL_NAME -n $CHAINCODE_ID -c $CCQUERY_ARGS >&log.txt
     test $? -eq 0 && let rc=0
  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on PEER$PEER is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
  fi
}

chaincodeInvoke () {
    PEER=$1
    CCINVOKE_ARGS=$2
    CHAINCODE_ID=$3
    echo "===================== invoking chaincode $3 on PEER$PEER on channel '$CHANNEL_NAME'... ===================== "
    setGlobals $PEER
    if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
		peer chaincode invoke -o $CORE_ORDERER_ADDRESS -C $CHANNEL_NAME -n $CHAINCODE_ID -c "$CCINVOKE_ARGS" >&log.txt
	else
		peer chaincode invoke -o $CORE_ORDERER_ADDRESS  --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n $CHAINCODE_ID -c "$CCINVOKE_ARGS" >&log.txt
	fi
	res=$?
	cat log.txt
	verifyResult $res "Invoke execution on PEER$PEER failed "
	echo "===================== Invoke transaction on PEER$PEER on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}

# Create channel
echo "Creating channel..."
createChannel

# Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel

# Set the anchor peers for each org in the channel
echo "Updating anchor peers for org1..."
updateAnchorPeers 0
echo "Updating anchor peers for org2..."
updateAnchorPeers 2

# Install chaincode on Peer0/Org1 and Peer2/Org2
echo "Installing chaincode on org1/peer0..."
installChaincode 0 chaincode_example02 mycc 1.0
echo "Install chaincode on org2/peer2..."
installChaincode 2 chaincode_example02 mycc 1.0

#Instantiate chaincode on Peer2/Org2
echo "Instantiating chaincode on org2/peer2..."
instantiateChaincode 2 '{"Args":["init","a","100","b","200"]}' mycc 1.0

#Query on chaincode on Peer0/Org1
echo "Querying chaincode on org1/peer0..."
chaincodeQuery 0 '{"Args":["query","a"]}' mycc 100

#Invoke on chaincode on Peer0/Org1
echo "Sending invoke transaction on org1/peer0..."
chaincodeInvoke 0 '{"Args":["invoke","a","b","10"]}' mycc

# Install chaincode on Peer3/Org2
echo "Installing chaincode on org2/peer3..."
installChaincode 3 chaincode_example02 mycc 1.0

#Query on chaincode on Peer3/Org2, check if the result is 90
echo "Querying chaincode on org2/peer3..."
chaincodeQuery 3 '{"Args":["query","a"]}' mycc 90

echo
echo "===================== All GOOD, End-2-End execution completed ===================== "
echo


# Install chaincode on Peer0/Org1 and Peer0/Org2
echo "Installing Mychaincode on org1/peer0..."
installChaincode 0 demo demo 1.0
echo "Install Mychaincode on org2/peer0..."
installChaincode 2 demo demo 1.0

#Instantiate chaincode on Peer2/Org2
echo "Instantiating Mychaincode on org1/peer0..."
instantiateChaincode 2 '{"Args":["Init","200"]}' demo 1.0

echo "Querying Mychaincode ..."
chaincodeQueryPrintResult 0 '{"Args":["Read","selftest"]}' demo

echo "Invoke initUserInfo Mychaincode ..."
chaincodeInvoke 0 '{"Args":["InitUserInfo","testuser@test.com","testuser","111112222233333"]}' demo

echo "query readUserInfo Mychaincode ..."
chaincodeQueryPrintResult 0 '{"Args":["ReadUserInfo","testuser@test.com"]}' demo

echo "Invoke changeUserInfo Mychaincode ..."
chaincodeInvoke 0 '{"Args":["ChangeUserInfo","testuser@test.com","testuser001","111112222233333"]}' demo

echo "query readUserInfo Mychaincode ..."
chaincodeQueryPrintResult 0 '{"Args":["ReadUserInfo","testuser@test.com"]}' demo

echo "Invoke deleteUserInfo Mychaincode ..."
chaincodeInvoke 0 '{"Args":["DeleteUserInfo","testuser@test.com"]}' demo

echo "query queryUserInfoByStatus Mychaincode ..."
chaincodeQueryPrintResult 0 '{"Args":["QueryUserInfoByStatus","00"]}' demo
chaincodeQueryPrintResult 0 '{"Args":["QueryUserInfoByStatus","99"]}' demo

echo "query GetHistoryForUserInfo Mychaincode ..."
chaincodeQueryPrintResult 0 '{"Args":["GetHistoryForUserInfo","testuser@test.com"]}' demo



echo
echo " _____   _   _   ____            _____   ____    _____ "
echo "| ____| | \ | | |  _ \          | ____| |___ \  | ____|"
echo "|  _|   |  \| | | | | |  _____  |  _|     __) | |  _|  "
echo "| |___  | |\  | | |_| | |_____| | |___   / __/  | |___ "
echo "|_____| |_| \_| |____/          |_____| |_____| |_____|"
echo

exit 0
