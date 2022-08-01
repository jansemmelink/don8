package db

import (
	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Group struct {
	ID            ID     `json:"id"`
	ParentGroupID ID     `json:"parent_group_id" db:"parent_group_id"`
	ParentTitle   string `json:"parent_title" db:"parent_title"`
	Title         string `json:"title" doc:"Descriptive title of the group, e.g. \"AHS/AHMP\" or \"Affies Wildsfees 2022\""`
	Description   string `json:"description" doc:"Descriptive paragraph about the group."`
}

func AddGroup(newGroup Group) (Group, error) {
	var parentTitle string
	if newGroup.ParentGroupID != "" {
		var parentGroup Group
		if err := db.Get(&parentGroup,
			"SELECT `parent_title`,`title` FROM `groups` WHERE id=?",
			newGroup.ParentGroupID,
		); err != nil {
			return Group{}, errors.Wrapf(err, "failed to get parent group")
		}
		parentTitle = parentGroup.ParentTitle
		if parentTitle != "" {
			parentTitle += " "
		}
		parentTitle += parentGroup.Title
	}

	id := uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO `groups` SET id=?,parent_group_id=?,parent_title=?,title=?,description=?,search=?",
		id,
		newGroup.ParentGroupID,
		parentTitle,
		newGroup.Title,
		newGroup.Description,
		parentTitle+" "+newGroup.Title+" "+newGroup.Description, //=search
	)
	if err != nil {
		return Group{}, errors.Wrapf(err, "failed to insert group")
	}
	newGroup.ID = ID(id)
	return newGroup, nil
}

func FindGroups(filter string, limit int) ([]Group, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var rows []Group
	if err := db.Select(&rows,
		"SELECT `id`,`parent_title`,`title`,`description` FROM `groups` WHERE search like ? LIMIT ?",
		"%"+filter+"%",
		limit); err != nil {
		return nil, errors.Wrapf(err, "failed to search groups")
	}
	return rows, nil
}

func GetGroup(id ID) (Group, error) {
	var g Group
	if err := db.Get(&g,
		"SELECT id,parent_group_id,parent_title,title,description FROM `groups` WHERE id=?",
		id,
	); err != nil {
		return Group{}, errors.Wrapf(err, "failed to get group(id=%s)", id)
	}
	return g, nil
}

func DelGroup(id ID) error {
	if _, err := db.Exec("DELETE FROM `groups` WHERE id=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete group(id=%s)", id)
	}
	return nil
}

func (g Group) Compare(gg Group) error {
	if g.ID != gg.ID {
		return errors.Errorf("id %s!=%s", g.ID, gg.ID)
	}
	if g.Title != gg.Title {
		return errors.Errorf("title %s!=%s", g.Title, gg.Title)
	}
	if g.Description != gg.Description {
		return errors.Errorf("description %s!=%s", g.Description, gg.Description)
	}
	if g.ParentTitle != gg.ParentTitle {
		return errors.Errorf("parent_title %s!=%s", g.ParentTitle, gg.ParentTitle)
	}
	if g.ParentGroupID != gg.ParentGroupID {
		return errors.Errorf("parent_group_id %s!=%s", g.ParentGroupID, gg.ParentGroupID)
	}
	return nil
}
