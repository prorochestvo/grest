package grest

import (
	"fmt"
	"github.com/prorochestvo/grest/db"
	"github.com/prorochestvo/grest/internal/helper"
	"github.com/prorochestvo/grest/internal/mux"
	"regexp"
)

type Controller interface {
	Path() string
	Actions() []Action
}

type ControllerWithID interface {
	Id() (name, pattern string)
	Controller
}

type ControllerWithModel interface {
	Model() Model
	ControllerWithID
}

type ControllerWithRouteCustom interface {
	Controller
	CustomRoutes(mux.Splitter)
}

type ControllerWithMigrations interface {
	db.MigrationController
	ControllerWithModel
}

/***********************************************************************************************************************
 * helper
 */
func getControllerID(controller Controller) string {
	name := "id"
	pattern := "[0-9]"
	if c, ok := controller.(ControllerWithID); ok == true && c != nil {
		name, pattern = c.Id()
		if len(pattern) == 0 {
			pattern = "[0-9]+"
			name = helper.HttpPathTrim(name)
		}
		if len(name) == 0 {
			name = "id"
		}
	}
	pattern = helper.HttpPathTrim(pattern)
	name = regexp.MustCompile(`[^A-Za-z_-]+`).ReplaceAllString(name, "")
	name = helper.HttpPathTrim(name)
	return fmt.Sprintf("{%s:%s}", name, pattern)
}

func getControllerModel(controller Controller) Model {
	var result Model = nil
	if c, ok := controller.(ControllerWithModel); ok == true && c != nil {
		if m := c.Model(); m != nil {
			result = m
		}
	}
	return result
}

func makeControllerActionPath(controller Controller, action Action) string {
	path := ""
	prefix := helper.HttpPathTrim(controller.Path())
	suffix := helper.HttpPathTrim(action.Path())
	id := makeControllerActionID(controller, action)
	if len(prefix) > 0 {
		path = fmt.Sprintf("%s/%s", path, prefix)
	}
	if len(suffix) > 0 {
		path = fmt.Sprintf("%s/%s", path, suffix)
	}
	if action.WithID() && len(id) > 0 {
		path = fmt.Sprintf("%s/%s", path, id)
	}
	return path
}

func makeControllerActionID(controller Controller, action Action) string {
	result := getActionID(action)
	if len(result) == 0 {
		result = getControllerID(controller)
	}
	return result
}
