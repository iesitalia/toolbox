package rest

import (
	"fmt"
	"github.com/getevo/evo/v2/lib/db"
	"github.com/getevo/evo/v2/lib/log"
	"github.com/iesitalia/toolbox/acl"
	"gorm.io/gorm/clause"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/getevo/evo/v2"
	"github.com/iancoleman/strcase"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	scm "github.com/getevo/evo/v2/lib/db/schema"
	"github.com/getevo/evo/v2/lib/outcome"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// GET represents the HTTP GET method.
const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	PUSH   Method = "PUSH"
	DELETE Method = "DELETE"
)

var (
	ListPermission = acl.Permission{
		Key:         "VIEW",
		Name:        "View item(s)",
		Description: "Show list of items or get single item",
	}
	CreatePermission = acl.Permission{
		Key:         "CREATE",
		Name:        "Create",
		Description: "Create new item",
	}
	UpdatePermission = acl.Permission{
		Key:         "UPDATE",
		Name:        "Update",
		Description: "Update items",
	}
	SelfUpdatePermission = acl.Permission{
		Key:         "SELF.UPDATE",
		Name:        "Self Update",
		Description: "Update self created items only",
	}
	DeletePermission = acl.Permission{
		Key:         "DELETE",
		Name:        "Delete",
		Description: "Delete item(s)",
	}
)

// Method represents an HTTP request method.
type Method string

// resources is a map that holds a collection of *Resource objects.
var resources map[string]*Resource

// Context represents the context of an HTTP request.
// It contains information about the request, the object being processed,
// the sample data, the action to be performed, the response, and the schema.
type Context struct {
	Request  *evo.Request
	Object   reflect.Value
	Sample   interface{}
	Action   *Endpoint
	Response *Pagination
	Schema   *schema.Schema
}

// Pagination represents the pagination metadata and data for a response.
type Pagination struct {
	Total      int64       `json:"total"`
	Offset     int         `json:"offset"`
	TotalPages int         `json:"total_pages"`
	Page       int         `json:"current_page"`
	Size       int         `json:"size"`
	Data       interface{} `json:"data"`
	Success    bool        `json:"success"`
	Error      string      `json:"error"`
	Type       string      `json:"type"`
	FilterView *FilterView `json:"filter_view"`
}

// Endpoint represents an API endpoint with specific properties and behaviors.
// - Name: the name of the endpoint.
// - Label: a non-exported field used for internal purposes.
// - Method: an HTTP request method.
// - URL: a non-exported field used for internal purposes.
// - PKUrl: a flag indicating if the endpoint includes primary key in the URL.
// - AbsoluteURI: the absolute URL of the endpoint.
// - Description: a brief description of the endpoint.
// - Object: a reflection value representing the object associated with the endpoint.
// - Handler: a function that handles the endpoint's request.
// - Resource: the resource to which the endpoint belongs.
// - URLParams: an array of filters applied to the URL.
//
// Additional information about related types:
// - Method: represents an HTTP request method.
// - Context: provides access to request data and
type Endpoint struct {
	Name        string                       `json:"name"`
	Permissions []acl.Permission             `json:"permissions"`
	Label       string                       `json:"-"`
	Method      Method                       `json:"method"`
	URL         string                       `json:"-"`
	PKUrl       bool                         `json:"pk_url"`
	AbsoluteURI string                       `json:"url"`
	Description string                       `json:"description"`
	Object      reflect.Value                `json:"-"`
	Handler     func(context *Context) error `json:"-"`
	Resource    *Resource                    `json:"-"`
	URLParams   []Filter                     `json:"-"`
}

// Resource represents a resource in an API.
// It holds information about the object, actions, path, schema, table, name, model, JavaScript model,
// and parameters of the resource.
type Resource struct {
	Object      reflect.Value  `json:"-"`
	Actions     []*Endpoint    `json:"actions"`
	Path        string         `json:"-"`
	Schema      *schema.Schema `json:"-"`
	Table       string         `json:"table"`
	Name        string         `json:"model"`
	Model       *scm.Model     `json:"-"`
	JSModel     string         `json:"js_model"`
	Params      []Param        `json:"params"`
	Feature     *Feature       `json:"feature"`
	Permissions acl.App        `json:"permissions"`
}

