#start fabric network:

    docker-compose -f docker-compose-cli.yaml up -d

#install chaincode

    docker exec -it cli /bin/bash
    bash ./scripts/script.sh

#demo chaincode info

    type UserInfo struct {
	    //docType is used to distinguish the various types of objects in state database
	    DocType         string `json:"docType"`         //user_info
	    UserEmail       string `json:"userEmail"`       //邮箱
	    UserNickname    string `json:"userNickname"`    //昵称
	    UserPwdHash     string `json:"userPwdHash"`     //密码hash值
	    UserStatus      string `json:"userStatus"`      //当前状态：00-init 99-作废
    }

#function and gars example

    function: InitUserInfo              args: "testuser@test.com","testuser","111112222233333"
    function: ReadUserInfo              args: "testuser@test.com"
    function: ChangeUserInfo            args: "testuser@test.com","testuser001","111112222233333"
    function: DeleteUserInfo            args: "testuser@test.com"
    function: QueryUserInfoByStatus     args: "00"
    function: GetHistoryForUserInfo     args: "testuser@test.com"
