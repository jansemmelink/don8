package db

import (
	"net/http"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Request struct {
	ID          ID      `json:"id" db:"id"`
	GroupID     ID      `json:"group_id" db:"group_id"`
	Title       string  `json:"title" db:"title"`
	Description *string `json:"description" db:"description"`
	Tags        *string `json:"tags" db:"tags" doc:"Written as |<tag>|<tag>|...| so we can search with SQL tags like \"|<sometag>|\""`
	Units       *string `json:"units" db:"units" doc:"Unit of measurement, e.g. \"items\" or \"kg\" or \"L\" or \"dozen\" etc..."`
	Qty         int     `json:"qty" db:"qty" doc:"Quantity requested in total from all donars"`
}

func TagsFromString(s string) []string {
	s = strings.ReplaceAll(s, ",", "|")
	s = strings.ReplaceAll(s, " ", "|")
	tags := strings.Split(s, "|")
	for i := 0; i < len(tags); i++ {
		tags[i] = strings.TrimSpace(tags[i])
	}
	nonEmptyTags := []string{}
	for _, t := range tags {
		if t != "" {
			nonEmptyTags = append(nonEmptyTags, t)
		}
	}
	return nonEmptyTags
}

func tagString(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "|" + strings.Join(tags, "|") + "|"
}

func (req *Request) Validate() error {
	if req.GroupID == "" {
		return errors.Errorf("missing group_id")
	}
	if req.Title == "" {
		return errors.Errorf("missing title")
	}
	//Description is optional, but remove outer spaces
	if req.Description != nil {
		*req.Description = strings.TrimSpace(*req.Description)
		if *req.Description == "" {
			req.Description = nil
		}
	}
	//Tags are optional, but convert to valid pipe-separated tags
	if req.Tags != nil {
		tags := tagString(TagsFromString(*req.Tags))
		if tags != "" {
			req.Tags = &tags
		} else {
			req.Tags = nil
		}
	}
	if req.Qty < 1 {
		return errors.Errorf("missing qty")
	}
	return nil
}

func AddRequest(r Request) (Request, error) {
	if err := r.Validate(); err != nil {
		return Request{}, errors.Errorc(http.StatusBadRequest, err.Error())
	}
	id := uuid.New().String()
	if _, err := db.Exec("INSERT INTO `requests` SET `id`=?,`group_id`=?,`title`=?,`description`=?,`tags`=?,`units`=?,`qty`=?",
		id,
		r.GroupID,
		r.Title,
		r.Description,
		r.Tags,
		r.Units,
		r.Qty,
	); err != nil {
		return Request{}, errors.Wrapf(err, "failed to add request")
	}
	r.ID = ID(id)
	return r, nil
}

func FindRequests(groupID ID, filter string, tags []string, limit int) ([]Request, error) {
	sql := "SELECT id,group_id,title,description,tags,units,qty FROM `requests` WHERE `group_id`=?"
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
	if err := db.Get(&request, "SELECT id,group_id,title,description,tags,units,qty FROM requests WHERE id=?", id); err != nil {
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