// GetResource retrieves a Resource object based on the provided input. It checks if a Resource with the same type already exists in the resources map and returns it if found. Otherwise
func GetResource(input interface{}) (*Resource, error) {
	if v, ok := resources[getObject(input).Type().String()]; ok {
		return v, nil
	}
	return nil, ErrorObjectNotExist
}

// AttachResource creates a new Resource object using the provided model and adds it to the resources map.
// It also defines a series of actions on the resource:
// - ORM: Creates an endpoint for the ORM SDK
// - ALL: Returns all objects in one call
// - PAGINATE: Paginates objects
// - GET: Returns a single object using its primary key
// - CREATE: Creates an object using given values
// - FIND: Searches for object(s) by given criteria
// - FIRST: Returns the first object from the database
// - LAST: Returns the last object from the database
// - TAKE: Takes an object from the database
// - UPDATE.PUT: Batch updates objects
// - UPDATE.POST: Updates a single object using its primary key
// - DELETE: Deletes an existing object using its primary key
// The function then adds parameters to the resource based on the fields in the model's schema.
func AttachResource(model *scm.Model) *Resource {
	var feature = GetFeatures(model.Sample)

	var resource = Resource{
		Object:  model.Value,
		Name:    model.Name,
		Model:   model,
		JSModel: model.Name + "Model",
		Table:   model.Table,
		Feature: feature,
	}
	resource.Schema = model.Schema
	resource.Path = model.Table
	resources[model.Name] = &resource
	if !feature.EnableAPI {
		return &resource
	}
	resource.Action(&Endpoint{
		Name:        "MODEL INFO",
		Method:      GET,
		URL:         "/info",
		Handler:     ModelInfo,
		Description: "return information of the model",
	})
	if !feature.DisableView {
		if v, ok := resource.Object.Interface().(interface{ FilterView() FilterView }); ok {
			if !feature.DisableView {
				resource.Action(&Endpoint{
					Name:        "FILTER VIEW",
					Method:      GET,
					PKUrl:       false,
					URL:         "/filter-view",
					URLParams:   v.FilterView().URLParams,
					Handler:     FilterViewHandler,
					Description: "return object filter view",
					Permissions: []acl.Permission{ListPermission},
				})
			}
		}

		resource.Action(&Endpoint{
			Name:        "ALL",
			Method:      GET,
			URL:         "/all",
			Handler:     All,
			Description: "return all objects in one call",
			Permissions: []acl.Permission{ListPermission},
		})

		resource.Action(&Endpoint{
			Name:        "PAGINATE",
			Method:      GET,
			URL:         "/paginate",
			Handler:     Paginate,
			Description: "paginate objects",
			Permissions: []acl.Permission{ListPermission},
		})

		resource.Action(&Endpoint{
			Name:        "GET",
			Method:      GET,
			URL:         "/",
			PKUrl:       true,
			Handler:     Get,
			Description: "get single object using primary key",
			Permissions: []acl.Permission{ListPermission},
		})
	}
	if !feature.DisableCreate {
		resource.Action(&Endpoint{
			Name:        "CREATE",
			Method:      PUT,
			URL:         "/",
			Handler:     Create,
			Description: "create an object using given values",
			Permissions: []acl.Permission{CreatePermission},
		})
	}
	if !feature.DisableUpdate {
		resource.Action(&Endpoint{
			Name:        "UPDATE",
			Method:      POST,
			URL:         "/",
			PKUrl:       true,
			Handler:     Update,
			Description: "update single object select using primary key",
			Permissions: []acl.Permission{UpdatePermission, SelfUpdatePermission},
		})
	}
	if !feature.DisableDelete {
		resource.Action(&Endpoint{
			Name:        "DELETE",
			Method:      DELETE,
			URL:         "/",
			PKUrl:       true,
			Handler:     Delete,
			Description: "delete existing object using primary key",
			Permissions: []acl.Permission{DeletePermission},
		})
	}

	if feature.EnableSetAPI {
		var key = ""
		for _, field := range model.Schema.Fields {
			if _, ok := field.TagSettings["SET_KEY"]; ok {
				key = field.DBName
				break
			}
		}
		if key == "" {
			log.Fatalf("object " + model.Name + " has rest.EnableSetAPI set to true, but no SET_KEY tag is found in the model definition.")
		}
		resource.Action(&Endpoint{
			Name:        "SET",
			Method:      PUT,
			URL:         "/:" + key + "/set",
			Handler:     Set,
			Description: "set multiple values base on set_key at once",
			Permissions: []acl.Permission{UpdatePermission},
		})
	}

	for _, field := range model.Schema.Fields {
		resource.Params = append(resource.Params, Param{
			Name:    field.DBName,
			Type:    field.FieldType.String(),
			Primary: field.PrimaryKey,
		})
	}

	return &resource
}

