package db

import (
	"database/sql"
	"fmt"
	"inspection-server/pkg/apis"
	"log"
)

func GetNotify(notifyID string) (*apis.Notify, error) {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, name, app_id, app_secret FROM notify WHERE id = ? LIMIT 1", notifyID)

	var id, name, appID, appSecret string
	notify := apis.NewNotify()
	err = row.Scan(&id, &name, &appID, &appSecret)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("没有找到匹配的数据")
		} else {
			return nil, err
		}
	} else {
		notify = &apis.Notify{
			ID:    id,
			Name:  name,
			AppID: appID,
		}
	}

	return notify, nil
}

func ListNotify() ([]*apis.Notify, error) {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return nil, err
	}
	defer DB.Close()

	rows, err := DB.Query("SELECT id, name, app_id, app_secret FROM notify")
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	notifys := apis.NewNotifys()
	for rows.Next() {
		var id, name, appID, appSecret string
		err = rows.Scan(&id, &name, &appID, &appSecret)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("没有找到匹配的数据")
			} else {
				return nil, err
			}
		}

		notifys = append(notifys, &apis.Notify{
			ID:    id,
			Name:  name,
			AppID: appID,
		})
	}

	return notifys, nil
}

func CreateNotify(notify *apis.Notify) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO notify(id, name, app_id, app_secret) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(notify.ID, notify.Name, notify.AppID, notify.AppSecret)
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()

	return nil
}

func UpdateNotify(notify *apis.Notify) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("UPDATE notify SET name = ?, app_id = ?, app_secret = ? WHERE id = ?", notify.Name, notify.AppID, notify.AppSecret, notify.ID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteNotify(notifyID string) error {
	DB, err := sql.Open(sqliteDriver, sqliteName)
	if err != nil {
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM notify WHERE id = ?", notifyID)
	if err != nil {
		return err
	}

	return nil
}
