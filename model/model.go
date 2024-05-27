package model

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
	"time"
)

// ErrInvalidUser represents an error that occurs when an invalid user is encountered.
var (
	ErrInvalidUser = fmt.Errorf("invalid user")
)

// CreatedAt represents the timestamp of when an entity or object was created.
// It is used to track the creation time of various entities.
type CreatedAt struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// UpdatedAt represents the timestamp when an entity was last updated.
// It is used to keep track of the latest modification of the entity.
type UpdatedAt struct {
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// DeletedAt struct represents the soft delete functionality in GORM.
// It embeds the `gorm.DeletedAt` struct, which provides the necessary fields for soft deletion.
type DeletedAt struct {
	Deleted   bool       `gorm:"column:deleted;index:deleted" json:"deleted"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

// IsDeleted returns true if the Deleted field of the DeletedAt object is set to true, indicating that the object has been deleted. Otherwise, it returns false.
func (o *DeletedAt) IsDeleted() bool {
	return o.Deleted
}

// Delete updates the `Deleted` and `DeletedAt` fields of the `DeletedAt` object.
// If `v` is true, `Deleted` is set to `v` and `DeletedAt` is set to the current time.
// If `v` is false, `Deleted` is set to `v` and `DeletedAt` is set to `nil`.
func (o *DeletedAt) Delete(v bool) {
	o.Deleted = v
	if v {
		var now = time.Now()
		o.DeletedAt = &now
	} else {
		o.DeletedAt = nil
	}
}

// ArchivedAt represents an archived status with an optional archived timestamp.
// It is used to indicate whether an entity is archived and provides the option to store the archived timestamp.
//
// Fields:
//   - Archived: a boolean field indicating whether the entity is archived or not.
//   - ArchivedAt: a pointer to a time.Time field representing the timestamp when the entity was archived.
//
// Example usage:
//
//	a := &ArchivedAt{}
//
//	// Check if the entity is archived
//	if a.IsArchived() {
//	  // Entity is archived
//	} else {
//	  // Entity is not archived
//	}
//
//	// Archive the entity
//	a.Archive(true)
//
//	// Access the archived timestamp if available
//	if a.ArchivedAt != nil {
//	  // Do something with the archived timestamp
//	}
//
//	// Remove the archived timestamp
//	a.ArchivedAt = nil
type ArchivedAt struct {
	Archived   bool       `gorm:"column:archived_at;index:archived" json:"archived"`
	ArchivedAt *time.Time `gorm:"column:archived_at" json:"archived_at"`
}

// IsArchived returns a boolean value indicating whether the object is archived or not.
func (o *ArchivedAt) IsArchived() bool {
	return o.Archived
}

// Archive sets the value of Archived attribute to the specified boolean value.
// If v is true, Archived is set to true and ArchivedAt is set to the current time.
// If v is false, Archived is set to false and ArchivedAt is set to nil.
// This method is used to mark an object as archived or present.
// Example:
//
//	o.Archive(true) // Marks the object as archived and sets the archive time.
//	o.Archive(false) // Marks the object as not archived and clears the archive time.
func (o *ArchivedAt) Archive(v bool) {
	o.Archived = v
	if v {
		var now = time.Now()
		o.ArchivedAt = &now
	} else {
		o.ArchivedAt = nil
	}
}

// LastEdit represents the last edit information of an entity or object.
// It stores the Identifier of the user who performed the last edit, as well as a reference to the User object itself.
type LastEdit struct {
	LastEditByUUID *string `gorm:"column:last_edit_by;fk:users.uuid;size:36" json:"last_edit_by"`
	LastEdit       *User   `gorm:"-" json:"last_edit,omitempty"`
}

// GetLastEdit returns the last edit user of the LastEdit object.
// It checks if LastEditByUUID is not empty and LastEdit is nil, then it queries the database to populate LastEdit.
// If the query fails, it returns an error.
// If LastEdit is still nil after the query, it returns ErrInvalidUser.
// Otherwise, it returns LastEdit and nil.
// The LastEditByUUID is used to query the database using the db.Where method, which returns a gorm.DB.
// The db.Where method is a query builder that filters the result based on a condition.
// The db.Where method is using the "uuid = ?" condition with the LastEditByUUID as the value.
// The db.Where method is encapsulated in the function Where, which takes a query and arguments and returns a *gorm.DB.
// The ErrInvalidUser is a predefined error indicating an invalid user.
//
// Example usage:
// user, err := lastEdit.GetLastEdit()
//
//	if err != nil {
//	    log.Fatalf("Failed to get last edit user: %v", err)
//	}
//
// fmt.Printf("Last edit user: %v", user)
//
// Declaration:
//
//	type LastEdit struct {
//	    LastEditByUUID *string `gorm:"column:last_edit_by;fk:users.uuid;size:36" json:"last_edit_by"`
//	    LastEdit       *User   `gorm:"-" json:"last_edit,omitempty"`
//	}
//
//	func (o *LastEdit) GetLastEdit() (*User, error) {
//	    ...
//	}
//
//	func (o *LastEdit) SetLastEdit(u *User) {
//	    ...
//	}
//
//	type User struct {
//	    Identifier
//	    FirstName string `gorm:"column:first_name;size:255" json:"first_name"`
//	    LastName  string `gorm:"column:last_name;size:255" json:"last_name"`
//	    Email     string `gorm:"column:email;size:255;unique" json:"email"`
//	}
//
//	func (User) TableName() string {
//	    return "users"
//	}
//
// func Where(query any, args ...
func (o *LastEdit) GetLastEdit() (*User, error) {
	if o.LastEditByUUID != nil && *o.LastEditByUUID != "" && o.LastEdit == nil {
		err := db.Where("uuid = ?", o.LastEditByUUID).Take(o.LastEdit).Error
		if err != nil {
			return nil, err
		}
	}
	if o.LastEdit == nil {
		return nil, ErrInvalidUser
	}
	return o.LastEdit, nil
}

// SetLastEdit sets the LastEdit and LastEditByUUID fields of the LastEdit object.
func (o *LastEdit) SetLastEdit(u *User) {
	o.LastEdit = u
	o.LastEditByUUID = &u.UUID
}

// User represents a user entity.
//
// It contains the following fields:
// - Identifier: the universally unique identifier of the user.
// - FirstName: the first name of the user.
// - LastName: the last name of the user.
// - Email: the email address of the user.
//
// Example usage:
//
// Create a new user:
//
//	u := User{
//	    FirstName: "John",
//	    LastName:  "Doe",
//	    Email:     "johndoe@example.com",
//	}
//
// Update an existing user:
// u.FirstName = "Jane"
// u.LastName = "Smith"
//
// Retrieve a user by Identifier:
// var user User
// db.Where("uuid = ?", uuid).First(&user)
//
// Delete a user:
// db.Delete(&user)
//
// Get the full name of a user:
// fullName := user.FirstName + " " + user.LastName
type User struct {
	UUID      string `gorm:"column:uuid;primaryKey;size:36" json:"uuid"`
	FirstName string `gorm:"column:first_name;size:255" validation:"alpha,required" json:"first_name"`
	LastName  string `gorm:"column:last_name;size:255" validation:"alpha,required" json:"last_name"`
	Email     string `gorm:"column:email;size:255;unique" validation:"email" json:"email"`
}

// TableName returns the name of the database table associated with the User struct.
func (User) TableName() string {
	return "users"
}
