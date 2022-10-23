package db

import (
	"database/sql"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Member struct {
	ID      ID     `db:"id"`
	GroupID ID     `db:"group_id"`
	UserID  ID     `db:"user_id"`
	Role    string `db:"role"`
}

func AddMember(c Member) (Member, error) {
	id := uuid.New().String()
	if _, err := db.Exec("INSERT INTO `members` SET `id`=?,`group_id`=?,`user_id`=?",
		id,
		c.GroupID,
		c.UserID,
	); err != nil {
		return Member{}, errors.Wrapf(err, "failed to add member")
	}
	c.ID = ID(id)
	return c, nil
}

type MemberListEntry struct {
	ID      ID     `json:"id" db:"id"`
	GroupID ID     `json:"-" db:"group_id"`
	Group   *Group `json:"group,omitempty" db:"-"`
	UserID  ID     `json:"-" db:"user_id"`
	User    *User  `json:"user" db:"-"`
	Role    string `json:"role" db:"role"`
}

func ListGroupMembers(groupID ID) ([]MemberListEntry, error) {
	var members []MemberListEntry
	if err := db.Select(&members,
		"SELECT c.`id`,c.`user_id`,c.`role`,u.`name`,u.`phone`,u.`email` FROM `members` as c JOIN `users` as u ON c.`user_id`=u.`id` WHERE c.`group_id`=? ORDER BY u.`name`",
		groupID,
	); err != nil {
		return nil, errors.Wrapf(err, "failed to list group members")
	}
	return members, nil
}

func GetMemberByEmail(groupID ID, email string) (*Member, error) {
	var member Member
	if err := db.Get(&member, "SELECT m.`id`,m.`group_id`,m.`user_id`,m.`role` FROM `members` AS m JOIN `users` AS u on m.`user_id`=u.`id` WHERE m.`group_id`=? AND u.`email`=?",
		groupID,
		email,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //not found
		}
		return nil, errors.Wrapf(err, "failed to get group(id:%s) member by email(%s)", groupID, email)
	}
	return &member, nil
} //GetMemberByEmail()

func GetMembersBy(groupID ID, by string) (map[string]MemberListEntry, error) {
	members, err := ListGroupMembers(groupID)
	if err != nil {
		log.Errorf("GetMembersBy(%s,%s): failed to get members: %+v", groupID, by, err)
		return nil, err
	}
	memberByEmail := map[string]MemberListEntry{}
	for _, m := range members {
		switch by {
		case "email":
			memberByEmail[m.User.Email] = m
		default:
			log.Errorf("failed to get members with unknown key field(%s)", by)
			return nil, errors.Errorf("failed to get members")
		}
	}
	return memberByEmail, nil
}

func DelMember(id string) error {
	if _, err := db.Exec("DELETE FROM `members` WHERE id=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete member(id=%s)", id)
	}
	return nil
}
