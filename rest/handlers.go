package rest

import (
	"errors"
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/generic"
	"github.com/iesitalia/toolbox"
	"gorm.io/gorm/clause"
)

// ErrorObjectNotExist represents an error indicating that the object does not exist.
var ErrorObjectNotExist = errors.New("object does not exists")

// ErrorColumnNotExist represents an error indicating that a column does not exist.
var ErrorColumnNotExist = errors.New("column does not exists")

var ErrorPermissionDenied = errors.New("permission denied")

var ErrorUnauthorized = errors.New("unauthorized")

func Set(context *Context) error {
	if err := context.HasPerm("UPDATE"); err != nil {
		return err
	}
	//var dbo = context.GetDBO()
	object := context.GetObject()

	var key = ""
	var fieldName = ""
	for _, field := range context.Schema.Fields {
		if _, ok := field.TagSettings["SET_KEY"]; ok {
			key = field.DBName
			fieldName = field.Name
			break
		}
	}

	array := context.GetObjectSlice()
	ptr := array.Addr()
	err := context.Request.BodyParser(ptr.Interface())
	if err != nil {
		return err
	}
	var pk = context.Request.Param(key)
	err = db.Debug().Where("`"+key+"` = ?", pk).Delete(object.Addr().Interface()).Error
	if err != nil {
		return err
	}
	for i := 0; i < array.Len(); i++ {
		err := generic.Parse(pk).Cast(ptr.Elem().Index(i).FieldByName(fieldName).Addr().Interface())
		if err != nil {
			return err
		}

	}
	err = db.Create(ptr.Interface()).Error
	if err != nil {
		return err
	}
	context.Response.Data = ptr.Interface()

	return nil
}

// Create takes a Context as input and creates a new object.
// It uses the context's Request and DBO to perform the creation.
// The object to be created is retrieved from the context's Object field.
// The object is parsed from the request's body using the BodyParser method.
// The object can optionally implement the BeforeCreate method, which is called before the creation.
// The object can optionally implement the ValidateCreate method, which is called to validate the object before creation.
// The object is then created in the database using the DBO's Create method.
// If the object implements the AfterCreate method, it is called after the creation.
// The created object is set as the data in the context's Response field.
// Returns an error if any error occurs during the creation process.
func Create(context *Context) error {
	if err := context.HasPerm("CREATE"); err != nil {
		return err
	}
	var dbo = context.GetDBO()
	object := context.GetObject()
	ptr := object.Addr().Interface()
	err := context.Request.BodyParser(ptr)
	if err != nil {
		return err
	}

	if obj, ok := ptr.(interface{ BeforeCreate(context *Context) error }); ok {
		err := obj.BeforeCreate(context)
		if err != nil {
			return err
		}
	}

	if obj, ok := ptr.(interface{ ValidateCreate(context *Context) error }); ok {
		if err := obj.ValidateCreate(context); err != nil {
			return err
		}
	}
	if err := dbo.Create(ptr).Error; err != nil {
		return err
	}

	if obj, ok := ptr.(interface{ AfterCreate(context *Context) error }); ok {
		if err := obj.AfterCreate(context); err != nil {
			return err
		}
	}
	context.Response.Data = ptr
	return nil
}

// Update updates an object in the database based on the provided context.
// It retrieves the database object, checks if it exists in the database,
// parses the request body to update the object, and executes the updates
// on the database. It also calls the BeforeUpdate and ValidateUpdate methods
// if they are implemented by the object to perform any necessary operations
// before and after the update. Finally, it sets the updated object as the response
// data in the context.
func Update(context *Context) error {
	if err := context.HasPerm("UPDATE"); err != nil {
		return err
	}
	var dbo = context.GetDBO()
	object := context.GetObject()
	ptr := object.Addr().Interface()
	key, err := context.FindByPrimaryKey(ptr)

	if err != nil {
		return err
	}
	if !key {
		return ErrorObjectNotExist
	}
	err = context.Request.BodyParser(ptr)
	if err != nil {
		return err
	}
	if obj, ok := ptr.(interface{ BeforeUpdate(context *Context) error }); ok {
		if err := obj.BeforeUpdate(context); err != nil {
			return err
		}
	}

	if obj, ok := ptr.(interface{ ValidateUpdate(context *Context) error }); ok {
		if err := obj.ValidateUpdate(context); err != nil {
			return err
		}
	}
	//evo.Dump(ptr)
	if err := dbo.Debug().Omit(clause.Associations).Save(ptr).Error; err != nil {
		return err
	}

	if obj, ok := ptr.(interface{ AfterUpdate(context *Context) error }); ok {
		if err := obj.AfterUpdate(context); err != nil {
			return err
		}
	}
	context.Response.Data = ptr

	return nil
}

