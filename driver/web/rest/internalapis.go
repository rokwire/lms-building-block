package rest

import (
	"lms/core"
	"lms/core/model"
)

// InternalApisHandler handles the rest internal APIs implementation
type InternalApisHandler struct {
	app    *core.Application
	config *model.Config
}
