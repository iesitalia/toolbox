package model

import (
	"context"
	"encoding/json"
	"github.com/getevo/evo/v2"
	"github.com/getevo/evo/v2/lib/db/types"
	"github.com/getevo/evo/v2/lib/generic"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"reflect"
	"strings"
	"toolbox"
	"toolbox/rest"
)

type Tag struct {
	Tag types.JSON `gorm:"column:tag;type:varchar(2048);default:[]"  json:"tag"`
}

// OnCreate calls OnCreateOrUpdate method with the same parameters.
//
// Parameters:
// - db (*gorm.DB): The gorm database connection.
// - object (reflect.Value): The reflect.Value of the object being created.
func (v *Tag) OnCreate(db *gorm.DB, object reflect.Value) {
	v.OnCreateOrUpdate(db, object)
}

// OnUpdate calls OnCreateOrUpdate method with the provided db and object
func (v *Tag) OnUpdate(db *gorm.DB, object reflect.Value) {
	v.OnCreateOrUpdate(db, object)
}

// RestFilter applies a filter to the provided query based on the given `filter` map parameter.
// It checks the 'condition' key in the `filter` map. If the value is "contains", it constructs a query that checks if the column value is in a list of IDs.
// Otherwise, it constructs a query that checks if the column value is equal to a single ID.
//
// Parameters:
// - context: The rest.Context object containing information about the request.
// - query: The gorm.DB object representing the query to be filtered.
// - filter: The map containing the filter conditions.
//
// Example usage:
//
//	filter := map[string]string{
//		"condition": "contains",
//		"value":     "1,2,3",
//	}
//	restFilter(restContext, gormQuery, filter)
//
// This will add a filter to the query such that the column value is checked against the IDs 1, 2, and 3.
func (v Tag) RestFilter(context *rest.Context, query *gorm.DB, filter map[string]string) {
	if filter["condition"] == "contains" {
		query = query.Where(context.Schema.PrimaryFields[0].DBName+" IN (SELECT `id` FROM tag_entity WHERE `table` = ? AND tag_key IN (?))", context.Schema.Table, strings.Split(filter["value"], ","))
	} else {
		query = query.Where(context.Schema.PrimaryFields[0].DBName+" IN (SELECT `id` FROM tag_entity WHERE `table` = ? AND tag_key = ?)", context.Schema.Table, strings.Split(filter["value"], ","))
	}
}

// OnCreateOrUpdate updates or creates tags and tag entities associated with the Tag object in the database.
// If the Tag object is nil, it sets the tag field to an empty JSON string.
// Otherwise, it unmarshals the JSON string from the tag field into a dictionary. If unmarshaling fails, it sets the tag field to an empty JSON string.
// It then iterates through each item in the dictionary and creates TagList and TagEntity objects based on that item. It also keeps track of the tag keys in a separate list.
// After creating all the necessary objects, it performs the following operations using the DBO:
// - If there are tags to create, it inserts the tags and tag entities into the database using the "IGNORE" modifier to handle duplicate entries. It also deletes any tag entities that
func (v *Tag) OnCreateOrUpdate(db *gorm.DB, object reflect.Value) {
	if v == nil {
		err := v.Tag.Scan("{}")
		if err != nil {
			return
		}
	} else {
		var dict = toolbox.Dictionary[string]{}
		err := json.Unmarshal([]byte(v.Tag.String()), &dict)

		if err != nil {
			err := v.Tag.Scan("{}")
			if err != nil {
				return
			}
		}

		var tags []TagList
		var tagEntity []TagEntity
		var tagList []string
		var id int64

		for _, field := range db.Statement.Schema.PrimaryFields {
			v, _ := field.ValueOf(context.Background(), object)
			id = generic.Parse(v).Int64()
		}
		for _, item := range dict {
			tags = append(tags, TagList{Key: item.Key, Value: item.Value})
			tagEntity = append(tagEntity, TagEntity{TagKey: item.Key, Table: db.Statement.Table, ID: id})
			tagList = append(tagList, item.Key)
		}
		dbo := evo.GetDBO()
		if len(tags) > 0 {
			dbo.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&tags)
			dbo.Clauses(clause.Insert{Modifier: "IGNORE"}).Create(&tagEntity)
			dbo.Where("`table` = ? AND `id` = ? AND `tag_key` NOT IN(?)", db.Statement.Table, id, tagList).Delete(&TagEntity{})
		} else {
			dbo.Where("`table` = ? AND `id` = ?", db.Statement.Table, id).Delete(&TagEntity{})
		}

	}
}

// OnDelete deletes the TagEntity associated with the Tag object from the database.
func (v *Tag) OnDelete(db *gorm.DB, object reflect.Value) {
	var id int64
	for _, field := range db.Statement.Schema.PrimaryFields {
		v, _ := field.ValueOf(context.Background(), object)
		id = generic.Parse(v).Int64()
	}
	dbo := evo.GetDBO()
	dbo.Where("`table` = ? AND `id` = ? ", db.Statement.Table, id).Delete(&TagEntity{})
}

// TagList represents the key-value pairs used for tagging entities or objects.
// It has two fields: Key and Value, both of type string.
// Key represents the tag key, while Value represents the corresponding tag value.
type TagList struct {
	Key   string `gorm:"column:key;primaryKey;index:idx_tag_list_key" json:"key,omitempty"`
	Value string `gorm:"column:value;primaryKey;index:idx_tag_list_value" json:"value,omitempty"`
}

// TableName returns the name of the table associated with the TagList struct.
func (TagList) TableName() string {
	return "tag_list"
}

// TableName returns the name of the database table for the Tag struct.
func (Tag) TableName() string {
	return "tag"
}

// TagEntity represents a tag entity that is used to associate tags with specific tables and IDs.
// TagKey is the key of the tag.
// It is stored in the column "tag_key" and is used as an index in the "tag_entity" table.
// It is JSON encoded and is available in the JSON field "json:"tag_key,omitempty"".
type TagEntity struct {
	TagKey string `gorm:"column:tag_key;index:tag_entity_idx;fk:tag_list" json:"tag_key,omitempty"`
	Table  string `gorm:"column:table;index:tag_entity_idx;index:entity_idx" json:"table"`
	ID     int64  `gorm:"column:id;index:tag_entity_idx;index:entity_idx" json:"id"`
}

// TableName returns the name of the table in the database where the TagEntity objects are stored.
func (TagEntity) TableName() string {
	return "tag_entity"
}