// Delete deletes an object from the database.
// It takes a Context pointer as a parameter.
// It returns an error if an error occurs during the deletion process.
func Delete(context *Context) error {
	if err := context.HasPerm("DELETE"); err != nil {
		return err
	}
	var dbo = context.GetDBO()
	object := context.GetObject()
	ptr := object.Addr().Interface()
	key, err := context.FindByPrimaryKey(ptr)
	if err != nil {
		return err
	}
	if !key {
		return ErrorObjectNotExist
	}
	if obj, ok := ptr.(interface{ BeforeDelete(context *Context) error }); ok {
		if err := obj.BeforeDelete(context); err != nil {
			return err
		}
	}

	// Try soft-delete
	if obj, ok := ptr.(interface{ Delete(v bool) }); ok {
		obj.Delete(true)
		if err := dbo.Updates(ptr).Error; err != nil {
			return err
		}
	} else {
		if err := dbo.Delete(ptr).Error; err != nil {
			return err
		}
	}

	if obj, ok := ptr.(interface{ AfterDelete(context *Context) error }); ok {
		if err := obj.AfterDelete(context); err != nil {
			return err
		}
	}

	return nil
}

// Get is a function that retrieves an object from the context.
// It performs pre and post get operations on the object if they are implemented.
// It finds the object by its primary key, sets it as the response data in the context, and returns nil if successful.
// If the object does not exist, it returns an error of type ErrorObjectNotExist.
// It returns an error if any operation fails.
func Get(context *Context) error {
	if err := context.HasPerm("VIEW"); err != nil {
		return err
	}
	object := context.GetObject()
	ptr := object.Addr().Interface()
	if obj, ok := ptr.(interface{ BeforeGet(context *Context) error }); ok {
		if err := obj.BeforeGet(context); err != nil {
			return err
		}
	}
	key, err := context.FindByPrimaryKey(ptr)
	if err != nil {
		return err
	}
	if !key {
		return ErrorObjectNotExist
	}

	if obj, ok := ptr.(interface{ AfterGet(context *Context) error }); ok {
		if err := obj.AfterGet(context); err != nil {
			return err
		}
	}

	context.Response.Data = ptr
	return nil
}

// All queries the database and retrieves all objects based on the given context.
// It applies filters, handles before and after events, and sets the response.
// It returns an error if any occurred during the process.
func All(context *Context) error {
	if err := context.HasPerm("VIEW"); err != nil {
		return err
	}
	var dbo = context.GetDBO()

	var slice = context.GetObjectSlice()
	ptr := slice.Addr().Interface()
	if obj, ok := context.GetObject().Addr().Interface().(interface{ BeforeGet(context *Context) error }); ok {
		if err := obj.BeforeGet(context); err != nil {
			return err
		}
	}

	var err error
	dbo, err = context.ApplyFilters(dbo)
	if err != nil {
		return err
	}
	if err := dbo.Find(ptr).Error; err != nil {
		return err
	}
	context.Response.Total = int64(slice.Len())
	context.Response.Size = slice.Len()

	if _, ok := context.GetObject().Addr().Interface().(interface{ AfterGet(context *Context) error }); ok {
		for i := 0; i < slice.Len(); i++ {
			if obj, ok := slice.Index(i).Addr().Interface().(interface{ AfterGet(context *Context) error }); ok {
				if err := obj.AfterGet(context); err != nil {
					return err
				}
			}
		}
	}

	context.Response.Data = ptr
	context.SetResponse(ptr)
	return nil
}

