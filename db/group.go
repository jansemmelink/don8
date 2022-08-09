package db

import (
	"strings"
	"time"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

var localTime = time.Now().Location()

type NewGroup struct {
	ParentGroupID ID       `json:"parent_group_id" db:"parent_group_id"`
	Title         string   `json:"title" db:"title" doc:"Descriptive title of the group, e.g. \"AHS/AHMP\" or \"Affies Wildsfees 2022\""`
	Description   *string  `json:"description" db:"description" doc:"Descriptive paragraph about the group."`
	Start         *string  `json:"start" db:"start" doc:"Start date and time as CCYY-MM-DD HH:MM or just CCYY-MM-DD"`
	End           *string  `json:"end" db:"end" doc:"End date and time as CCYY-MM-DD HH:MM or just CCYY-MM-DD"`
	UserRole      string   `json:"user_role" db:"user_role" doc:"Role/Title of the current user in this group, e.g. Head Master or Event Organiser etc..."`
	startTime     *SqlTime //from Validate() and optional
	endTime       *SqlTime //from Validate() and optional
}

func (g *NewGroup) Validate() error {
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" {
		return errors.Errorf("missing title")
	}
	if g.Description != nil {
		*g.Description = strings.TrimSpace(*g.Description) //optional
	}

	if g.Start != nil || g.End != nil {
		if g.Start == nil {
			return errors.Errorf("end specified without start")
		}
		*g.Start = strings.TrimSpace(*g.Start)
		st, err := time.ParseInLocation("2006-01-02 15:04", *g.Start, localTime)
		if err != nil {
			if st, err = time.ParseInLocation("2006-01-02", *g.Start, localTime); err != nil {
				return errors.Errorf("invalid start \"%s\" expected CCYY-MM-DD or CCYY-MM-DD HH:MM", *g.Start)
			}
		}
		if g.End == nil || *g.End == "" {
			g.End = g.Start
		}
		*g.End = strings.TrimSpace(*g.End) //optional but correct format if specified
		et, err := time.ParseInLocation("2006-01-02 15:04", *g.End, localTime)
		if err != nil {
			if et, err = time.ParseInLocation("2006-01-02", *g.End, localTime); err != nil {
				return errors.Errorf("invalid end \"%s\" expected CCYY-MM-DD or CCYY-MM-DD HH:MM", *g.End)
			}
		}
		if et.Before(st) {
			return errors.Errorf("end \"%s\" is before start \"%s\"", *g.End, *g.Start)
		}
		sqlst := SqlTime(st)
		g.startTime = &sqlst
		sqlet := SqlTime(et)
		g.endTime = &sqlet
	} //if time specified
	g.UserRole = strings.TrimSpace(g.UserRole) //required
	if g.UserRole == "" {
		return errors.Errorf("missing user_role")
	}
	return nil
}

type Group struct {
	ID            ID       `json:"id"`
	ParentGroupID ID       `json:"parent_group_id,omitempty" db:"parent_group_id"`
	Title         string   `json:"title" db:"title" doc:"Descriptive title of the group, e.g. \"AHS/AHMP\" or \"Affies Wildsfees 2022\""`
	Description   *string  `json:"description,omitempty" db:"description" doc:"Descriptive paragraph about the group."`
	Start         *SqlTime `json:"start,omitempty" db:"start" doc:"Optional start date and time as CCYY-MM-DD HH:MM or just CCYY-MM-DD"`
	End           *SqlTime `json:"end,omitempty" db:"end" doc:"Optional end date and time as CCYY-MM-DD HH:MM or just CCYY-MM-DD"`
}

//Upload logo separately
//Gallery of pictures and documents in other table... generic attachments

func AddGroup(user User, newGroup NewGroup) (Group, error) {
	id := uuid.New().String()
	_, err := db.Exec(
		"INSERT INTO `groups` SET id=?,parent_group_id=?,title=?,description=?,start=?,end=?",
		id,
		newGroup.ParentGroupID,
		newGroup.Title,
		newGroup.Description,
		newGroup.Start,
		newGroup.End,
	)
	if err != nil {
		return Group{}, errors.Wrapf(err, "failed to insert group")
	}

	g := Group{
		ID:            ID(id),
		ParentGroupID: newGroup.ParentGroupID,
		Title:         newGroup.Title,
		Description:   newGroup.Description,
	}
	if newGroup.startTime != nil {
		g.Start = newGroup.startTime
		g.End = newGroup.endTime
	}

	//add user as the group admin
	cid := uuid.New().String()
	if _, err := db.Exec("INSERT INTO members SET id=?,group_id=?,user_id=?,role=?",
		cid,
		g.ID,
		user.ID,
		newGroup.UserRole, //no meaning - user can change it later... permissions are in member_permissions
	); err != nil {
		log.Errorf("failed to add member: %+v", err)
		return Group{}, errors.Errorf("failed to create group")
	}

	if _, err := db.Exec("INSERT INTO member_permissions SET member_id=?,permissions=?",
		cid,
		"*", //all permissions
	); err != nil {
		log.Errorf("failed to add group_member: %+v", err)
		return Group{}, errors.Wrapf(err, "failed to create group")
	}
	return g, nil
}

