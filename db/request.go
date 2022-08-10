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

//tagString return "|xxx|yyy|zzz|" as we store tags in the db for easy match with SQL like %"|xxx|"%
func tagsDbString(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "|" + strings.Join(tags, "|") + "|"
}

//tagCSV return "xxx,yyy,zzz" as we display tags to the API
func TagsCSV(tags []string) string {
	return strings.Join(tags, ",")
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
		tags := tagsDbString(TagsFromString(*req.Tags))
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

type FullRequest struct {
	Group Group `json:"group"`
	Request
	Promises []Promise `json:"promises,omitempty"`
	//Receives []Receive `json:"receives,omitempty"`
}

func GetFullRequest(id ID) (FullRequest, error) {
	r, err := GetRequest(id)
	if err != nil {
		return FullRequest{}, err
	}
	g, err := GetGroup(r.GroupID)
	if err != nil {
		return FullRequest{}, err
	}
	fr := FullRequest{
		Group:   g,
		Request: r,
		//		Promises: []Promise{},
	}
	// if err := db.Select(&fr.Promises, "SELECT id,parent_group_id,title,description FROM `groups` WHERE `parent_group_id`=? ORDER BY `title`", id); err != nil {
	// 	log.Errorf("failed to read group(%s).children: %+v", id, err)
	// }

	//return tags as CSV to the API
	return fr, nil
} //GetFullRequest()

type UpdRequestRequest struct {
	ID          ID      `json:"id"`
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Tags        *string `json:"tags,omitempty"`
	Units       *string `json:"units,omitempty"`
	Qty         *int    `json:"qty,omitempty"`
}

func (req *UpdRequestRequest) Validate() error {
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
		//allow empty description...
		// if *req.Description == "" {
		// 	return errors.Errorf("empty description not allowed")
		// }
	}

	//Tags are optional, but convert to valid pipe-separated tags
	if req.Tags != nil {
		tags := tagsDbString(TagsFromString(*req.Tags))
		if tags != "" {
			*req.Tags = tags
		} else {
			req.Tags = nil
		}
	}
	if req.Units != nil {
		*req.Units = strings.TrimSpace(*req.Units) //empty units are allowed
	}
	if req.Qty != nil {
		if *req.Qty < 0 {
			return errors.Errorf("invalid new qty:%d", *req.Qty)
		}
	}
	return nil
}

func UpdRequest(req UpdRequestRequest) error {
	sql := "UPDATE `requests` SET"
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
	if req.Description != nil { //may be ""
		if changes > 0 {
			sql += ","
		} else {
			sql += " "
		}
		sql += "`description`=?"
		args = append(args, *req.Description)
		changes++
	}
	if req.Tags != nil { //may be ""
		if changes > 0 {
			sql += ","
		} else {
			sql += " "
		}
		sql += "`tags`=?"
		args = append(args, *req.Tags)
		changes++
	}
	if req.Units != nil { //may be ""
		if changes > 0 {
			sql += ","
		} else {
			sql += " "
		}
		sql += "`units`=?"
		args = append(args, *req.Description)
		changes++
	}
	if req.Qty != nil { //may be 0
		if changes > 0 {
			sql += ","
		} else {
			sql += " "
		}
		sql += "`qty`=?"
		args = append(args, *req.Qty)
		changes++
	}
	if changes < 1 {
		return errors.Errorf("no changes specified")
	}

	//finish the query SQL then exec
	sql += " WHERE `id`=?"
	args = append(args, req.ID)
	_, err := db.Exec(sql, args...)
	if err != nil {
		log.Errorf("failed to update request: %+v", err)
		return errors.Errorf("failed to update")
	}
	return nil
} //UpdRequest()
