package db

import (
	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Coordinator struct {
	ID      ID     `db:"id"`
	GroupID ID     `db:"group_id"`
	UserID  ID     `db:"user_id"`
	Role    string `db:"role"`
}

func AddCoordinator(c Coordinator) (Coordinator, error) {
	id := uuid.New().String()
	if _, err := db.Exec("INSERT INTO `coordinators` SET `id`=?,`group_id`=?,`user_id`=?",
		id,
		c.GroupID,
		c.UserID,
	); err != nil {
		return Coordinator{}, errors.Wrapf(err, "failed to add coordinator")
	}
	c.ID = ID(id)
	return c, nil
}

type CoordinatorListEntry struct {
	ID        ID     `db:"id"`
	GroupID   ID     `db:"group_id"`
	UserID    ID     `db:"user_id"`
	Role      string `db:"role"`
	UserName  string `db:"name"`
	UserPhone string `db:"phone"`
}

func ListGroupCoordinators(groupID ID) ([]CoordinatorListEntry, error) {
	var coordinators []CoordinatorListEntry
	if err := db.Select(&coordinators,
		"SELECT c.`id`,c.`user_id`,c.`role`,u.`name`,u.`phone` FROM `coordinators` as c JOIN `users` as u ON c.`user_id`=u.`id` WHERE c.`group_id`=? ORDER BY u.`name`",
		groupID,
	); err != nil {
		return nil, errors.Wrapf(err, "failed to list group coordinators")
	}
	return coordinators, nil
}

func DelCoordinator(id string) error {
	if _, err := db.Exec("DELETE FROM `coordinators` WHERE id=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete coordinator(id=%s)", id)
	}
	return nil
}