// GetFeatures represents the features of a resource
func GetFeatures(v interface{}) *Feature {
	var features = Feature{}
	var typ = reflect.ValueOf(v)
	for i := 0; i < typ.NumField(); i++ {
		switch typ.Field(i).Type().String() {
		case "rest.API":
			features.EnableAPI = true
		case "rest.DisableCreate":
			features.DisableCreate = true
		case "rest.EnableSetAPI":
			features.EnableSetAPI = true
		case "rest.DisableUpdate":
			features.DisableUpdate = true
		case "rest.DisableDelete":
			features.DisableDelete = true
		case "rest.DisableView":
			features.DisableView = true
		}

	}
	return &features
}

// getObject retrieves the reflect.Value representation of the object passed as a parameter.
// If the object is a pointer, it gets the value it points to.
// If the object is not a struct, it panics with an error message.
// The reflect.Value representation of the object is then returned.
func getObject(object interface{}) reflect.Value {
	var v = reflect.ValueOf(object)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("Invalid type " + v.Kind().String() + " provided")
	}
	return v
}

// Action is a method of the Resource type that registers an action for the resource.
// It takes an Action pointer as a parameter and
func (res *Resource) Action(action *Endpoint) {
	//todo: check duplicate actions and override
	action.Name = strcase.ToCamel(action.Name)
	if action.Method == "" {
		action.Method = POST
	}
	if action.URL == "" {
		action.URL = strcase.ToSnake(action.Name)
	}
	action.URL = strings.Trim(action.URL, "/")
	if action.PKUrl {
		for _, item := range res.Schema.PrimaryFields {
			action.URL += "/:" + item.DBName
		}
	}

	for _, item := range action.URLParams {
		action.URL += "/:" + item.Name
	}

	action.Object = res.Object
	res.Path = res.Table
	action.AbsoluteURI = "/" + strings.Trim(PREFIX+"/rest/"+res.Path+"/"+strings.Trim(action.URL, "/"), "/")

	action.Resource = res

	switch action.Method {
	case GET:
		evo.Get(action.AbsoluteURI, action.requestHandler)
	case POST:
		evo.Post(action.AbsoluteURI, action.requestHandler)
	case PUT:
		evo.Put(action.AbsoluteURI, action.requestHandler)
	case DELETE:
		evo.Delete(action.AbsoluteURI, action.requestHandler)
	default:
		panic("invalid method passed")
	}
	res.Actions = append(res.Actions, action)

	for idx, perm := range action.Permissions {
		var found = false
		for _, item := range res.Permissions.Permissions {
			if item.Key == perm.Key {
				found = true
				break
			}
		}
		if !found {
			res.Permissions.Permissions = append(res.Permissions.Permissions, action.Permissions[idx])
		}
	}
}

