package db

import (
	"strings"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Request struct {
	ID          ID     `db:"id"`
	GroupID     ID     `db:"group_id"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Tags        string `db:"tags" doc:"Written as |<tag>|<tag>|...| so we can search with SQL tags like \"|<sometag>|\""`
	Unit        string `db:"unit" doc:"Unit of measurement, e.g. \"items\" or \"kg\" or \"L\" or \"dozen\" etc..."`
	Qty         int    `db:"qty" doc:"Quantity requested in total from all donars"`
}

func AddRequest(r Request) (Request, error) {
	id := uuid.New().String()
	if _, err := db.Exec("INSERT INTO `requests` SET `id`=?,`group_id`=?,`title`=?,`description`=?,`tags`=?,`unit`=?,`qty`=?",
		id,
		r.GroupID,
		r.Title,
		r.Description,
		r.Tags,
		r.Unit,
		r.Qty,
	); err != nil {
		return Request{}, errors.Wrapf(err, "failed to add request")
	}
	r.ID = ID(id)
	return r, nil
}

func FindRequests(groupID ID, filter string, tags []string, limit int) ([]Request, error) {
	sql := "SELECT id,group_id,title,description,tags,unit,qty FROM `requests` WHERE `group_id`=?"
	args := []interface{}{groupID}

	if filter != "" {
		sql += " AND (title like ? OR description like ?)"
		args = append(args, "%"+filter+"%") //for title like ...
		args = append(args, "%"+filter+"%") //for description like ...
	}

	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			sql += " AND tags like ?"
			args = append(args, "%|"+tag+"|%")
		}
	}

	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	sql += " LIMIT ?"
	args = append(args, limit)

	var requests []Request
	if err := db.Select(&requests, sql, args...); err != nil {
		return nil, errors.Wrapf(err, "failed to find requests")
	}
	return requests, nil
}

func GetRequest(id ID) (Request, error) {
	var request Request
	if err := db.Get(&request, "SELECT id,group_id,title,description,tags,unit,qty FROM requests WHERE id=?", id); err != nil {
		return Request{}, errors.Wrapf(err, "failed to get request(id=%s)", id)
	}
	return request, nil
}

func DelRequest(id ID) error {
	if _, err := db.Exec("DELETE FROM `requests` WHERE `id`=?", id); err != nil {
		return errors.Wrapf(err, "failed to delete request(id=%s)", id)
	}
	return nil
}
