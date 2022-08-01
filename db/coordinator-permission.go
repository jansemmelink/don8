package db

import "github.com/go-msvc/errors"

type Permission string

type CoordinatorPermission struct {
	CoordinatorID ID
	Permission    Permission
}

func AddCoordinatorPermission(cp CoordinatorPermission) (CoordinatorPermission, error) {
	if _, err := db.Exec(
		"INSERT INTO `coordinator_permissions` SET coordinator_id=?,permission=?",
		cp.CoordinatorID,
		cp.Permission,
	); err != nil {
		return CoordinatorPermission{}, errors.Wrapf(err, "failed to add coordinator permission")
	}
	return cp, nil
}

func ListCoordinatorPermissions(coordinatorID ID) ([]Permission, error) {
	var cps []CoordinatorPermission
	if err := db.Select(&cps,
		"SELECT FROM `coordinator_permissions` WHERE coordinator_id=? ORDER BY permission",
		coordinatorID,
	); err != nil {
		return nil, errors.Wrapf(err, "failed to list coordinator permissions")
	}
	list := make([]Permission, len(cps))
	for i, cp := range cps {
		list[i] = cp.Permission
	}
	return list, nil
}

func DelCoordinatorPermission(coordinatorID ID, permissionList []Permission) error {
	if len(permissionList) == 1 && permissionList[0] == "*" {
		if _, err := db.Exec(
			"DELETE FROM `coordinator_permissions` WHERE coordinator_id=?",
			coordinatorID,
		); err != nil {
			return errors.Wrapf(err, "failed to delete all permissions for coordinator(id=%s)", coordinatorID)
		}
		return nil
	} //if delete all

	//delete selected permissions
	for _, p := range permissionList {
		if _, err := db.Exec(
			"DELETE FROM `coordinator_permissions` WHERE coordinator_id=? AND permission=?",
			coordinatorID,
			p,
		); err != nil {
			return errors.Wrapf(err, "failed to delete coordinator(id=%s) permission(%s)", coordinatorID, p)
		}
	}
	return nil
}
