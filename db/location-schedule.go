package db

import (
	"time"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type LocationSchedule struct {
	ID         ID
	LocationID ID
	OpenTime   SqlTime
	CloseTime  SqlTime
	MemberID   ID
}

func AddLocationSchedule(ls LocationSchedule) (LocationSchedule, error) {
	id := uuid.New().String()
	if _, err := db.Exec("INSERT into `location_schedules` SET id=?,location_id=?,open_time=?,close_time=?,member_id=?",
		id,
		ls.LocationID,
		ls.OpenTime,
		ls.CloseTime,
		ls.MemberID,
	); err != nil {
		return LocationSchedule{}, errors.Wrapf(err, "failed to add location_schedule")
	}
	ls.ID = ID(id)
	return ls, nil
}

func ListLocationSchedules(locationID ID, from *time.Time, to *time.Time, memberID *ID) ([]LocationSchedule, error) {
	sql := "SELECT `id`,`location_id`,`open_time`,`close_time`,`member_id` FROM `location_schedules` WHERE `location_id`=?"
	args := []interface{}{locationID}

	if from != nil {
		sql += " AND open_time>=?"
		args = append(args, *from)
	}

	if to != nil {
		sql += " AND close_time<=?"
		args = append(args, *to)
	}

	if memberID != nil {
		sql += " AND member_id=?"
		args = append(args, *memberID)
	}

	sql += " ORDER BY open_time"
	var lss []LocationSchedule
	if err := db.Select(&lss, sql, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to list location schedules")
	}
	return lss, nil
}
