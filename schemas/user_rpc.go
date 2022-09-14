package schemas

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

const TOPIC_USER = "user-service"

const GET_BY_EMAIL_QUEUE = "getByEmail"
const GET_BY_EMAIL_RSP_QUEUE = "getByEmailRsp"
const REGISTER_USER_QUEUE = "registerUser"
const REGISTER_USER_RSP_QUEUE = "registerUserRsp"
const SET_USER_SESSION_QUEUE = "setUserSession"
const SET_USER_SESSION_RSP_QUEUE = "setUserSessionRsp"
const GET_USER_BY_SESSION_QUEUE = "getUserBySession"
const GET_USER_BY_SESSION_RSP_QUEUE = "getUserBySessionRsp"

type UserRPC interface {
	GetByEmail(ctx context.Context, correlationId, email string) (bool, User)
	Register(ctx context.Context, correlationId string, user User) (bool, User)
	SaveSession(ctx context.Context, correlationId string, user User) bool
	GetBySession(ctx context.Context, correlationId, session string) User
}

type user_rpc struct {
	conn *amqp091.Connection
}

func (u user_rpc) GetByEmail(ctx context.Context, correlationId, email string) (bool, User) {
	user := User{
		Email: email,
	}
	requestPayload, _ := json.Marshal(user)
	ch := newChannel(u.conn, TOPIC_USER)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_USER, GET_BY_EMAIL_QUEUE, GET_BY_EMAIL_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &user)
		return user.Uuid != "", user
	}
	return false, user
}

func (u user_rpc) Register(ctx context.Context, correlationId string, user User) (bool, User) {
	requestPayload, _ := json.Marshal(user)
	ch := newChannel(u.conn, TOPIC_USER)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_USER, REGISTER_USER_QUEUE, REGISTER_USER_RSP_QUEUE, correlationId, requestPayload)
	if response != nil {
		json.Unmarshal(response, &user)
		return user.Uuid != "", user
	}
	return false, user
}

func (u user_rpc) SaveSession(ctx context.Context, correlationId string, user User) bool {
	requestPayload, _ := json.Marshal(user)
	ch := newChannel(u.conn, TOPIC_USER)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_USER, SET_USER_SESSION_QUEUE, SET_USER_SESSION_RSP_QUEUE, correlationId, requestPayload)
	return response != nil && string(response) == "true"
}

func (u user_rpc) GetBySession(ctx context.Context, correlationId string, session string) User {
	ch := newChannel(u.conn, TOPIC_USER)
	defer ch.Close()
	response := waitForResponse(ctx, ch, TOPIC_USER, GET_USER_BY_SESSION_QUEUE, GET_USER_BY_SESSION_RSP_QUEUE, correlationId, []byte(session))
	user := User{}
	json.Unmarshal(response, &user)
	return user
}
