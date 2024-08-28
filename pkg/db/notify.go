package db

import (
	"inspection-server/pkg/apis"
	"log"
)

// GetNotify retrieves a notification by its ID from the database.
func GetNotify(notifyID string) (*apis.Notify, error) {
	row := DB.QueryRow("SELECT id, name, app_id, app_secret, webhook_url, secret FROM notify WHERE id = ? LIMIT 1", notifyID)

	var id, name, appID, appSecret, webhookURL, secret string
	notify := apis.NewNotify()
	err := row.Scan(&id, &name, &appID, &appSecret, &webhookURL, &secret)
	if err != nil {
		log.Printf("Error scanning row: %v", err)
		return nil, err
	}

	notify = &apis.Notify{
		ID:         id,
		Name:       name,
		AppID:      appID,
		AppSecret:  appSecret,
		WebhookURL: webhookURL,
		Secret:     secret,
	}

	return notify, nil
}

// ListNotify retrieves all notifications from the database.
func ListNotify() ([]*apis.Notify, error) {
	rows, err := DB.Query("SELECT id, name, app_id, app_secret, webhook_url, secret FROM notify")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return nil, err
	}
	defer rows.Close()

	notifys := apis.NewNotifys()
	for rows.Next() {
		var id, name, appID, appSecret, webhookURL, secret string
		err = rows.Scan(&id, &name, &appID, &appSecret, &webhookURL, &secret)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		notifys = append(notifys, &apis.Notify{
			ID:         id,
			Name:       name,
			AppID:      appID,
			WebhookURL: webhookURL,
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
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback() // Ensure transaction is rolled back if not committed

	stmt, err := tx.Prepare("INSERT INTO notify(id, name, app_id, app_secret, webhook_url, secret) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(notify.ID, notify.Name, notify.AppID, notify.AppSecret, notify.WebhookURL, notify.Secret)
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
	_, err := DB.Exec("UPDATE notify SET name = ?, app_id = ?, app_secret = ?, webhook_url = ?, secret = ? WHERE id = ?", notify.Name, notify.AppID, notify.AppSecret, notify.WebhookURL, notify.Secret, notify.ID)
	if err != nil {
		log.Printf("Error updating notification with ID %s: %v", notify.ID, err)
		return err
	}

	return nil
}

// DeleteNotify removes a notification from the database by its ID.
func DeleteNotify(notifyID string) error {
	_, err := DB.Exec("DELETE FROM notify WHERE id = ?", notifyID)
	if err != nil {
		log.Printf("Error deleting notification with ID %s: %v", notifyID, err)
		return err
	}

	return nil
}
