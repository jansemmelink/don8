package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/gorilla/mux"
	"github.com/jansemmelink/don8/db"
	"github.com/jansemmelink/events/email"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func main() {
	addrPtr := flag.String("addr", ":3500", "HTTP Server address")
	flag.Parse()
	r := mux.NewRouter()
	//auth
	r.HandleFunc("/register", hdlr(register, authNone)).Methods(http.MethodPost)
	r.HandleFunc("/activate", hdlr(activate, authNone)).Methods(http.MethodPost)
	r.HandleFunc("/reset", hdlr(reset, authNone)).Methods(http.MethodPost)
	r.HandleFunc("/login", hdlr(login, authNone)).Methods(http.MethodPost)
	r.HandleFunc("/logout", hdlr(logout, authSession)).Methods(http.MethodPost)
	//groups
	r.HandleFunc("/group/{id}", hdlr(getGroup, authNone)).Methods(http.MethodGet) //no auth: given the uuid, you can see a group, e.g. from an invite
	r.HandleFunc("/group/{id}", hdlr(updGroup, authSession)).Methods(http.MethodPut)
	r.HandleFunc("/groups", hdlr(addGroup, authSession)).Methods(http.MethodPost)
	r.HandleFunc("/groups", hdlr(listGroups, authSession)).Methods(http.MethodGet)
	//requests
	r.HandleFunc("/request/{id}", hdlr(getRequest, authSession)).Methods(http.MethodGet)
	r.HandleFunc("/request/{id}", hdlr(updRequest, authSession)).Methods(http.MethodPut)
	r.HandleFunc("/requests", hdlr(addRequest, authSession)).Methods(http.MethodPost)
	r.HandleFunc("/requests", hdlr(listRequests, authSession)).Methods(http.MethodGet)

	http.Handle("/", Log(CORS(r)))

	log.Infof("Listening on %s ...", *addrPtr)
	http.ListenAndServe(*addrPtr, nil)
}

func Log(h http.Handler) http.Handler {
	//todo...
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func CORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		fmt.Printf("HTTP %s %s (origin:%s)\n", r.Method, r.URL.Path, origin)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,GET,POST,PUT,DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, content-type, Accept, Authorization, Don8-Auth-Sid")
			//w.WriteHeader(http.StatusNoContent)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

type authRequirment int

const (
	authNone authRequirment = iota
	// authTpw
	authSession
)

// type CtxAuthUser struct{}
type CtxAuthSession struct{}
type CtxParams struct{}

func hdlr(fnc interface{}, auth authRequirment) http.HandlerFunc {
	fncType := reflect.TypeOf(fnc)
	fncValue := reflect.ValueOf(fnc)
	var reqType reflect.Type
	if fncType.NumIn() > 1 {
		reqType = fncType.In(1)
	}

	type ErrorResponse struct {
		Error string `json:"error"`
	}

	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		ctx := context.Background()
		var status int = http.StatusInternalServerError
		var err error
		var res interface{}
		defer func() {
			if err != nil {
				//log full error but in response, only log the base error
				log.Errorf("Failed: %+v\n", err)
				for {
					if baseErr, ok := err.(errors.IError); ok {
						if baseErr.Code() > 0 {
							status = baseErr.Code()
						}
						if baseErr.Parent() != nil {
							err = baseErr.Parent()
						} else {
							break
						}
					}
				}
				res = ErrorResponse{Error: fmt.Sprintf("%+s", err)}
			}
			httpRes.Header().Set("Content-Type", "application/json")
			httpRes.WriteHeader(status)
			if res != nil {
				jsonRes, _ := json.Marshal(res)
				httpRes.Write(jsonRes)
				fmt.Printf("-> %s\n", jsonRes)
			}
		}()

		params := newParams()
		for n, v := range httpReq.URL.Query() {
			params = params.With(n, strings.Join(v, ","))
		}
		vars := mux.Vars(httpReq)
		for n, v := range vars {
			params = params.With(n, v)
		}
		ctx = context.WithValue(ctx, CtxParams{}, params)

		switch auth {
		case authNone: //do nothing

		case authSession: //get session id for logged in user
			//get user details if required
			authSidHeader := "Don8-Auth-Sid"
			sid := httpReq.Header.Get(authSidHeader)
			if sid == "" {
				err = errors.Errorc(http.StatusUnauthorized,
					fmt.Sprintf("missing header %s", authSidHeader))
				return
			}

			var s db.Session
			s, err = db.GetSession(db.ID(sid))
			if err != nil {
				err = errors.Errorc(http.StatusUnauthorized, fmt.Sprintf("invalid %s header value: %s", authSidHeader, err))
				return
			}
			log.Debugf("HTTP %s %s Session:%+v User:%+v", httpReq.Method, httpReq.URL.Path, s, *s.User)
			ctx = context.WithValue(ctx, CtxAuthSession{}, s)
		default:
			err = errors.Errorc(http.StatusInternalServerError, "invalid auth specification")
			return
		} //switch(auth)

		//prepare fnc arguments
		args := []reflect.Value{reflect.ValueOf(ctx)}

		if fncType.NumIn() > 1 {
			ct := httpReq.Header.Get("Content-Type")
			if ct != "" && ct != "application/json" {
				err = errors.Errorc(http.StatusBadRequest, fmt.Sprintf("invalid Content-Type: %+s, expecting application/json", ct))
				return
			}

			reqValuePtr := reflect.New(reqType)
			if err = json.NewDecoder(httpReq.Body).Decode(reqValuePtr.Interface()); err != nil {
				err = errors.Errorc(http.StatusBadRequest, fmt.Sprintf("cannot parse JSON body: %+s", err))
				return
			}

			if validator, ok := reqValuePtr.Interface().(Validator); ok {
				if err = validator.Validate(); err != nil {
					log.Errorf("Invalid (%T): %+v:  %+v", reqValuePtr.Interface(), err, reqValuePtr.Interface())
					err = errors.Errorc(http.StatusBadRequest, err.Error())
					return
				}
				log.Debugf("Validated (%T) %+v", reqValuePtr.Interface(), reqValuePtr.Interface())
			} else {
				log.Debugf("Not Validating (%T) %+v", reqValuePtr.Interface(), reqValuePtr.Interface())
			}
			args = append(args, reqValuePtr.Elem())
		}

		results := fncValue.Call(args)

		errValue := results[len(results)-1] //last result is error
		if !errValue.IsNil() {
			err = errors.Wrapf(errValue.Interface().(error), "handler failed")
			return
		}

		if fncType.NumOut() > 1 {
			if results[0].IsValid() {
				if results[0].Type().Kind() == reflect.Ptr && !results[0].IsNil() {
					res = results[0].Elem().Interface() //dereference the pointer
				} else {
					res = results[0].Interface()
				}
			}
		}

		//success: set status code
		switch httpReq.Method {
		case http.MethodPost, http.MethodPut:
			status = http.StatusAccepted
		case http.MethodGet:
			status = http.StatusOK
		case http.MethodDelete:
			status = http.StatusNoContent
			res = nil
		}
	}
}

