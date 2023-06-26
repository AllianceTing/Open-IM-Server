package msg

import (
	"context"
	"encoding/json"
	"time"

	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/constant"
	unRelationTb "github.com/OpenIMSDK/Open-IM-Server/pkg/common/db/table/unrelation"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/log"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/tokenverify"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/errs"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/proto/msg"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/proto/sdkws"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/utils"
)

func (m *msgServer) RevokeMsg(ctx context.Context, req *msg.RevokeMsgReq) (*msg.RevokeMsgResp, error) {
	defer log.ZDebug(ctx, "RevokeMsg return line")
	if req.UserID == "" {
		return nil, errs.ErrArgs.Wrap("user_id is empty")
	}
	if req.ConversationID == "" {
		return nil, errs.ErrArgs.Wrap("conversation_id is empty")
	}
	if req.Seq < 0 {
		return nil, errs.ErrArgs.Wrap("seq is invalid")
	}
	if err := tokenverify.CheckAccessV3(ctx, req.UserID); err != nil {
		return nil, err
	}
	user, err := m.User.GetUserInfo(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	_, _, msgs, err := m.MsgDatabase.GetMsgBySeqs(ctx, req.UserID, req.ConversationID, []int64{req.Seq})
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 || msgs[0] == nil {
		return nil, errs.ErrRecordNotFound.Wrap("msg not found")
	}
	// todo: 判断是否已经撤回
	data, _ := json.Marshal(msgs[0])
	log.ZInfo(ctx, "GetMsgBySeqs", "conversationID", req.ConversationID, "seq", req.Seq, "msg", string(data))
	if !tokenverify.IsAppManagerUid(ctx) {
		switch msgs[0].SessionType {
		case constant.SingleChatType:
			if err := tokenverify.CheckAccessV3(ctx, msgs[0].SendID); err != nil {
				return nil, err
			}
		case constant.SuperGroupChatType:
			members, err := m.Group.GetGroupMemberInfoMap(ctx, msgs[0].GroupID, utils.Distinct([]string{req.UserID, msgs[0].SendID}), true)
			if err != nil {
				return nil, err
			}
			if req.UserID != msgs[0].SendID {
				switch members[req.UserID].RoleLevel {
				case constant.GroupOwner:
				case constant.GroupAdmin:
					if members[msgs[0].SendID].RoleLevel != constant.GroupOrdinaryUsers {
						return nil, errs.ErrNoPermission.Wrap("no permission")
					}
				default:
					return nil, errs.ErrNoPermission.Wrap("no permission")
				}
			}
		default:
			return nil, errs.ErrInternalServer.Wrap("msg sessionType not supported")
		}
	}
	err = m.MsgDatabase.RevokeMsg(ctx, req.ConversationID, req.Seq, &unRelationTb.RevokeModel{
		UserID:   req.UserID,
		Nickname: user.Nickname,
		Time:     time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}

	tips := sdkws.RevokeMsgTips{
		RevokerUserID:  req.UserID,
		ClientMsgID:    "",
		RevokeTime:     utils.GetCurrentTimestampByMill(),
		Seq:            req.Seq,
		SesstionType:   msgs[0].SessionType,
		ConversationID: req.ConversationID,
	}
	var recvID string
	if msgs[0].SessionType == constant.SuperGroupChatType {
		recvID = msgs[0].GroupID
	} else {
		recvID = msgs[0].RecvID
	}
	if err := m.notificationSender.NotificationWithSesstionType(ctx, req.UserID, recvID, constant.MsgRevokeNotification, msgs[0].SessionType, &tips); err != nil {
		return nil, err
	}
	return &msg.RevokeMsgResp{}, nil
}
