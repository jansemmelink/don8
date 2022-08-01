package db

import (
	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Location struct {
	ID               ID     `db:"id"`
	GroupID          ID     `db:"group_id"`
	Title            string `db:"title"`
	Description      string `db:"description"`
	FinalDestination bool   `db:"final_destination"`
}

func AddLocation(c Location) (Location, error) {
	id := uuid.New().String()
	if _, err := db.Exec("INSERT INTO `locations` SET `id`=?,`group_id`=?,`title`=?,`description`=?,`final_destination`=?",
		id,
		c.GroupID,
		c.Title,
		c.Description,
		c.FinalDestination,
	); err != nil {
		return Location{}, errors.Wrapf(err, "failed to add location")
	}
	c.ID = ID(id)
	return c, nil
}

func ListGroupLocations(groupID ID) ([]Location, error) {
	var locations []Location
	if err := db.Select(&locations,
		"SELECT `id`,`title`,`description`,`final_destination` FROM `locations` WHERE `group_id`=? ORDER BY `title`",
		groupID,
	); err != nil {
		return nil, errors.Wrapf(err, "failed to list group locations")
	}
	return locations, nil
}

func DelLocation(id string) error {
	if _, err := db.Exec("DELETE FROM `locations` WHERE id=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete location(id=%s)", id)
	}
	return nil
}
