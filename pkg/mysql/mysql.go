package mysql

import (
	"database/sql"
	"fmt"
	"github.com/vinkdong/gox/log"
)

type Mysql struct {
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	Username string `json:"name"`
	Password string `json:"password"`
}

func (m *Mysql) Connect() (db *sql.DB, err error) {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8&timeout=3s", m.Username, m.Password, m.Host, m.Port)
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Errorf("Failed to connect mysql: %s", err)
		return
	}
	err = db.Ping()
	if err != nil {
		log.Errorf("Failed to ping mysql: %s", err)
	}
	return
}
