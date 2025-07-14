// Copyright 2022 Board of Trustees of the University of Illinois.
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

package corebb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"lms/core/model"
	"log"
	"net/http"
	"strings"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

// Adapter implements the Core interface
type Adapter struct {
	logger                logs.Logger
	coreURL               string
	serviceAccountManager *auth.ServiceAccountManager

	appID string
	orgID string
}

// RetrieveCoreUserAccount retrieves Core user account
func (a *Adapter) RetrieveCoreUserAccount(token string) (*model.CoreAccount, error) {
	if len(token) > 0 {
		url := fmt.Sprintf("%s/services/account", token)
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: error creating load user data request - %s", err)
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: error loading user data - %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveCoreUserAccount: error with response code - %d", resp.StatusCode)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: error with response code != 200")
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: unable to parse json: %s", err)
		}

		var coreAccount model.CoreAccount
		err = json.Unmarshal(data, &coreAccount)
		if err != nil {
			log.Printf("RetrieveCoreUserAccount: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveAuthmanGroupMembersError: unable to parse json: %s", err)
		}

		return &coreAccount, nil
	}
	return nil, nil
}

// RetrieveCoreServices retrieves Core service registrations
func (a *Adapter) RetrieveCoreServices(serviceIDs []string) ([]model.CoreService, error) {
	if len(serviceIDs) > 0 {
		url := fmt.Sprintf("%s/bbs/service-regs?ids=%s", a.coreURL, strings.Join(serviceIDs, ","))
		client := &http.Client{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("RetrieveCoreServices: error creating load core service regs - %s", err)
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("RetrieveCoreServices: error loading core service regs data - %s", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("RetrieveCoreServices: error with response code - %d", resp.StatusCode)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: error with response code != 200")
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("RetrieveCoreServices: unable to read json: %s", err)
			return nil, fmt.Errorf("RetrieveCoreUserAccount: unable to parse json: %s", err)
		}

		var coreServices []model.CoreService
		err = json.Unmarshal(data, &coreServices)
		if err != nil {
			log.Printf("RetrieveCoreServices: unable to parse json: %s", err)
			return nil, fmt.Errorf("RetrieveCoreServices: unable to parse json: %s", err)
		}

		return coreServices, nil
	}
	return nil, nil
}

// GetAccountsByNetIDs retrieves accounts by net ids
func (a *Adapter) GetAccountsByNetIDs(netIDs []string) ([]model.CoreAccount, error) {
	searchParams := map[string]interface{}{
		"external_ids.net_id": netIDs,
	}
	return a.GetAccounts(searchParams)
}

// GetAccounts retrieves account for provided params
func (a *Adapter) GetAccounts(searchParams map[string]interface{}) ([]model.CoreAccount, error) {
	if a.serviceAccountManager == nil {
		log.Println("GetAccounts: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/accounts?app_id=%s&org_id=%s", a.coreURL, a.appID, a.orgID)

	bodyBytes, err := json.Marshal(searchParams)
	if err != nil {
		log.Printf("GetAccounts: error marshalling body - %s", err)
		return nil, err
	}

	respBody, err := a.makeRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	var items []model.CoreAccount
	err = json.Unmarshal(respBody, &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (a *Adapter) makeRequest(method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("gateway adapter: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, a.appID, a.orgID)
	if err != nil {
		log.Printf("gateway adapter: error sending request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("gateway adapter: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("gateway adapter: error with response code != 200 but %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("gateway adapter: unable to read json: %s", err)
		return nil, fmt.Errorf("gateway adapter: unable to parse json: %s", err)
	}
	return data, nil
}

// LoadDeletedMemberships loads deleted memberships
func (a *Adapter) LoadDeletedMemberships() ([]model.DeletedUserData, error) {

	if a.serviceAccountManager == nil {
		log.Println("LoadDeletedMemberships: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/deleted-memberships?service_id=%s", a.coreURL, a.serviceAccountManager.AuthService.ServiceID)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		a.logger.Errorf("delete membership: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
	if err != nil {
		log.Printf("LoadDeletedMemberships: error sending request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("LoadDeletedMemberships: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("LoadDeletedMemberships: error with response code != 200")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to read json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	var deletedMemberships []model.DeletedUserData
	err = json.Unmarshal(data, &deletedMemberships)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to parse json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	return deletedMemberships, nil
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(coreURL string, serviceAccountManager *auth.ServiceAccountManager, orgID string, appID string) *Adapter {
	return &Adapter{coreURL: coreURL, serviceAccountManager: serviceAccountManager, appID: appID, orgID: orgID}
}
