package db

import (
	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type User struct {
	ID    ID     `json:"id"`
	Name  string `json:"name" doc:"User name and surname used when coordinators contact the donar, or when user is represents a user as a coordinator."`
	Phone string `json:"phone" doc:"Phone must have 0 + 9 digits"`
}

func AddUser(newUser User) (User, error) {
	id := uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO `users` SET id=?,name=?,phone=?",
		id,
		newUser.Name,
		newUser.Phone,
	)
	if err != nil {
		return User{}, errors.Wrapf(err, "failed to insert user")
	}
	newUser.ID = ID(id)
	return newUser, nil
}

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
		"SELECT `id`,`name`,`phone` FROM `users` WHERE id=?",
		id,
	); err != nil {
		return User{}, errors.Wrapf(err, "failed to get user(id=%s)", id)
	}
	return u, nil
}

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
	return nil
}
