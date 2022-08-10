package db

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Session struct {
	ID         ID      `json:"id" db:"id" doc:"Session ID"`
	UserID     ID      `json:"user_id,omitempty" db:"user_id"`
	User       *User   `json:"user,omitempty" db:"-" doc:"User record associated with this session"`
	StartTime  SqlTime `json:"start_time" db:"start_time"`
	ExpiryTime SqlTime `json:"expiry_time" db:"expiry_type"`
}

func NewSession(user User) (Session, error) {
	if _, err := db.Exec("DELETE FROM `sessions` WHERE `user_id`=?", user.ID); err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("DELETE session(user_id:%s): %+v", user.ID, err)
		}
	}

	id := ID(uuid.New().String())
	stt := time.Now()
	exp := stt.Add(time.Minute * 5)
	if _, err := db.Exec("INSERT INTO `sessions` SET `id`=?,`user_id`=?,start_time=?,expiry_time=?",
		id,
		user.ID,
		SqlTime(stt),
		SqlTime(exp),
	); err != nil {
		log.Errorf("failed to create session: %+v", err)
		return Session{}, errors.Errorc(http.StatusUnauthorized, "failed to create session")
	}

	//remove private info from user that will be returned to the app
	user.PwdHash = nil
	user.Tpw = nil
	user.TpwExp = nil
	return Session{
		ID:         id,
		UserID:     "",
		User:       &user,
		StartTime:  SqlTime(stt),
		ExpiryTime: SqlTime(exp),
	}, nil
} //NewSession()

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginRequest) Validate() error {
	if req.Email == "" {
		return errors.Errorf("missing email")
	}
	if req.Password == "" {
		return errors.Errorf("missing password")
	}
	return nil
}

func Login(req LoginRequest) (Session, error) {
	user, err := GetUserByEmail(req.Email)
	if err != nil {
		return Session{}, err
	}
	if user.PwdHash == nil {
		return Session{}, errors.Errorc(http.StatusUnauthorized, "account not yet activated")
	}
	pwdHash := HashPassword(req.Email, req.Password)
	if *user.PwdHash != pwdHash {
		log.Errorf("user(id:%s, email:%s) wrong password(%s)->%s != %s", user.ID, user.Email, req.Password, pwdHash, *user.PwdHash)
		return Session{}, errors.Errorc(http.StatusUnauthorized, "wrong password")
	}
	return NewSession(user)
}

func Logout(sid ID) error {
	result, err := db.Exec("DELETE FROM `sessions` WHERE `id`=?", sid)
	if err != nil {
		log.Errorf("failed to delete session(%s): %+v", sid, err)
		return errors.Errorf("failed to delete session")
	}
	if n, _ := result.RowsAffected(); n != 1 {
		log.Errorf("delete %d for session(%s): %+v", n, sid, err)
		return errors.Errorf("failed to delete session")
	}
	return nil
}

//Read session data and update to extend it
func GetSession(sid ID) (Session, error) {
	//read session and user data at once into this temp structure
	type SessionAndUser struct {
		ID     ID      `db:"id"`
		UserID ID      `db:"user_id"`
		Start  SqlTime `db:"start_time"`
		Expiry SqlTime `db:"expiry_time"`
		Name   string  `db:"name"`
		Email  string  `db:"email"`
		Phone  string  `db:"phone"`
	}
	var su SessionAndUser
	if err := db.Get(&su,
		"SELECT s.`id`,s.`user_id`,s.`start_time`,s.`expiry_time`,u.`name`,u.`email`,u.`phone`"+
			" FROM `sessions` AS s"+
			" JOIN `users` AS u ON u.id=s.user_id"+
			" WHERE s.`id`=?",
		sid,
	); err != nil {
		if err == sql.ErrNoRows {
			return Session{}, errors.Errorf("unknown session(%s)", sid)
		}
		log.Errorf("failed to get session(id:%s): %+v", sid, err)
		return Session{}, errors.Errorf("failed to get session(%s)", sid)
	}

	if time.Time(su.Expiry).Before(time.Now()) {
		DelSession(sid)
		return Session{}, errors.Errorf("session expired")
	}

	//extend the session
	exp := time.Now().Add(time.Minute * 5)
	result, err := db.Exec("UPDATE `sessions` SET `expiry_time`=? WHERE `id`=?",
		SqlTime(exp),
		sid)
	if err != nil {
		log.Errorf("failed to extend session(id:%s): %+v", sid, err)
		return Session{}, errors.Errorf("failed to extend session")
	}
	if n, _ := result.RowsAffected(); n != 1 {
		log.Errorf("failed to extend session(id:%s) affected %d rows", sid, n)
		return Session{}, errors.Errorf("failed to extend session")
	}

	//define the session that we return
	return Session{
		ID: sid,
		User: &User{
			ID:    su.UserID,
			Name:  su.Name,
			Email: su.Email,
			Phone: su.Phone,
		},
		StartTime:  SqlTime(su.Start),
		ExpiryTime: SqlTime(exp),
	}, nil
} //GetSession()

func DelSession(sid ID) error {
	if _, err := db.Exec("DELETE FROM `sessions` WHERE `id`=?", sid); err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("failed to delete session(sid:%s): %+v", sid, err)
			return errors.Wrapf(err, "failed to delete session(sid:%s): %+v", sid, err)
		}
	}
	return nil
} //DelSession()