type RegisterRequest struct {
	db.User
	ActivateLink string `json:"activate_link" doc:"Activation link to send to user in email"`
}

func (req RegisterRequest) Validate() error {
	if err := req.User.Validate(); err != nil {
		return err
	}
	if req.ActivateLink == "" {
		return errors.Errorf("missing activate_link")
	}
	return nil
}

func register(ctx context.Context, req RegisterRequest) (db.User, error) {
	user, err := db.AddUser(req.User)
	if err != nil {
		return db.User{}, err
	}
	//send email to user
	if err := email.Send(
		email.Message{
			From:        email.Email{Addr: "accounts@don8.com", Name: "Don8 Accounts"},
			To:          []email.Email{{Addr: user.Email, Name: user.Name}},
			Subject:     "Don8 Account",
			ContentType: "text/html",
			Content: `
			<h1>New Don8 Account</h1>
			<p>Your email address was registered at don8.</p>
			<p>If you did not register your address, ignore this email and we will forget your address.</p>
			<p>If you did register, click <a href="` + req.ActivateLink + `/` + string(*user.Tpw) + `">here</a> to activate your account.</p>
			`,
		},
	); err != nil {
		return db.User{}, errors.Errorc(http.StatusInternalServerError, "failed to send activation link to your email address")
	}
	return user, nil
}

func activate(ctx context.Context, req db.ActivateRequest) (db.Session, error) {
	return db.Activate(req)
}

type ResetRequest struct {
	db.ResetRequest
	ResetLink string `json:"reset_link"`
}

func (req ResetRequest) Validate() error {
	if err := req.ResetRequest.Validate(); err != nil {
		return err
	}
	if req.ResetLink == "" {
		return errors.Errorc(http.StatusBadRequest, "missing reset_link")
	}
	return nil
}

func reset(ctx context.Context, req ResetRequest) error {
	user, err := db.Reset(req.ResetRequest)
	if err != nil {
		return err
	}
	//send email to user
	if err := email.Send(
		email.Message{
			From:        email.Email{Addr: "accounts@don8.com", Name: "Don8 Accounts"},
			To:          []email.Email{{Addr: user.Email, Name: user.Name}},
			Subject:     "Don8 Account",
			ContentType: "text/html",
			Content: `
			<h1>Password Reset</h1>
			<p>We received a request to reset your password.</p>
			<p>If you did not make the request, delete this email and your current password remains as it is.</p>
			<p>To set a new password, click <a href="` + req.ResetLink + `/` + string(*user.Tpw) + `">here</a>.</p>
			`,
		},
	); err != nil {
		return errors.Errorc(http.StatusInternalServerError, "failed to send password reset link to your email address")
	}
	return nil
}

func login(ctx context.Context, req db.LoginRequest) (db.Session, error) {
	return db.Login(req)
}

func logout(ctx context.Context) error {
	s := ctx.Value(CtxAuthSession{}).(db.Session)
	return db.Logout(s.ID)
}

func addGroup(ctx context.Context, req db.NewGroup) (db.Group, error) {
	s := ctx.Value(CtxAuthSession{}).(db.Session)
	g, err := db.AddGroup(*s.User, req)
	if err != nil {
		return db.Group{}, err
	}
	return g, nil
}

