package db

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type User struct {
	ID      ID       `json:"id"`
	Name    string   `json:"name" doc:"User name and surname used when members contact the donar, or when user is represents a user as a member."`
	Phone   string   `json:"phone" doc:"Phone must have 0 + 9 digits"`
	Email   string   `json:"email" doc:"Email address"`
	PwdHash *string  `json:"pwd_hash,omitempty" db:"pwd_hash,omitempty"`
	Tpw     *string  `json:"tpw,omitempty" db:"tpw,omitempty"`
	TpwExp  *SqlTime `json:"tpw_exp,omitempty" db:"tpw_exp,omitempty"`
}

const phonePattern = "0[0-9]{9}"

var phoneRegex = regexp.MustCompile("^" + phonePattern + "$")

func (u *User) Validate() error {
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return errors.Errorf("invalid name \"%s\"", u.Name)
	}
	var err error
	if u.Phone, err = nationalPhone(u.Phone); err != nil {
		return err
	}
	u.Email = strings.TrimSpace(u.Email) //not pattern checked... will send message to it to verify
	if u.Email == "" {
		return errors.Errorf("missing email")
	}
	return nil
}

func nationalPhone(s string) (string, error) {
	s = strings.TrimSpace(s)
	err := errors.Errorf("phone \"%s\" must start with 27 or 0 followed by 9 digits", s)
	s = strings.TrimPrefix(s, "+")
	switch len(s) {
	case 11: //27 82 123 4567
		if s[0:2] != "27" {
			return "", errors.Wrapf(err, "11 digits not starting with 27")
		}
		s = "0" + s //change to national format
	case 10: //0 82 123 4567
		if s[0:1] != "0" {
			return "", errors.Wrapf(err, "10 digits not starting with 0")
		}
	default:
		return "", errors.Wrapf(err, "neither 10 nor 11 digits")
	}
	if !phoneRegex.MatchString(s) {
		return "", errors.Wrapf(err, "not only digits")
	}
	return s, nil
} //nationalPhone()

func AddUser(newUser User) (User, error) {
	//if user exists with either phone or email, cannot add this
	//and validate already solved the phone format
	if err := newUser.Validate(); err != nil {
		return User{}, errors.Errorc(http.StatusBadRequest, err.Error())
	}
	if _, err := GetUserByEmail(newUser.Email); err == nil {
		return User{}, errors.Errorc(http.StatusConflict, fmt.Sprintf("%s is already registered", newUser.Email))
	}
	if _, err := GetUserByPhone(newUser.Phone); err == nil {
		return User{}, errors.Errorc(http.StatusConflict, fmt.Sprintf("%s is already registered", newUser.Phone))
	}

	newUser.ID = ID(uuid.New().String())
	{
		tpw := uuid.New().String()
		newUser.Tpw = &tpw
		tpwExp := SqlTime(time.Now().Add(time.Hour * 24))
		newUser.TpwExp = &tpwExp
	}
	_, err := db.Exec(
		"INSERT INTO `users` SET id=?,name=?,phone=?,email=?,tpw=?,tpw_exp=?,pwd_hash=null",
		newUser.ID,
		newUser.Name,
		newUser.Phone,
		newUser.Email,
		newUser.Tpw,
		newUser.TpwExp,
	)
	if err != nil {
		return User{}, errors.Wrapf(err, "failed to insert user")
	}
	return newUser, nil
}

type ActivateRequest struct {
	Tpw string `json:"tpw" doc:"Temporary password"`
	Pwd string `json:"pwd" doc:"New password selected by the user"`
}

func (req ActivateRequest) Validate() error {
	if req.Tpw == "" {
		return errors.Errorf("missing tpw")
	}
	if req.Pwd == "" {
		return errors.Errorf("missing pwd")
	}
	return nil
}

func Activate(req ActivateRequest) (Session, error) {
	if err := req.Validate(); err != nil {
		return Session{}, errors.Errorc(http.StatusBadRequest, "invalid activation request")
	}
	var user User
	if err := db.Get(&user, "SELECT `id`,`name`,`email`,`phone`,`tpw_exp` FROM `users` WHERE `tpw`=?", req.Tpw); err != nil {
		log.Errorf("failed to get user(tpw:%s) for activation: %+v", req.Tpw, err)
		return Session{}, errors.Errorc(http.StatusNotFound, "failed to activate")
	}
	if user.TpwExp == nil {
		return Session{}, errors.Errorc(http.StatusInternalServerError, "tpw_exp not set") //found tpw, so tpw_exp must be set!
	}
	if time.Time(*user.TpwExp).Before(time.Now()) {
		log.Errorf("user(tpw:%s,id:%s) tpw expired at %s", req.Tpw, user.ID, *user.TpwExp)
		return Session{}, errors.Errorc(http.StatusUnauthorized, "activation link expired")
	}

	passwordHash := HashPassword(user.Email, req.Pwd)
	if _, err := db.Exec(
		"UPDATE `users` SET `tpw`=null,`tpw_exp`=null,`pwd_hash`=?",
		passwordHash,
	); err != nil {
		log.Errorf("failed to set password: %+v", err)
		return Session{}, errors.Errorc(http.StatusInternalServerError, "failed to set password")
	}

	user.Tpw = nil
	user.TpwExp = nil
	user.PwdHash = nil //do not reveal to session or outside

	//create a new session for this users to auto-login upon account activation
	return NewSession(user)
} //Activate()

