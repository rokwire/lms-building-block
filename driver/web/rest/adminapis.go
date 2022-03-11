package rest

import (
	"lms/core"
	"lms/core/model"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app    *core.Application
	config *model.Config
}
