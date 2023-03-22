package logic

import (
	"book/service/user/model"
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"strings"
	"time"

	"book/service/user/api/internal/svc"
	"book/service/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) getJwtToken(secretKey string, iat, seconds, userId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims["userId"] = userId
	token := jwt.New(jwt.SigningMethodHS512)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginReply, err error) {
	if len(strings.TrimSpace(req.Password)) == 0 || len(strings.TrimSpace(req.Username)) == 0 {
		return nil, errors.New("参数错误")
	}

	var userInfo *model.User
	userInfo, err = l.svcCtx.UserModel.FindOneByNumber(l.ctx, req.Username)
	switch err {
	case nil:
		break
	case model.ErrNotFound:
		return nil, errors.New("用户不存在")
	default:
		return nil, err
	}

	if userInfo.Password != req.Password {
		return nil, errors.New("密码错误")
	}

	var token string
	now := time.Now().Unix()
	token, err = l.getJwtToken(l.svcCtx.Config.Auth.AccessSecret, now, l.svcCtx.Config.Auth.AccessExpire, userInfo.Id)
	if err != nil {
		log.Printf("ErrorHere: %s", err)
		return nil, err
	}

	return &types.LoginReply{
		Id:           userInfo.Id,
		Name:         userInfo.Name,
		Gender:       userInfo.Gender,
		AccessToken:  token,
		AccessExpire: now + l.svcCtx.Config.Auth.AccessExpire,
		RefreshAfter: now + l.svcCtx.Config.Auth.AccessExpire - 600,
	}, nil
}