// requestHandler handles the incoming request and returns a response.
// It takes in a `Request` object and returns an `interface{}`.
// It creates a new `Context` object with the request, action, object, and default response.
// If the action has a handler defined
func (action *Endpoint) requestHandler(request *evo.Request) interface{} {
	context := &Context{
		Request: request,
		Action:  action,
		Object:  action.Object,
		Response: &Pagination{
			TotalPages: 1,
			Total:      1,
			Page:       1,
			Size:       1,
			Success:    true,
		},
	}

	var stmt = evo.GetDBO().Model(action.Object.Interface()).Statement
	err := stmt.Parse(action.Object.Interface())
	if err != nil {
		return err
	}
	context.Schema = stmt.Schema
	if action.Handler != nil {
		if err := action.Handler(context); err != nil {
			context.SetError(err)
		}
	} else {
		context.SetError(fmt.Errorf("unimplemented handler"))
	}

	return outcome.Json(context.GetResponse())
}

// GetObject is a method of the Context type that returns a new indirect reflect.Value of the context Object's type.
func (context *Context) GetObject() reflect.Value {
	return reflect.Indirect(reflect.New(context.Object.Type()))
}

// GetResponse is a method of the Context type.
// It returns the response object of the context.
// If the response is not successful (success flag is false), it sets the data, size, page, total, total pages, and offset fields of the response to 0.
// It then returns the response object.
func (context *Context) GetResponse() interface{} {
	if !context.Response.Success {
		context.Response.Data = 0
		context.Response.Size = 0
		context.Response.Page = 0
		context.Response.Total = 0
		context.Response.TotalPages = 0
		context.Response.Offset = 0
	}
	return context.Response
}

// SetResponse is a method of the Context type that sets the response data.
// It takes a response interface{} as a parameter and marshals it to JSON.
// If the response is nil, it returns immediately.
// If the response is not a slice, it marshals it as a single element slice.
// Otherwise, it marshals the response as is.
// If there is an error during marshaling, it returns immediately without setting the response.
// Note: The response is set to the `context.Request` using the `JSON` method.
func (context *Context) SetResponse(response interface{}) {
	// context.Response.Type = context.Object.Type().String()
	if response == nil {
		return
	}
	var v = reflect.ValueOf(response)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		err := context.Request.JSON([]interface{}{response})
		if err != nil {
			return
		}
	} else {
		err := context.Request.JSON(response)
		if err != nil {
			return
		}
	}

}

// SetError is a method of the Context type that sets the error message in the Response field and marks the Response as unsuccessful.
// It takes an error parameter.
func (context *Context) SetError(error error) {
	context.Response.Error = error.Error()
	context.Response.Success = false
}

// GetObjectSlice returns a new indirect reflect value of a slice of the type of the Object field in the Context.
func (context *Context) GetObjectSlice() reflect.Value {
	return reflect.Indirect(reflect.New(reflect.SliceOf(context.Object.Type())))
}

// FindByPrimaryKey is a method that searches for a record in the database based on the primary key values provided.
// The method takes an input parameter, which can be a struct or a
func (context *Context) FindByPrimaryKey(input interface{}) (bool, error) {
	var err error
	var dbo = context.GetDBO().Debug()
	var association = context.Request.Query("associations").String()
	if association != "" {
		if association == "1" || association == "true" {
			dbo = dbo.Preload(clause.Associations)
		} else if association == "deep" {
			var preload = getAssociations("", context.Schema)
			for _, item := range preload {
				dbo = dbo.Preload(item)
			}
		} else {
			var ls = strings.Split(association, ",")
			for _, item := range ls {
				dbo = dbo.Preload(item)
			}
		}

	}
	var where []string
	var params []interface{}
	for _, field := range context.Action.Resource.Schema.PrimaryFields {
		var v interface{} = context.Request.Param(field.DBName).String()
		if v == "" {
			v = getValueByFieldName(input, field.Name)
		}
		where = append(where, field.DBName+" = ?")
		params = append(params, v)
	}

	var join = context.Request.Query("join").String()
	if len(join) > 0 {
		if relations := relationsMapper(join); relations != "" {
			dbo = dbo.Preload(relations)
		}
	}
	dbo, err = filterMapper(context.Request.QueryString(), context, dbo)
	return dbo.Where(strings.Join(where, " AND "), params...).Take(input).RowsAffected != 0, err
}

