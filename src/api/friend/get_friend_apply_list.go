package friend

import (
	"Open_IM/src/common/config"
	"Open_IM/src/common/log"
	pbFriend "Open_IM/src/proto/friend"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/skiffer-git/grpc-etcdv3/getcdv3"
	"net/http"
	"strings"
)

type paramsGetFriendApplyList struct {
	OperationID string `json:"operationID" binding:"required"`
}
type UserInfo struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Gender     int32  `json:"gender"`
	Mobile     string `json:"mobile"`
	Birth      string `json:"birth"`
	Email      string `json:"email"`
	Ex         string `json:"ex"`
	ReqMessage string `json:"reqMessage"`
	ApplyTime  string `json:"applyTime"`
	Flag       int32  `json:"flag"`
}

func GetFriendApplyList(c *gin.Context) {
	log.Info("", "", "api get_friend_apply_list init ....")

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName)
	client := pbFriend.NewFriendClient(etcdConn)
	defer etcdConn.Close()

	params := paramsGetFriendApplyList{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &pbFriend.GetFriendApplyReq{
		OperationID: params.OperationID,
		Token:       c.Request.Header.Get("token"),
	}
	log.Info(req.Token, req.OperationID, "api get friend apply list  is server")
	RpcResp, err := client.GetFriendApplyList(context.Background(), req)
	if err != nil {
		log.Error(req.Token, req.OperationID, "err=%s,call get friend apply list rpc server failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": "call get friend apply list rpc server failed"})
		return
	}
	log.InfoByArgs("call get friend apply list rpc server success,args=%s", RpcResp.String())
	if RpcResp.ErrorCode == 0 {
		userInfoList := make([]UserInfo, 0)
		for _, applyUserinfo := range RpcResp.Data {
			var un UserInfo
			un.UID = applyUserinfo.Uid
			un.Name = applyUserinfo.Name
			un.Icon = applyUserinfo.Icon
			un.Gender = applyUserinfo.Gender
			un.Mobile = applyUserinfo.Mobile
			un.Birth = applyUserinfo.Birth
			un.Email = applyUserinfo.Email
			un.Ex = applyUserinfo.Ex
			un.Flag = applyUserinfo.Flag
			un.ApplyTime = applyUserinfo.ApplyTime
			un.ReqMessage = applyUserinfo.ReqMessage
			userInfoList = append(userInfoList, un)
		}
		resp := gin.H{"errCode": RpcResp.ErrorCode, "errMsg": RpcResp.ErrorMsg, "data": userInfoList}
		c.JSON(http.StatusOK, resp)
	} else {
		resp := gin.H{"errCode": RpcResp.ErrorCode, "errMsg": RpcResp.ErrorMsg}
		c.JSON(http.StatusOK, resp)
	}
	log.InfoByArgs("api get friend apply list success return,get args=%s,return args=%s", req.String(), RpcResp.String())
}
