// Description: Interacts with the sqlite3 database
//				by running query's and returning results.
//
// Author: Chamod

package main

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Database pointer
type Database struct {
	*sqlx.DB
}

// Link to a PC
type Link struct {
	LinkID     string `json:"link,omitempty"`
	Permission string `json:"permission,omitempty"`
	PCID       int    `json:"pcId,omitempty"`
}

// Links to send back upon creation
type Links struct {
	EditID string `json:"editId,omitempty" db:"link_id"`
	ViewID string `json:"viewId,omitempty" db:"link_id"`
}

// PCList to store pc name, info and view link
type PCList struct {
	Name   string `json:"name,omitempty" db:"name"`
	Info   string `json:"info,omitempty" db:"info"`
	ViewID string `json:"viewId,omitempty" db:"link_id"`
}

// PC Info
type PC struct {
	PCID   int     `json:"pcId,omitempty" db:"pc_id"`
	Name   string  `json:"name,omitempty" db:"name"`
	Info   string  `json:"info,omitempty" db:"info"`
	Parts  []Part  `json:"parts"`
	Images []Image `json:"images"`
	Links  Links   `json:"links"`
}

// Image contains info of a PC image
type Image struct {
	ImageID int    `json:"imageId,omitempty" db:"image_id"`
	Link    string `json:"link,omitempty" db:"link"`
	PCID    int    `json:"pcId,omitempty" db:"pc_id"`
}

// Part of a PC
type Part struct {
	PartID int    `json:"partId,omitempty" db:"part_id"`
	Type   string `json:"type,omitempty" db:"type"`
	Brand  string `json:"brand,omitempty" db:"brand"`
	Model  string `json:"model,omitempty" db:"model"`
	Qty    int    `json:"qty,omitempty" db:"qty"`
	PCID   int    `json:"pcId,omitempty" db:"pc_id"`
}

// ConnectToDB connects to the database
func ConnectToDB() (*Database, error) {
	db, err := sqlx.Connect("sqlite3", "file:pcshowdown.db?_fk=true")
	return &Database{db}, err
}

// AddPC adds a PC to the database
// returns links and error
func (database *Database) AddPC(pc PC) (Links, error) {
	var links Links

	// Start transaction due to multiple items
	tx, err := database.Beginx()

	if err != nil {
		return links, err
	}

	// add pc  to pc table
	query := `INSERT INTO pc (name, info) VALUES (?, ?);`
	result, err := tx.Exec(query, pc.Name, pc.Info)
	if err != nil {
		tx.Rollback()

		return links, err
	}

	// get the pc id
	pcID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return links, err
	}

	// add parts to parts table
	for _, part := range pc.Parts {
		query = `INSERT INTO part (type, brand, model, qty, pc_id) VALUES (?, ?, ?, ?, ?);`
		_, err = tx.Exec(query, part.Type, part.Brand, part.Model, part.Qty, pcID)
		if err != nil {
			tx.Rollback()
			return links, err
		}
	}

	// add images to image table
	for _, image := range pc.Images {
		query = `INSERT INTO image (pc_id, link) VALUES (?, ?);`
		_, err = tx.Exec(query, pcID, image.Link)
		if err != nil {
			tx.Rollback()
			return links, err
		}
	}

	// Make and return links
	links, err = database.createLinks(tx, pcID)
	if err != nil {
		tx.Rollback()
		return links, err
	}

	tx.Commit()
	return links, nil
}

// createLinks creates a view and edit link
// returns a Links, edit and view link and error
func (database *Database) createLinks(tx *sqlx.Tx, pcID int64) (Links, error) {
	var links Links

	// add edit link
	query := `INSERT INTO link (pc_id, permission) VALUES (?, ?);`
	_, err := tx.Exec(query, pcID, "edit")
	if err != nil {
		return links, err
	}

	// get edit link id
	linkQuery := `SELECT link_id FROM link WHERE pc_id = ? AND permission = ?;`
	err = tx.Get(&(links.EditID), linkQuery, pcID, "edit")
	if err != nil {
		return links, err
	}
	// add view link
	_, err = tx.Exec(query, pcID, "view")
	if err != nil {
		return links, err
	}
	// get view link id
	err = tx.Get(&(links.ViewID), linkQuery, pcID, "view")
	if err != nil {
		return links, err
	}

	return links, nil
}

