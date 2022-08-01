package db

import (
	"fmt"
	"time"
)

// type LogRecord struct {
// 	ID        ID          `db:"id"`
// 	Table     string      `db:"table"`
// 	Timestamp SqlTime     `db:"timestamp"`
// 	UserID    ID          `db:"user_id"`
// 	Action    string      `db:"action"`
// 	Values    interface{} `db:"values"`
// }

func Log(id ID, table string, user *User, action string, values interface{}) {
	ts := time.Now()
	if _, err := db.Exec("INSERT INTO `logs` SET `id`=?,`table`=?,`timestamp`=?,`user_id`=?,`action`=?,`values`=?",
		id,
		table,
		SqlTime(ts),
		user.ID,
		action,
		values,
	); err != nil {
		panic(fmt.Sprintf("failed to add log: %+v", err))
	}
}