type ResetRequest struct {
	Email string `json:"email"`
}

func (req ResetRequest) Validate() error {
	if req.Email == "" {
		return errors.Errorf("missing email")
	}
	return nil
}

func Reset(req ResetRequest) (User, error) {
	user, err := GetUserByEmail(req.Email)
	if err != nil {
		return User{}, err
	}
	{
		tpw := uuid.New().String()
		user.Tpw = &tpw
		tpwExp := SqlTime(time.Now().Add(time.Hour * 24))
		user.TpwExp = &tpwExp
	}
	if _, err := db.Exec(
		"UPDATE `users` SET tpw=?,tpw_exp=? WHERE id=?",
		user.Tpw,
		user.TpwExp,
		user.ID,
	); err != nil {
		log.Errorf("failed to reset user(%+v): %+v", user, err)
		return User{}, errors.Errorf("failed to reset user")
	}
	return user, nil
} //Reset()

func FindUsers(filter string, limit int) ([]User, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var rows []User
	if err := db.Select(&rows,
		"SELECT `id`,`name`,`phone` FROM `users` WHERE name like ? OR phone like ? LIMIT ?",
		"%"+filter+"%",
		"%"+filter+"%",
		limit); err != nil {
		return nil, errors.Wrapf(err, "failed to search users")
	}
	return rows, nil
}

func GetUser(id ID) (User, error) {
	var u User
	if err := db.Get(&u,
		"SELECT `id`,`name`,`phone`,`email`,`pwd_hash`,`tpw`,`tpw_exp` FROM `users` WHERE id=?",
		id,
	); err != nil {
		return User{}, errors.Wrapf(err, "failed to get user(id=%s)", id)
	}
	return u, nil
} //GetUser()

func GetUserByTpw(tpw ID) (User, error) {
	var u User
	if err := db.Get(&u,
		"SELECT `id`,`name`,`phone`,`email`,`pwd_hash`,`tpw`,`tpw_exp` FROM `users` WHERE tpw=?",
		tpw,
	); err != nil {
		return User{}, errors.Wrapf(err, "failed to get user(id=%s)", tpw)
	}
	return u, nil
} //GetUserByTpw()

func GetUserByPhone(phone string) (User, error) {
	var err error
	phone, err = nationalPhone(phone)
	if err != nil {
		return User{}, errors.Wrapf(err, "cannot find user with invalid phone")
	}
	var u User
	if err := db.Get(&u,
		"SELECT `id`,`name`,`phone`,`email`,`pwd_hash`,`tpw`,`tpw_exp` FROM `users` WHERE phone=?",
		phone,
	); err != nil {
		return User{}, errors.Wrapf(err, "failed to get user(phone=%s)", phone)
	}
	return u, nil
} //GetUserByPhone()

func GetUserByEmail(email string) (User, error) {
	var u User
	if err := db.Get(&u,
		"SELECT `id`,`name`,`phone`,`email`,`pwd_hash`,`tpw`,`tpw_exp` FROM `users` WHERE email=?",
		email,
	); err != nil {
		return User{}, errors.Wrapf(err, "failed to get user(email=%s)", email)
	}
	return u, nil
} //GetUserByEmail()

func DelUser(id ID) error {
	if _, err := db.Exec("DELETE FROM `users` WHERE id=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete user(id=%s)", id)
	}
	return nil
}

func (u User) Compare(uu User) error {
	if u.ID != uu.ID {
		return errors.Errorf("id %s!=%s", u.ID, uu.ID)
	}
	if u.Name != uu.Name {
		return errors.Errorf("name %s!=%s", u.Name, uu.Name)
	}
	if u.Phone != uu.Phone {
		return errors.Errorf("phone %s!=%s", u.Phone, uu.Phone)
	}
	if u.Email != uu.Email {
		return errors.Errorf("email %s!=%s", u.Email, uu.Email)
	}
	return nil
}

var salt = "naephiesha9odahX5reewoutaico3oop" //default that can be changed with env var PASSWORD_SALT

func init() {
	if s := os.Getenv("PASSWORD_SALT"); s != "" {
		salt = s
	}
}

func HashPassword(email, pw string) string {
	h := sha1.New()
	s := email + pw + salt
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
