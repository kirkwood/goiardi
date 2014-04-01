/*
 * Copyright (c) 2013-2014, Jeremy Bingham (<jbingham@gmail.com>)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// General functions for goiardi database connections, if running in that mode.
// Database engine specific functions are in their respective source files.
package data_store

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"fmt"
	"bytes"
	"encoding/gob"
)

var Dbh *sql.DB

// Interface for db handle types that can execute queries
type Dbhandle interface {
	Prepare(query string) (*sql.Stmt, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type ResRow interface {
	Scan(dest ...interface{}) error
}

// Connect to a database with the database name and a map of connection options.
func ConnectDB(dbEngine string, params interface{}) (*sql.DB, error) {
	switch strings.ToLower(dbEngine) {
		case "mysql":
			connectStr, cerr := formatMysqlConStr(params)
			if cerr != nil {
				return nil, cerr
			}
			db, err := sql.Open(strings.ToLower(dbEngine), connectStr)
			if err != nil {
				return nil, err
			}
			if err = db.Ping(); err != nil {
				return nil, err
			}
			return db, nil
		default:
			err := fmt.Errorf("cannot connect to database: unsupported database type %s", dbEngine)
			return nil, err
	}
}

// Encode a slice or map of goiardi object data to save in the database
func EncodeBlob(obj interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	var err error
	defer func() {
		if x := recover(); x != nil {
			err = fmt.Errorf("Something went wrong encoding an object for storing in the database with Gob")
		}
	}()
	err = enc.Encode(&obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode the data encoded with EncodeBlob that was stored in the database so it
// can be loaded back into a goiardi object.
func DecodeBlob(data []byte, obj interface{}) (interface{}, error) {
	dbuf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(dbuf)
	err := dec.Decode(&obj)
	if err != nil {
		return nil, err
	}
	/* Tried to do a pointer to an interface as an argument here, but that
	 * made the compiler pretty unhappy. */
	return obj, nil
}

func CheckForOne(dbhandle Dbhandle, kind string, name string) (int32, error){
	var obj_id int32
	prepStatement := fmt.Sprintf("SELECT id FROM %s WHERE name = ?", kind)
	stmt, err := dbhandle.Prepare(prepStatement)
	defer stmt.Close()
	if err != nil {
		return 0, err
	}
	err = stmt.QueryRow(name).Scan(&obj_id)
	return obj_id, err
}