// orderRegex is a regular expression that matches strings in the format of "[field] [asc|desc]" where:
//   - [field] represents a sequence of letters, numbers, hyphens, and underscores
//   - [asc|desc] represents either the word "asc" or "desc"
//
// The regular expression is case-insensitive and accepts leading and trailing whitespace characters.
var orderRegex = regexp.MustCompile(`(?mi)\s*[a-zA-Z0-9-_]+\s+(asc|desc)\s*`)

// ApplyFilters applies filters to the query based on the request parameters in the context. It modifies the
func (context *Context) ApplyFilters(query *gorm.DB) (*gorm.DB, error) {
	/*	if context.Request.Query("associations").String() != "" {
			query = query.Preload(clause.Associations)
		}
	*/

	var association = context.Request.Query("associations").String()
	if association != "" {
		if association == "1" || association == "true" {
			query = query.Preload(clause.Associations)
		} else if association == "deep" {
			var preload = getAssociations("", context.Schema)
			fmt.Println(context.Schema.Name, preload)
			for _, item := range preload {
				query = query.Preload(item)
			}
		} else {
			var ls = strings.Split(association, ",")
			for _, item := range ls {
				query = query.Preload(item)
			}
		}

	}

	var order = context.Request.Query("order").String()
	if order != "" {
		valid := true
		for _, item := range strings.Split(order, ",") {
			if !orderRegex.MatchString(item) {
				valid = false
				break
			}
		}
		if valid {
			query = query.Order(order)
		}
	}

	var fields = context.Request.Query("fields").String()
	if len(fields) > 0 {
		splitFields := strings.Split(fields, ",")
		query = query.Select(splitFields)
	}

	var join = context.Request.Query("join").String()
	if len(join) > 0 {
		if relations := relationsMapper(join); relations != "" {
			query = query.Preload(relations)
		}
	}
	var err error
	query, err = filterMapper(context.Request.QueryString(), context, query)

	var offset = context.Request.Query("offset").Int()
	if offset > 0 {
		query = query.Offset(offset)
	}

	var limit = context.Request.Query("limit").Int()
	if limit > 0 {
		query = query.Limit(limit)
	}
	return query, err
}

func getAssociations(prefix string, s *schema.Schema, loaded ...string) []string {
	var preload []string
	if len(loaded) == 0 {
		loaded = []string{s.Table}
	}

	var relations []*schema.Relationship
	relations = append(relations, s.Relationships.HasOne...)
	relations = append(relations, s.Relationships.BelongsTo...)
	relations = append(relations, s.Relationships.HasMany...)

	for idx, _ := range relations {

		var relation = relations[idx]

		var chunks = strings.Split(loaded[0], ".")
		if len(chunks) > 2 && chunks[len(chunks)-2] == s.Table {
			continue
		}
		if len(chunks) > 4 {
			continue
		}
		loaded[0] += "." + relation.Field.Schema.Table
		preload = append(preload, strings.TrimLeft(prefix+"."+relation.Field.Name, "."))
		/*	query = query.Preload(strings.TrimLeft(prefix+"."+relation.Field.Name, "."))
			fmt.Println("preload:", strings.TrimLeft(prefix+"."+relation.Field.Name, "."))
		*/
		for i, _ := range relation.FieldSchema.Relationships.Relations {
			var item = relation.FieldSchema.Relationships.Relations[i]
			if item.Schema.Table == s.Table {
				continue
			}
			preload = append(preload, getAssociations(prefix+"."+relation.Field.Name, item.Schema, loaded...)...)

		}

	}
	return preload
}

