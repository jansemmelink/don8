package db

import (
	"strings"
	"time"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Promise struct {
	ID         ID      `db:"id"`
	RequestID  ID      `db:"request_id" doc:"This describes the items being donated"`
	UserID     ID      `db:"user_id" doc:"The user promising to make the donation"`
	LocationID *ID     `db:"location_id" doc:"Location where user intend to make the donation, or NULL if cannot commit."`
	Qty        int     `db:"qty" doc:"Quantity that user promise to donate"`
	Date       SqlTime `db:"date" doc:"Date by when user promise to make the donation"`
}

func AddPromise(p Promise) (Promise, error) {
	id := uuid.New().String()
	if _, err := db.Exec(
		"INSERT INTO `promises` SET `id`=?,`user_id`=?,`request_id`=?,`location_id`=?,`qty`=?,`date`=?",
		id,
		p.UserID,
		p.RequestID,
		p.LocationID,
		p.Qty,
		p.Date,
	); err != nil {
		return Promise{}, errors.Wrapf(err, "failed to add promise")
	}
	p.ID = ID(id)
	return p, nil
}

type PromiseListEntry struct {
	ID            ID      `db:"id"`
	GroupID       ID      `db:"group_id"`
	UserID        ID      `db:"user_id"`
	UserName      string  `db:"user_name"`
	UserPhone     string  `db:"user_phone"`
	RequestID     ID      `db:"request_id"`
	RequestTitle  string  `db:"request_title"`
	RequestQty    int     `db:"request_qty"`
	LocationID    *ID     `db:"location_id" doc:"Location where user intend to make the donation"`
	LocationTitle *string `db:"location_title"`
	Qty           int     `db:"promise_qty" doc:"Quantity that user promise to donate"`
	Date          SqlTime `db:"date" doc:"Date by when user promise to make the donation"`
}

//groupID is required
func GetPromises(groupID string, userID string, requestID string, locationID string, beforeDate *time.Time, orderColumns []string) ([]PromiseListEntry, error) {
	if groupID == "" {
		return nil, errors.Errorf("missing group_id filter")
	}

	sql := "SELECT p.`id`,r.`group_id`,p.`user_id`,u.`name` AS `user_name`,u.`phone` AS `user_phone`,p.`request_id`,r.`title` AS `request_title`,p.`location_id`,l.`title` AS `location_title`,p.`qty` AS `promise_qty`,p.`date`,r.`qty` AS `request_qty` FROM `promises` as p JOIN `requests` as r ON p.`request_id`=r.`id` JOIN `users` AS u ON p.`user_id`=u.`id` JOIN `locations` AS l ON p.`location_id`=l.`id` WHERE r.`group_id`=?"
	args := []interface{}{groupID}

	if userID != "" {
		sql += " AND `user_id`=?"
		args = append(args, userID)
	}
	if requestID != "" {
		sql += " AND `request_id`=?"
		args = append(args, requestID)
	}
	if locationID != "" {
		sql += " AND `location_id`=?"
		args = append(args, locationID)
	}
	if beforeDate != nil {
		sql += " AND `date`<?"
		args = append(args, *beforeDate)
	}

	if len(orderColumns) > 0 {
		sql += " ORDER BY " + strings.Join(orderColumns, ",")
	}

	var promises []PromiseListEntry
	if err := db.Select(&promises, sql, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to list promises")
	}
	return promises, nil
}

func GetPromise(id ID) (Promise, error) {
	var p Promise
	if err := db.Get(&p, "SELECT `id`,`request_id`,`user_id`,`location_id`,`qty`,`date` FROM `promises` WHERE `id`=?", id); err != nil {
		return Promise{}, errors.Wrapf(err, "failed to get promise(id=%s)", id)
	}
	return p, nil
}
