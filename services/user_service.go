package services

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"wahyuade.com/simple-e-commerce/schemas"
)

const TOPIC_USER = "user-service"
const GET_BY_EMAIL_QUEUE = "getByEmail"
const REGISTER_USER_QUEUE = "registerUser"
const SET_USER_SESSION_QUEUE = "setUserSession"
const GET_USER_BY_SESSION_QUEUE = "getUserBySession"

type UserService struct {
	conn   *amqp091.Connection
	dbConn *sql.DB
}

func (uS UserService) Init(dbConn *sql.DB, conn *amqp091.Connection) ServiceRunner {
	return UserService{
		conn:   conn,
		dbConn: dbConn,
	}
}

func (us UserService) Start() {
	us.RegisterAllHandler()
}

func (uS UserService) RegisterAllHandler() {
	go uS.getByEmailHandler()
	go uS.registerUserHandler()
	go uS.setUserSessionHandler()
	go uS.getUserBySessionHandler()
}

func (uS UserService) Name() string {
	return "user-service"
}

func (uS UserService) getByEmailHandler() {
	log.Printf("%s Handler Registered\n", GET_BY_EMAIL_QUEUE)
	ch := newChannel(uS.conn, TOPIC_USER)
	msgs := listen(ch, TOPIC_USER, GET_BY_EMAIL_QUEUE)
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", GET_BY_EMAIL_QUEUE, d.CorrelationId)
			user := schemas.User{}
			json.Unmarshal(d.Body, &user)
			row := uS.dbConn.QueryRow(`SELECT uuid, email, password, name FROM "user" WHERE email = $1`,
				user.Email,
			)
			err := row.Scan(&user.Uuid, &user.Email, &user.Password, &user.Name)
			resp, _ := json.Marshal(user)
			if err != nil {
				log.Println(err)
			}
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (uS UserService) registerUserHandler() {
	log.Printf("%s Handler Registered\n", REGISTER_USER_QUEUE)
	ch := newChannel(uS.conn, TOPIC_USER)
	msgs := listen(ch, TOPIC_USER, REGISTER_USER_QUEUE)
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", REGISTER_USER_QUEUE, d.CorrelationId)
			user := schemas.User{}
			json.Unmarshal(d.Body, &user)
			uuidV4 := uuid.New().String()
			row := uS.dbConn.QueryRow(`INSERT INTO "user"(uuid, email, password, name) VALUES ($1, $2, $3, $4) RETURNING uuid, email, name`,
				uuidV4,
				user.Email,
				user.Password,
				user.Name,
			)
			err := row.Scan(&user.Uuid, &user.Email, &user.Name)
			resp, _ := json.Marshal(user)
			if err != nil {
				log.Println(err)
			}
			reply(ch, d, resp)
		}
	}()
	<-forever
}

func (uS UserService) setUserSessionHandler() {
	log.Printf("%s Handler Registered\n", SET_USER_SESSION_QUEUE)
	ch := newChannel(uS.conn, TOPIC_USER)
	defer ch.Close()
	msgs := listen(ch, TOPIC_USER, SET_USER_SESSION_QUEUE)
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("Received request for %s with correlation-id: %s", SET_USER_SESSION_QUEUE, d.CorrelationId)
			user := schemas.User{}
			json.Unmarshal(d.Body, &user)
			row := uS.dbConn.QueryRow(`UPDATE "user" SET session = $1 WHERE uuid = $2 RETURNING 'true'`,
				user.Session,
				user.Uuid,
			)
			var isTrue string
			row.Scan(&isTrue)
			reply(ch, d, []byte(isTrue))
		}
	}()
	<-forever
}

func (uS UserService) getUserBySessionHandler() {
	log.Printf("%s Handler Registered\n", GET_USER_BY_SESSION_QUEUE)
	ch := newChannel(uS.conn, TOPIC_USER)
	defer ch.Close()
	msgs := listen(ch, TOPIC_USER, GET_USER_BY_SESSION_QUEUE)
	var forever chan struct{}
	go func() {
		for d := range msgs {
			authorization := strings.Replace(string(d.Body), "Bearer ", "", 1)
			log.Printf("Received request for %s with correlation-id: %s", GET_USER_BY_SESSION_QUEUE, d.CorrelationId)
			row := uS.dbConn.QueryRow(`SELECT uuid, email, name, session FROM "user" WHERE session = $1`,
				authorization,
			)
			user := schemas.User{}
			err := row.Scan(&user.Uuid, &user.Email, &user.Name, &user.Session)
			result, _ := json.Marshal(user)
			if err != nil {
				log.Println(err)
			}
			reply(ch, d, result)
		}
	}()
	<-forever
}