// GetDBO returns a pointer to the *gorm.DB object.
// It retrieves the *gorm.DB object from the `evo` package.
// If the "language" header is present in the request, it sets
func (context *Context) GetDBO() *gorm.DB {
	var dbo = evo.GetDBO()
	if context.Request.Header("language") != "" {
		dbo = db.Set("lang", context.Request.Header("language"))
	} else {
		if context.Request.Cookie("l10n-language") != "" {
			dbo = db.Set("lang", context.Request.Cookie("l10n-language"))
		}
	}
	return dbo
}

func (context *Context) HasPerm(s string) error {
	if context.Action.Resource.Feature.CheckPermission {
		var user = context.Request.User()
		if user.Anonymous() {
			return ErrorUnauthorized
		}
		if !user.HasPermission(context.Action.Resource.Permissions.App + "." + s) {
			return ErrorPermissionDenied
		}
	}

	return nil
}

// getValueByFieldName retrieves the value of a field by its name from the given input object.
// It uses reflection to access the field value and returns it as an interface{}.
// If the field does not exist, it returns nil.
func getValueByFieldName(input interface{}, field string) interface{} {
	ref := reflect.ValueOf(input)
	f := reflect.Indirect(ref).FieldByName(field)
	return f.Interface()
}

// filterConditions is a map that defines the filter conditions used in the filterMapper function.
// The keys represent the condition name, and the values represent the corresponding condition symbol or keyword.
var filterConditions = map[string]string{
	"eq":       "=",
	"neq":      "!=",
	"gt":       ">",
	"lt":       "<",
	"gte":      ">=",
	"lte":      "<=",
	"in":       "in",
	"contains": "LIKE",
	"isnull":   "IS NULL",
	"notnull":  "IS NOT NULL",
}

// ContainOperator represents the string value "contains"
// which is used as an operator for containment operations.
// Examples of containment operations could be checking if a string contains
// a specific substring or if an array contains a specific element.
// This constant is used to indicate the containment operator in code logic.
// NotNullOperator represents the string value "notnull"
// which is used as an operator for checking if a value is not null.
// This constant is used to indicate the not null operator in code logic.
// IsNullOperator represents the string value "isnull"
// which is used as an operator for checking if a value is null.
// This constant is used to indicate the is null operator in code logic.
// InOperator represents the string
const (
	ContainOperator = "contains"
	NotNullOperator = "notnull"
	IsNullOperator  = "isnull"
	InOperator      = "in"
)

// relationsMapper maps the given string of relations into a formatted nested relation string.
// It splits the input string by comma and then splits each relation by dot.
// It converts the first letter of each relation to uppercase using proper language casing rules.
// It joins the titled relation strings with dots.
// If the resulting nested relation string is not empty, it is returned. Otherwise, an empty string is returned.
func relationsMapper(joins string) string {
	relations := strings.Split(joins, ",")
	for _, relation := range relations {
		nestedRelationsSlice := strings.Split(relation, ".")
		titledSlice := make([]string, len(nestedRelationsSlice))
		for i, relation := range nestedRelationsSlice {
			titledSlice[i] = cases.Title(language.English, cases.NoLower).String(relation)
		}
		nestedRelation := strings.Join(titledSlice, ".")
		if len(nestedRelation) > 0 {
			return nestedRelation
		}
	}
	return ""
}