// Paginate applies pagination to a database query based on the context provided.
// It modifies the context's response object with the paginated data.
func Paginate(context *Context) error {
	if err := context.HasPerm("VIEW"); err != nil {
		return err
	}
	var slice = context.GetObjectSlice()

	if obj, ok := context.GetObject().Addr().Interface().(interface{ BeforeGet(context *Context) error }); ok {
		if err := obj.BeforeGet(context); err != nil {
			return err
		}
	}

	ptr := slice.Addr().Interface()
	var p toolbox.Pagination
	p.SetLimit(context.Request.Query("size").Int())
	p.SetCurrentPage(context.Request.Query("page").Int())
	context.Response.Size = p.Limit
	context.Response.Offset = p.GetOffset()
	context.Response.Page = p.CurrentPage

	var query = db.Model(ptr)
	var err error
	query, err = context.ApplyFilters(query)
	if err != nil {
		return err
	}
	query.Model(ptr).Count(&context.Response.Total)
	p.Records = int(context.Response.Total)
	p.SetPages()
	context.Response.TotalPages = p.Pages
	if err := query.Limit(p.Limit).Offset(p.GetOffset()).Find(ptr).Error; err != nil {
		return err
	}
	if _, ok := context.GetObject().Addr().Interface().(interface{ AfterGet(context *Context) error }); ok {
		for i := 0; i < slice.Len(); i++ {
			if obj, ok := slice.Index(i).Addr().Interface().(interface{ AfterGet(context *Context) error }); ok {
				if err := obj.AfterGet(context); err != nil {
					return err
				}
			}
		}
	}
	context.Response.Data = ptr
	context.SetResponse(ptr)
	return nil
}

// FilterViewHandler filters the view based on the request parameters and updates the context.Response accordingly.
func FilterViewHandler(context *Context) error {
	if err := context.HasPerm("VIEW"); err != nil {
		return err
	}
	if obj, ok := context.Object.Interface().(interface{ FilterView() FilterView }); ok {
		var fv = obj.FilterView()

		context.Response.Offset = context.Request.Query("offset").Int()
		context.Response.Page = context.Request.Query("page").Int()
		context.Response.Size = context.Request.Query("size").Int()

		if context.Response.Offset == 0 && context.Response.Page > 0 {
			context.Response.Page = context.Response.Page - 1
			context.Response.Offset = context.Response.Page * context.Response.Size
		}

		if context.Response.Page == 0 && context.Response.Offset > 0 {
			context.Response.Page = int(float64(context.Response.Offset / context.Response.Size))
		}
		if context.Response.Size > 100 {
			context.Response.Size = 100
		}
		if context.Response.Size == 0 {
			context.Response.Size = 25
		}
		err, total, data := fv.GetData(context.Response.Offset, context.Response.Size, context.Request)
		if err != nil {
			return err
		}

		context.Response.Total = total
		if context.Response.Size > 0 {
			context.Response.TotalPages = int(context.Response.Total/int64(context.Response.Size)) + 1
		}
		context.Response.Data = data
		context.Response.FilterView = &fv
		context.Response.Type = "filterview"
		return nil
	}
	return fmt.Errorf("invalid filterview handler")
}

// Info represents a structured information object.
//
// It contains the following fields:
// - Name: The name of the object
// - ID: The ID of the object
// - Fields: An array of Field objects that represent the fields of the object
// - Endpoints: An array of Endpoint objects that represent the endpoints associated with the object.
type Info struct {
	Name      string      `json:"name,omitempty"`
	ID        string      `json:"id,omitempty"`
	Fields    []Field     `json:"fields,omitempty"`
	Endpoints []*Endpoint `json:"endpoints,omitempty"`
}

// ModelInfo retrieves information about a model and populates the response data with the info.
func ModelInfo(context *Context) error {
	if err := context.HasPerm("VIEW"); err != nil {
		return err
	}
	var info = Info{
		Name: context.Object.Type().Name(),
		ID:   context.Schema.Table,
	}

	for _, item := range context.Schema.Fields {
		info.Fields = append(info.Fields, Field{
			Name:    item.Name,
			DBName:  item.DBName,
			Type:    item.FieldType.Name(),
			Default: item.DefaultValue,
			PK:      item.PrimaryKey,
		})
	}
	info.Endpoints = resources[context.Action.Resource.Name].Actions
	context.Response.Data = info
	return nil
}
