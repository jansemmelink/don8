package main

import (
	"context"
	"encoding/json"

	"github.com/go-msvc/errors"
	"github.com/go-redis/redis/v8"
	"github.com/jansemmelink/don8/db"
	"github.com/jansemmelink/events/email"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

//subscribes to redis queue and send group invites
func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	qName := "Q:group-invitations"
	log.Infof("Subscribing on queue(%s) ...", qName)

	ctx := context.Background()
	redisPubSub := redisClient.Subscribe(ctx, qName)
	defer func() {
		log.Infof("Unsubscribing from qName(%s).", qName)
		redisPubSub.Close()
	}()

	log.Infof("Waiting for messages on queue(%s) ...", qName)
	for {
		if err := process(redisPubSub); err != nil {
			log.Errorf("process failed: %+v", err)
		}
	}
} //main()

func process(r *redis.PubSub) error {
	ctx := context.Background()
	//if supply context with timeout like this, we get a redis panic when the timeout
	//fires with no messages, so we have to catch the panic and recover
	//but redis still prints a message in that case, so rather no give a timeout at all
	// defer func() {
	// 	r := recover()
	// 	if r != nil {
	// 		log.Errorf("Got R: %+v", r)
	// 	}
	// }()
	// ctx, cancelFunc := context.WithTimeout(ctx, time.Second)
	// defer func() {
	// 	log.Debugf("Cancelling context")
	// 	cancelFunc()
	// }()
	// panic prints this then terminates:
	//redis: 2022/08/11 08:59:01 pubsub.go:159: redis: discarding bad PubSub connection: read tcp [::1]:49500->[::1]:6379: i/o timeout

	msg, err := r.ReceiveMessage(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to receive")
	}

	var req db.Invitation
	if err := json.Unmarshal([]byte(msg.Payload), &req); err != nil {
		return errors.Wrapf(err, "%s: invalid JSON", msg.Channel)
	}
	if err := req.Validate(); err != nil {
		return errors.Wrapf(err, "%s: invalid request", msg.Channel)
	}

	//group must exist
	group, err := db.GetGroup(req.GroupID)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to get group(id:%s)", msg.Channel, req.GroupID)
	}

	//do not send if already joined
	member, err := db.GetMemberByEmail(req.GroupID, req.Email)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to get member(email:%s)", msg.Channel, req.Email)
	}
	if member != nil {
		return errors.Wrapf(err, "%s: group(id:%s).member(id:%s,email:%s) already exists", msg.Channel, req.GroupID, member.ID, req.Email)
	}

	//do not invite if already invited
	inv, err := db.GetInvitationByEmail(req.GroupID, req.Email)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to get invitation(email:%s)", msg.Channel, req.Email)
	}
	if inv != nil {
		//todo: add option to request to resent existing invites...
		return errors.Wrapf(err, "%s: group(id:%s).invitation(id:%s,email:%s,cre:%s,upd:%s) already exists", msg.Channel, req.GroupID, inv.ID, req.Email, inv.TimeCreated, inv.TimeUpdated)
	}

	//not yet member, nor invited, so create new invitation
	inv, err = db.AddInvitation(req)
	if err != nil {
		return errors.Wrapf(err, "%s: failed to create invitation", msg.Channel)
	}
	log.Debugf("Created invitation: %+v", inv)

	//new invitation status is "sent", but if send fails,
	//we delete if with this defer function
	sent := false
	defer func() {
		if !sent {
			if err := db.DelInviation(inv.ID); err != nil {
				log.Errorf("failed to delete invitation(id:%s) after fail to send: %+v", inv.ID, err)
			} else {
				log.Errorf("deleted invitation(id:%s) after fail to send", inv.ID)
			}
		}
	}()

	//send invitation by email
	//todo: make message configurable or load a template etc...
	if err := email.Send(email.Message{
		From:        email.Email{Addr: "invitations@don8.com"},
		To:          []email.Email{{Addr: req.Email}},
		Subject:     "Invitation to join " + group.Title,
		ContentType: "text/html",
		Content: "<h1>Group Invitation</h1>" +
			"<p>You are invited to join " + group.Title + ".</p>" +
			"<button onClick='/invitation/" + string(inv.ID) + "'>Join</button>" +
			"<p>If you do not want to join immediately, you can just ignore this and join later.</p>" +
			"<p>Click <a href='/invitation/" + string(inv.ID) + "/block'>here</a> to block this group permanently from sending you more invites and messages, click the next button...</p>",
	}); err != nil {
		return errors.Wrapf(err, "%s: failed to send invitation", msg.Channel)
	}
	sent = true //set this not to delete in the defer func above
	log.Debugf("Sent invitation")
	return nil
}
