package db

import "github.com/go-msvc/errors"

type Permission string

type MemberPermission struct {
	MemberID   ID
	Permission Permission
}

func AddMemberPermission(cp MemberPermission) (MemberPermission, error) {
	if _, err := db.Exec(
		"INSERT INTO `member_permissions` SET member_id=?,permission=?",
		cp.MemberID,
		cp.Permission,
	); err != nil {
		return MemberPermission{}, errors.Wrapf(err, "failed to add member permission")
	}
	return cp, nil
}

func ListMemberPermissions(memberID ID) ([]Permission, error) {
	var cps []MemberPermission
	if err := db.Select(&cps,
		"SELECT FROM `member_permissions` WHERE member_id=? ORDER BY permission",
		memberID,
	); err != nil {
		return nil, errors.Wrapf(err, "failed to list member permissions")
	}
	list := make([]Permission, len(cps))
	for i, cp := range cps {
		list[i] = cp.Permission
	}
	return list, nil
}

func DelMemberPermission(memberID ID, permissionList []Permission) error {
	if len(permissionList) == 1 && permissionList[0] == "*" {
		if _, err := db.Exec(
			"DELETE FROM `member_permissions` WHERE member_id=?",
			memberID,
		); err != nil {
			return errors.Wrapf(err, "failed to delete all permissions for member(id=%s)", memberID)
		}
		return nil
	} //if delete all

	//delete selected permissions
	for _, p := range permissionList {
		if _, err := db.Exec(
			"DELETE FROM `member_permissions` WHERE member_id=? AND permission=?",
			memberID,
			p,
		); err != nil {
			return errors.Wrapf(err, "failed to delete member(id=%s) permission(%s)", memberID, p)
		}
	}
	return nil
}
