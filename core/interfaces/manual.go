// Copyright 2024 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interfaces

import (
	"lms/core/model"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// Manual exposes manually defined APIs to the driver adapters
type Manual interface {
	//Client
	GetUserData(claims *tokenauth.Claims) (*model.UserDataResponse, error)
}

// Shared exposes shared APIs for other interface implementations
type Shared interface {
	GetUserData(claims tokenauth.Claims) (*model.UserDataResponse, error)
}
