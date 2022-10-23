package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/jansemmelink/events/email"
)

type Invitation struct {
	ID          ID               `json:"id" db:"id"`
	GroupID     ID               `json:"-" db:"group_id"`
	Group       *Group           `json:"group,omitempty" db:"-"`
	Email       string           `json:"email" db:"email"`
	TimeCreated SqlTime          `json:"time_created" db:"time_updates"`
	TimeUpdated SqlTime          `json:"time_updated" db:"time_updated"`
	Status      InvitationStatus `json:"status" db:"status"`
}

//Validate new invitation when received and about to create in db
func (inv *Invitation) Validate() error {
	if inv.GroupID == "" {
		return errors.Errorf("missing group_id")
	}
	if inv.Email == "" {
		return errors.Errorf("missing email")
	}
	if validEmail, err := email.Valid(inv.Email); err != nil {
		return errors.Errorf("invalid email(%s)", inv.Email)
	} else {
		inv.Email = validEmail
	}
	inv.TimeCreated = SqlTime(time.Now())
	inv.TimeUpdated = inv.TimeCreated
	return nil
}

func AddInvitation(req Invitation) (*Invitation, error) {
	if err := req.Validate(); err != nil {
		return nil, errors.Wrapf(err, "cannot add invalid invitation")
	}

	req.ID = ID(uuid.New().String())
	if _, err := db.Exec("INSERT INTO `invitations` SET `id`=?,`group_id`=?,`email`=?,`time_created`=?,`time_updated`=?",
		req.ID,
		req.GroupID,
		req.Email,
		req.TimeCreated,
		req.TimeUpdated,
	); err != nil {
		return nil, errors.Wrapf(err, "failed to create invitation")
	}
	return &req, nil
} //AddInvitation()

func GetInvitationByEmail(groupID ID, email string) (*Invitation, error) {
	var inv Invitation
	if err := db.Get(&inv, "SELECT `id`,`group_id`,`email`,`time_created`,`time_updates` FROM `invitations` AS i WHERE i.`group_id`=? AND i.`email`=?",
		groupID,
		email,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil //not found
		}
		return nil, errors.Wrapf(err, "failed to get group(id:%s) invitation by email(%s)", groupID, email)
	}
	return &inv, nil
} //GetInvitationByEmail()

//load pending invites for the specified group
//(any invite is pending, as they are deleted when the user joins the group)
//output is keyed on email, similar to GetMembers() output
func GetPendingInvitations(groupID ID) (map[string]Invitation, error) {
	var invitations []Invitation
	if err := db.Select(&invitations, "SELECT id,group_id,email,time_create,time_updated FROM `invitations` WHERE group_id=?", groupID); err != nil {
		log.Errorf("failed to get group(id:%s) invitations: %+v", groupID, err)
		return nil, errors.Errorf("failed to get invitations")
	}
	invitationByEmail := map[string]Invitation{}
	for _, i := range invitations {
		invitationByEmail[i.Email] = i
	}
	return invitationByEmail, nil
} //GetPendingInvitations()

func DelInviation(id ID) error {
	if _, err := db.Exec("DELETE FROM `invitations` WHERE id=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete invitation")
	}
	return nil
} //DelInvitation()

//=====[ ENUM: InvitationStatus ]=====
type InvitationStatus int

const (
	InvitationStatusNone InvitationStatus = iota
	InvitationStatusSent
	InvitationStatusBlocked
)

var (
	InvitationStatusValToStr = map[InvitationStatus]string{
		InvitationStatusSent:    "sent",
		InvitationStatusBlocked: "blocked",
	}
	InvitationStatusStrToVal = map[string]InvitationStatus{}
)

func init() {
	for val, str := range InvitationStatusValToStr {
		InvitationStatusStrToVal[str] = val
	}
}

func (t *InvitationStatus) Scan(value interface{}) error {
	*t = InvitationStatusNone
	if value == nil {
		return nil //=None
	}
	if byteArray, ok := value.([]uint8); ok {
		strValue := string(byteArray)
		if strValue == "" {
			return nil //=None
		}
		var ok bool
		if *t, ok = InvitationStatusStrToVal[strValue]; !ok {
			return errors.Errorf("unknown invitation status(%s)", strValue)
		}
		return nil
	}
	return errors.Errorf("%T is not []uint8", value)
}

func (t InvitationStatus) Value() (driver.Value, error) {
	return InvitationStatusValToStr[t], nil
}

func (t InvitationStatus) String() string {
	return InvitationStatusValToStr[t]
}

func (t *InvitationStatus) UnmarshalJSON(v []byte) error {
	s := string(v)
	if len(s) < 2 || !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
		return errors.Errorf("invalid invitation status string %s (expects quoted string)", s)
	}
	return t.Scan(v[1 : len(v)-1])
}

func (t InvitationStatus) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", t.String())
	return []byte(s), nil
}