// GetLinks gets links permitted to the given link
func (database *Database) GetLinks(linkID string) (Links, error) {
	var links Links
	var permission string

	// check for permission
	query := `SELECT permission FROM link WHERE link_id = ?;`
	err := database.Get(&permission, query, linkID)
	if err != nil {
		return links, err
	}

	// if edit permission, get view id and return edit and view id
	if permission == "edit" {
		query = `SELECT link_id FROM link
				WHERE permission = "view"
				AND pc_id IN (SELECT pc_id FROM link
				WHERE link_id = ? LIMIT 1);`
		err := database.Get((&links.ViewID), query, linkID)
		if err != nil {
			return links, err
		}
		links.EditID = linkID
		return links, nil
	}

	// else just view id
	links.ViewID = linkID

	return links, nil
}

// GetPCS gets all the PCs at a range
// Returns a list of pcs and error
func (database *Database) GetPCS(oldID int, limit int) ([]PCList, error) {
	var pcs []PCList

	query := `SELECT name, info, link_id
			FROM pc INNER JOIN link ON link.pc_id = pc.pc_id
			WHERE permission = "view" AND pc.pc_id > ?
			ORDER BY pc.pc_id ASC
			LIMIT ?;`

	err := database.Select(&pcs, query, oldID, limit)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	return pcs, nil
}

// GetPC gets a pc belong to given link id
// Returns a pc and an error
func (database *Database) GetPC(linkID string) (PC, error) {
	var pc PC

	// get the pc name and info
	query := `SELECT name, info FROM pc
			 WHERE pc_id IN (SELECT pc_id
			 FROM link WHERE link_id = ?);`

	err := database.Get(&pc, query, linkID)

	if err != nil {
		return pc, err
	}

	// get all the parts
	query = `SELECT type, model, brand, qty FROM part
			 WHERE pc_id IN (SELECT pc_id
			 FROM link WHERE link_id = ?);`

	err = database.Select(&(pc.Parts), query, linkID)

	if err != nil {
		return pc, err
	}

	// get all the images
	query = `SELECT link FROM image
			 WHERE pc_id IN (SELECT pc_id
			 FROM link WHERE link_id = ?);`

	err = database.Select(&(pc.Images), query, linkID)

	if err != nil {
		return pc, err
	}

	// get links
	pc.Links, err = database.GetLinks(linkID)

	if err != nil {
		return pc, err
	}

	return pc, nil
}

// UpdatePC updates a PC with new info
// returns updated pc and an error
func (database *Database) UpdatePC(linkID string, updatedPC PC) (PC, error) {
	// get the pc id
	query := `SELECT pc_id
			 FROM link WHERE link_id = ?
			 AND permission = "edit"; `

	err := database.Get(&(updatedPC.PCID), query, linkID)
	if err != nil {
		return updatedPC, err
	}

	// begin transaction since multiple updates
	tx, err := database.Beginx()
	if err != nil {
		return updatedPC, err
	}

	// update the PC name and info
	query = `UPDATE pc SET name = ?, info = ?
			 WHERE pc_id = ?;`
	_, err = tx.Exec(query, updatedPC.Name, updatedPC.Info, updatedPC.PCID)
	if err != nil {
		tx.Rollback()
		return updatedPC, err
	}

	// delete the previous parts
	query = `DELETE FROM part WHERE pc_id = ?;`
	_, err = tx.Exec(query, updatedPC.PCID)
	if err != nil {
		tx.Rollback()
		return updatedPC, err
	}

	// add new parts
	for _, part := range updatedPC.Parts {
		query = `INSERT INTO part
				 (type, brand, model, qty, pc_id)
				 VALUES (?, ?, ?, ?, ?);`
		_, err = tx.Exec(query, part.Type, part.Brand, part.Model, part.Qty, updatedPC.PCID)
		if err != nil {
			tx.Rollback()
			return updatedPC, err
		}
	}

	// delete the previous images
	query = `DELETE FROM image WHERE pc_id = ?;`
	_, err = tx.Exec(query, updatedPC.PCID)
	if err != nil {
		tx.Rollback()
		return updatedPC, err
	}

	// add new images
	for _, image := range updatedPC.Images {
		query = `INSERT INTO image
				(link, pc_id) VALUES (?, ?);`
		_, err = tx.Exec(query, image.Link, updatedPC.PCID)
		if err != nil {
			tx.Rollback()
			return updatedPC, err
		}
	}

	// commit the updates values
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return updatedPC, err
	}

	// get the links
	links, err := database.GetLinks(linkID)
	if err != nil {
		return updatedPC, err
	}

	// set the links
	updatedPC.Links = links

	return updatedPC, nil
}

// DeletePC deletes pc for given link from db
// Returns error
func (database *Database) DeletePC(linkID string) error {

	query := `DELETE FROM pc WHERE pc_id IN
			 (SELECT pc_id FROM link WHERE
			 link_id = ? AND permission = "edit");`

	_, err := database.Exec(query, linkID)
	if err != nil {
		return err
	}
	return nil
}
