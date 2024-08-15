package db

import (
	"database/sql"
	"inspection-server/pkg/apis"
	"log"
)

// GetNotify retrieves a notification by its ID from the database.
func GetNotify(notifyID string) (*apis.Notify, error) {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return nil, err
	}
	defer DB.Close()

	row := DB.QueryRow("SELECT id, name, app_id, app_secret FROM notify WHERE id = ? LIMIT 1", notifyID)

	var id, name, appID, appSecret string
	notify := apis.NewNotify()
	err = row.Scan(&id, &name, &appID, &appSecret)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No notify found with ID: %s", notifyID)
			return nil, nil // Return nil, nil to indicate no result
		}
		log.Printf("Error scanning row: %v", err)
		return nil, err
	}

	notify = &apis.Notify{
		ID:        id,
		Name:      name,
		AppID:     appID,
		AppSecret: appSecret,
	}

	return notify, nil
}

// ListNotify retrieves all notifications from the database.
func ListNotify() ([]*apis.Notify, error) {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return nil, err
	}
	defer DB.Close()

	rows, err := DB.Query("SELECT id, name, app_id, app_secret FROM notify")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return nil, err
	}
	defer rows.Close()

	notifys := apis.NewNotifys()
	for rows.Next() {
		var id, name, appID, appSecret string
		err = rows.Scan(&id, &name, &appID, &appSecret)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		notifys = append(notifys, &apis.Notify{
			ID:    id,
			Name:  name,
			AppID: appID,
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return nil, err
	}

	return notifys, nil
}

// CreateNotify inserts a new notification into the database.
func CreateNotify(notify *apis.Notify) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback() // Ensure transaction is rolled back if not committed

	stmt, err := tx.Prepare("INSERT INTO notify(id, name, app_id, app_secret) VALUES(?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(notify.ID, notify.Name, notify.AppID, notify.AppSecret)
	if err != nil {
		log.Printf("Error executing statement: %v", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}

// UpdateNotify updates an existing notification in the database.
func UpdateNotify(notify *apis.Notify) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("UPDATE notify SET name = ?, app_id = ?, app_secret = ? WHERE id = ?", notify.Name, notify.AppID, notify.AppSecret, notify.ID)
	if err != nil {
		log.Printf("Error updating notification with ID %s: %v", notify.ID, err)
		return err
	}

	return nil
}

// DeleteNotify removes a notification from the database by its ID.
func DeleteNotify(notifyID string) error {
	DB, err := GetDB()
	if err != nil {
		log.Printf("Error getting database connection: %v", err)
		return err
	}
	defer DB.Close()

	_, err = DB.Exec("DELETE FROM notify WHERE id = ?", notifyID)
	if err != nil {
		log.Printf("Error deleting notification with ID %s: %v", notifyID, err)
		return err
	}

	return nil
}