func listGroups(ctx context.Context) ([]db.MyGroup, error) {
	s := ctx.Value(CtxAuthSession{}).(db.Session)
	params := ctx.Value(CtxParams{}).(params)
	filter := params.String("filter", "")
	return db.MyGroups(*s.User, filter, nil, nil)
}

//getGroup gives the app a good view of the group, including parent description and immediate child list
func getGroup(ctx context.Context) (db.FullGroup, error) {
	params := ctx.Value(CtxParams{}).(params)
	log.Infof("params: %+v", params)
	id := params.String("id", "")
	if id == "" {
		return db.FullGroup{}, errors.Errorc(http.StatusBadRequest, "missing URL param id")
	}
	fg, err := db.GetFullGroup(db.ID(id))
	if err != nil {
		return db.FullGroup{}, errors.Errorc(http.StatusNotFound, "unknown group")
	}

	//load requests (can also filter on params)
	if fg.Requests, err = listRequests(ctx); err != nil {
		log.Errorf("failed to load group requests")
		fg.Requests = nil
	}
	return fg, nil
}

func updGroup(ctx context.Context, req db.UpdGroupRequest) (db.FullGroup, error) {
	//todo: check permission on this group
	if err := db.UpdGroup(req); err != nil {
		return db.FullGroup{}, errors.Errorf("failed to update group")
	}
	fg, err := db.GetFullGroup(req.ID)
	if err != nil {
		log.Errorf("failed to get group after update: %+v", err)
		return db.FullGroup{}, errors.Errorf("failed to get group after update")
	}
	return fg, nil
}

func addRequest(ctx context.Context, req db.Request) (db.Request, error) {
	return db.AddRequest(req)
}

func listRequests(ctx context.Context) ([]db.Request, error) {
	//s := ctx.Value(CtxAuthSession{}).(db.Session)
	//todo: check must be member of group
	params := ctx.Value(CtxParams{}).(params)
	groupID := params.String("id", "")
	if groupID == "" {
		return nil, errors.Errorc(http.StatusBadRequest, "missing param id")
	}
	filter := params.String("filter", "")
	tags := db.TagsFromString(params.String("tags", ""))
	limit := params.Int("limit", 10, 1, 100)
	requests, err := db.FindRequests(db.ID(groupID), filter, tags, limit)
	if err != nil {
		return nil, err
	}
	for i, r := range requests {
		if r.Tags != nil {
			tags := db.TagsCSV(db.TagsFromString(*r.Tags))
			if tags != "" {
				requests[i].Tags = &tags
			} else {
				requests[i].Tags = nil
			}
		}
	}
	return requests, nil
}

//getRequest including group title and summary of receives and promises etc...
func getRequest(ctx context.Context) (db.FullRequest, error) {
	params := ctx.Value(CtxParams{}).(params)
	log.Infof("params: %+v", params)
	id := params.String("id", "")
	if id == "" {
		return db.FullRequest{}, errors.Errorc(http.StatusBadRequest, "missing URL param id")
	}
	fr, err := db.GetFullRequest(db.ID(id))
	if err != nil {
		log.Errorf("failed to get full request(%s): %+v", id, err)
		return db.FullRequest{}, errors.Errorc(http.StatusNotFound, "unknown request")
	}

	//present tags as CSV in the API
	if fr.Tags != nil {
		tags := db.TagsCSV(db.TagsFromString(*fr.Tags))
		if tags != "" {
			*fr.Tags = tags
		} else {
			fr.Tags = nil
		}
	}
	return fr, nil
}

func updRequest(ctx context.Context, req db.UpdRequestRequest) (db.FullRequest, error) {
	//todo: check permission on this group
	if err := db.UpdRequest(req); err != nil {
		return db.FullRequest{}, errors.Errorf("failed to update request")
	}
	fr, err := db.GetFullRequest(req.ID)
	if err != nil {
		log.Errorf("failed to get request after update: %+v", err)
		return db.FullRequest{}, errors.Errorf("failed to get request after update")
	}
	//present tags as CSV in the API
	if fr.Tags != nil {
		tags := db.TagsCSV(db.TagsFromString(*fr.Tags))
		if tags != "" {
			*fr.Tags = tags
		} else {
			fr.Tags = nil
		}
	}
	return fr, nil
}

type Validator interface {
	Validate() error
}

type params struct {
	value map[string]string
}

func newParams() params {
	return params{
		value: map[string]string{},
	}
}

func (p params) With(n, v string) params {
	p.value[n] = v
	return p
}

func (p params) String(n, defaultValue string) string {
	if s, ok := p.value[n]; !ok {
		return defaultValue
	} else {
		return s
	}
}

func (p params) Int(n string, defaultValue, minValue, maxValue int) int {
	s, ok := p.value[n]
	if !ok {
		return defaultValue
	}
	i64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultValue
	}
	if int(i64) < minValue {
		return minValue
	}
	if int(i64) > maxValue {
		return maxValue
	}
	return int(i64)
}