//list my groups and my permissions
//which will indicate when I follow or coordinate
type MyGroup struct {
	ID          ID       `json:"group_id" db:"group_id" doc:"Group ID"`
	ParentID    ID       `json:"parent_group_id" db:"parent_group_id,omitempty" doc:"Parent Group ID"`
	Title       string   `json:"title" db:"title" doc:"Title of this group"`
	Description *string  `json:"description" db:"description" doc:"Group description"`
	Start       *string  `json:"start" db:"start" doc:"Group start date and time"`
	End         *string  `json:"end" db:"end" doc:"Group end date and time"`
	Role        string   `json:"role" db:"role" doc:"My role in this group"`
	Permission  []string `json:"permissions" db:"-" doc:"List of permissions, |*| for group owner(s)."`
}

func MyGroups(user User, filter string, fromTime *time.Time, toTime *time.Time) ([]MyGroup, error) {
	var list []MyGroup
	sql := "SELECT g.`id` AS `group_id`,g.`parent_group_id`,g.`title`,g.`description`,g.`start`,g.`end` FROM `groups` AS g JOIN `members` AS m ON m.`group_id`=g.`id` WHERE m.`user_id`=?"
	args := []interface{}{user.ID}

	filter = strings.TrimSpace(filter)
	if filter != "" {
		sql += " AND g.`title` LIKE ?"
		args = append(args, "%"+filter+"%")
	}
	if fromTime != nil {
		sql += " AND g.`end` > ?"
		args = append(args, SqlTime(*fromTime))
	}
	if toTime != nil {
		sql += " AND g.`start` < ?"
		args = append(args, SqlTime(*toTime))
	}
	sql += " ORDER BY g.`start`"
	if err := db.Select(&list, sql, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to list my own groups")
	}
	return list, nil
}

func GetGroup(id ID) (Group, error) {
	var g Group
	if err := db.Get(&g,
		"SELECT id,parent_group_id,title,description FROM `groups` WHERE id=?",
		id,
	); err != nil {
		return Group{}, errors.Wrapf(err, "failed to get group(id=%s)", id)
	}
	return g, nil
}

type FullGroup struct {
	Parent *Group `json:"parent,omitempty"`
	Group
	Requests []Request `json:"requests,omitempty"` //empty, but can be populated by calling Requests()
	Children []Group   `json:"children,omitempty"`
}

func GetFullGroup(id ID) (FullGroup, error) {
	var g Group
	if err := db.Get(&g,
		"SELECT id,parent_group_id,title,description FROM `groups` WHERE id=?",
		id,
	); err != nil {
		return FullGroup{}, errors.Wrapf(err, "failed to get group(id=%s)", id)
	}
	fg := FullGroup{
		Group: g,
	}
	if g.ParentGroupID != "" {
		pg, err := GetGroup(g.ParentGroupID)
		if err != nil {
			return FullGroup{}, errors.Wrapf(err, "failed to get group(id=%s).parent(id=%s)", id, g.ParentGroupID)
		}
		fg.Parent = &pg
	}
	if err := db.Select(&fg.Children, "SELECT id,parent_group_id,title,description FROM `groups` WHERE `parent_group_id`=? ORDER BY `title`", id); err != nil {
		log.Errorf("failed to read group(%s).children: %+v", id, err)
	}
	return fg, nil
}

type UpdGroupRequest struct {
	ID          ID      `json:"id"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (req UpdGroupRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	if req.Title != nil {
		*req.Title = strings.TrimSpace(*req.Title)
		if *req.Title == "" {
			return errors.Errorf("empty title not allowed")
		}
	}
	if req.Description != nil {
		*req.Description = strings.TrimSpace(*req.Description)
		if *req.Description == "" {
			return errors.Errorf("empty description not allowed")
		}
	}
	return nil
}

func UpdGroup(req UpdGroupRequest) error {
	sql := "UPDATE `groups` SET"
	args := []interface{}{}
	changes := 0
	if req.Title != nil && *req.Title != "" {
		if changes > 0 {
			sql += ","
		} else {
			sql += " "
		}
		sql += "`title`=?"
		args = append(args, *req.Title)
		changes++
	}
	if req.Description != nil && *req.Description != "" {
		if changes > 0 {
			sql += ","
		} else {
			sql += " "
		}
		sql += "`description`=?"
		args = append(args, *req.Description)
		changes++
	}
	if changes < 1 {
		return errors.Errorf("no changes specified")
	}
	_, err := db.Exec(sql, args...)
	if err != nil {
		log.Errorf("failed to update group: %+v", err)
		return errors.Errorf("failed to update")
	}

	return nil
} //UpdGroup()

func DelGroup(id ID) error {
	if _, err := db.Exec("DELETE FROM `member_permissions`"+
		" WHERE `member_id` in ("+
		"SELECT p.`member_id` FROM `member_permissions` AS p"+
		" JOIN `members` as m on m.`id`=p.`member_id`"+
		" JOIN `groups` AS g on g.`id`=m.`group_id`"+
		" WHERE g.`id`=?)",
		id,
	); err != nil {
		return errors.Wrapf(err, "failed to delete group member_permissions")
	}

	if _, err := db.Exec("DELETE FROM `members` WHERE group_id=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete group members")
	}

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
	if optStr(g.Description) != optStr(gg.Description) {
		return errors.Errorf("description %s!=%s", optStr(g.Description), optStr(gg.Description))
	}
	if g.ParentGroupID != gg.ParentGroupID {
		return errors.Errorf("parent_group_id %s!=%s", g.ParentGroupID, gg.ParentGroupID)
	}
	return nil
}

func optStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