// filterMapper applies filters to the given query based on the provided filter string.
// It parses the filter
func filterMapper(filters string, context *Context, query *gorm.DB) (*gorm.DB, error) {
	fRegEx := filterRegEx(filters)
	for _, filter := range fRegEx {
		var obj = context.GetObject().Interface()
		var ref = reflect.ValueOf(obj)
		fnd := false
		var fieldName = ""
		filter["value"], _ = url.QueryUnescape(filter["value"])
		for _, field := range context.Schema.Fields {
			if field.DBName == filter["column"] {
				fieldName = field.Name
				fnd = true
				break
			}
		}
		if !fnd {
			return nil, ErrorColumnNotExist
		}
		v := ref.FieldByName(fieldName)

		if obj, ok := v.Interface().(interface {
			RestFilter(context *Context, query *gorm.DB, filter map[string]string)
		}); ok {
			obj.RestFilter(context, query, filter)
			return query, nil
		}

		if filter["condition"] == NotNullOperator || filter["condition"] == IsNullOperator {
			if filter["column"] == "deleted_at" {
				query = query.Unscoped()
			}
			query = query.Where(fmt.Sprintf("`%s` %s", filter["column"], filterConditions[filter["condition"]]))
		} else {
			if filter["condition"] == ContainOperator {
				query = query.Where(fmt.Sprintf("`%s` %s ?", filter["column"], "LIKE"), fmt.Sprintf("%%%s%%", filter["value"]))
			} else if filter["condition"] == InOperator {
				valSlice := strings.Split(filter["value"], ",")
				query = query.Where(fmt.Sprintf("`%s` IN (?)", filter["column"]), valSlice)
			} else {
				if v, ok := filterConditions[filter["condition"]]; ok {
					query = query.Where(fmt.Sprintf("`%s` %s ?", filter["column"], v), filter["value"])
				} else {
					return query, fmt.Errorf("invalid filter condition %s", filter["condition"])
				}

			}
		}
	}
	query = query.Debug()
	return query, nil
}

// result will be [{"column":"column1","condition":"condition1","value":"value1"},{"column":"column2","condition":"condition2","value":"value2"},{"column":"column3","condition":"condition
func filterRegEx(str string) []map[string]string {
	var re = regexp.MustCompile(`(?m)((?P<column>[a-zA-Z_\-0-9]+)\[(?P<condition>[a-zA-Z]+)\](\=((?P<value>[a-zA-Z_\-0-9\s\%\,]+))){0,1})\&*`)
	var keys = re.SubexpNames()
	var result = []map[string]string{}
	for _, match := range re.FindAllStringSubmatch(str, -1) {
		item := map[string]string{}
		for i, name := range keys {
			if i != 0 && name != "" {
				item[name] = match[i]
			}
		}
		result = append(result, item)
	}
	return result
}

// API represents an API feature.
type API struct{}

// EnableSetAPI enable set api endpoint
type EnableSetAPI struct{}

// DisableCreate is an empty struct that can be embedded in a struct to disable the creation of instances.
type DisableCreate struct{}

// DisableUpdate is a marker type used to disable the update functionality for certain structs.
// It is typically embedded in other struct types to indicate that updating instances of that struct is not allowed.
type DisableUpdate struct{}

// DisableDelete represents a type that disables the delete functionality in a REST API.
type DisableDelete struct{}

// DisableView represents a view that should be disabled, indicating that no content
// should be rendered for the corresponding HTTP request.
type DisableView struct{}

func (c API) RESTFeature() bool {
	return true
}

// Feature struct represents a set of features that can be enabled or disabled.
// Each feature has a corresponding boolean field that indicates whether it is enabled or disabled.
type Feature struct {
	EnableAPI              bool
	DisableCreate          bool
	DisableUpdate          bool
	DisableView            bool
	DisablePermissionCheck bool
	DisableDelete          bool
	CheckPermission        bool
	EnableSetAPI           bool
}

type AppPermission struct {
	App               string           `gorm:"column:app;size:64;primaryKey" json:"app"`
	Name              string           `gorm:"column:name;size:64" json:"name"`
	Description       string           `gorm:"column:description;size:64" json:"description"`
	CustomPermissions []acl.Permission `gorm:"-" json:"permissions"`
	Objects           []interface{}
}

func SetPermission(app *AppPermission) {
	var a = acl.App{
		App:         app.App,
		Name:        app.Name,
		Description: app.Description,
		Permissions: append([]acl.Permission{ListPermission, CreatePermission, UpdatePermission, SelfUpdatePermission, DeletePermission}, app.CustomPermissions...),
	}

	for _, obj := range app.Objects {
		resource, err := GetResource(obj)
		if err != nil {
			continue
		}
		resource.Permissions = a
		resource.Feature.CheckPermission = true
	}

	acl.SetPermission(&a)

}
