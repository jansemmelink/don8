package db

import (
	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type Donation struct {
	ID         ID     `db:"id"`
	LocationID ID     `db:"location_id" doc:"Location where donation was made."`
	RequestID  *ID    `db:"request_id,omitempty" doc:"Request is defined if received requested or promised items. Absent for ad hoc drop of general goods not specifically requested."`
	PromiseID  *ID    `db:"promise_id,omitempty" doc:"Promise is defined when a user delivers on a promise. Absent for ad hoc anonymous drops."`
	Title      string `db:"title" doc:"From request.title, or free text when receiving items without specific request."`
	Unit       string `db:"unit" doc:"From request.unit, or free text when receiving items without specific request."`
	Qty        int    `db:"qty" doc:"Nr of units donated"`
}

func AddDonation(d Donation) (Donation, error) {
	//validate optional promise reference
	var promise *Promise
	if d.PromiseID != nil {
		p, err := GetPromise(*d.PromiseID)
		if err != nil {
			return Donation{}, errors.Errorf("unknown PromiseID(%s)", *d.PromiseID)
		}
		promise = &p
		if d.RequestID == nil {
			d.RequestID = &promise.RequestID
		}
	}

	var request *Request
	if d.RequestID != nil {
		if promise != nil && promise.RequestID != *d.RequestID {
			return Donation{}, errors.Errorf("donation.request_id=%s != donation.promise.request_id=%s", *d.RequestID, promise.RequestID)
		}

		r, err := GetRequest(*d.RequestID)
		if err != nil {
			return Donation{}, errors.Wrapf(err, "failed to get request(id=%s)")
		}
		request = &r
		if d.Title != "" && d.Title != request.Title {
			return Donation{}, errors.Errorf("donation.title(%s) != donation.request.title(%s)", d.Title, request.Title)
		}
		d.Title = request.Title
		if d.Unit != "" && d.Unit != request.Unit {
			return Donation{}, errors.Errorf("donation.unit(%s) != donation.request.unit(%s)", d.Unit, request.Unit)
		}
		d.Unit = request.Unit
	}

	//Title and Unit must be specified or be obtained from the request
	if d.Title == "" {
		return Donation{}, errors.Errorf("cannot add donation without request and no title")
	}
	if d.Unit == "" {
		return Donation{}, errors.Errorf("cannot add donation without request and no unit")
	}
	if d.Qty < 1 {
		return Donation{}, errors.Errorf("cannot add donation with qty:%d (it is < 1)", d.Qty)
	}

	//ok to insert
	id := uuid.New().String()
	if _, err := db.Exec(
		"INSERT INTO `received` SET `id`=?,`location_id`=?,`request_id`=?,`promise_id`=?,`title`=?,`unit`=?,`qty`=?",
		id,
		d.LocationID,
		d.RequestID,
		d.PromiseID,
		d.Title,
		d.Unit,
		d.Qty,
	); err != nil {
		return Donation{}, errors.Wrapf(err, "failed to insert donation")
	}
	d.ID = ID(id)
	return d, nil
}
