package syncman

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/spaceuptech/space-cloud/config"
	"github.com/spaceuptech/space-cloud/model"
	"github.com/spaceuptech/space-cloud/utils"
)

// SetGlobalConfig sets the global config. This must be called before the Start command.
func (s *SyncManager) SetGlobalConfig(c *config.Config) {
	s.lock.Lock()
	s.projectConfig = c
	s.lock.Unlock()
}

// GetGlobalConfig gets the global config
func (s *SyncManager) GetGlobalConfig() *config.Config {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.projectConfig
}

func makeRequest(method, token, url string, data *bytes.Buffer) error {

	// Create the http request
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return err
	}

	// Add token header
	req.Header.Add("Authorization", "Bearer "+token)

	// Create a http client and fire the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	m := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(m["error"].(string))
	}

	return nil
}

// SetStaticConfig applies the set project config command to the raft log
func (s *SyncManager) SetStaticConfig(token string, static *config.Static) error {
	// Acquire a lock
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.raft.VerifyLeader().Error() != nil {
		// Marshal json into byte array
		data, _ := json.Marshal(static)

		// Get the raft leader addr
		addr := strings.Split(string(s.raft.Leader()), ":")[0]

		// Make the http request
		return makeRequest("POST", token, "http://"+string(addr)+":4122/v1/api/config/static", bytes.NewBuffer(data))
	}

	// Create a raft command
	c := &model.RaftCommand{Kind: utils.RaftCommandSetStatic, Static: static}
	data, _ := json.Marshal(c)

	// Apply the command to the raft log
	return s.raft.Apply(data, 0).Error()
}

// AddInternalRoutes adds the provided routes to the internal routes
func (s *SyncManager) AddInternalRoutes(token string, static *config.Static) error {
	// Acquire a lock
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.raft.VerifyLeader().Error() != nil {
		// Marshal json into byte array
		data, _ := json.Marshal(static)

		// Get the raft leader addr
		addr := strings.Split(string(s.raft.Leader()), ":")[0]

		// Make the http request
		return makeRequest("POST", token, "http://"+string(addr)+":4122/v1/api/config/static/internal", bytes.NewBuffer(data))
	}

	// Create a raft command
	c := &model.RaftCommand{Kind: utils.RaftCommandAddInternalRouteOperation, Static: static}
	data, _ := json.Marshal(c)

	// Apply the command to the raft log
	return s.raft.Apply(data, 0).Error()
}

// SetOperationModeConfig applies the operation config to the raft log
func (s *SyncManager) SetOperationModeConfig(token string, op *config.OperationConfig) error {
	// Acquire a lock to make sure only a single operation occurs at any given point of time
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.raft.VerifyLeader().Error() != nil {
		// Marshal json into byte array
		data, _ := json.Marshal(op)

		// Get the raft leader addr
		addr := strings.Split(string(s.raft.Leader()), ":")[0]

		// Make the http request
		return makeRequest("POST", token, "http://"+string(addr)+":4122/v1/api/config/operation", bytes.NewBuffer(data))
	}

	// Create a raft command
	c := &model.RaftCommand{Kind: utils.RaftCommandSetOperation, Operation: op}
	data, _ := json.Marshal(c)

	// Apply the command to the raft log
	return s.raft.Apply(data, 0).Error()
}

// SetProjectConfig applies the config to the raft log
func (s *SyncManager) SetProjectConfig(token string, project *config.Project) error {
	// Acquire a lock to make sure only a single operation occurs at any given point of time
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.raft.VerifyLeader().Error() != nil {
		// Marshal json into byte array
		data, _ := json.Marshal(project)

		// Get the raft leader addr
		addr := strings.Split(string(s.raft.Leader()), ":")[0]

		// Make the http request
		return makeRequest("POST", token, "http://"+string(addr)+":4122/v1/api/config/projects", bytes.NewBuffer(data))
	}

	// Validate the operation
	if !s.adminMan.ValidateSyncOperation(s.projectConfig, project) {
		return errors.New("Please upgrade your instance")
	}

	// Create a raft command
	c := &model.RaftCommand{Kind: utils.RaftCommandSet, Project: project, ID: project.ID}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// Apply the command to the raft log
	return s.raft.Apply(data, 0).Error()
}

// SetDeployConfig applies the config to the raft log
func (s *SyncManager) SetDeployConfig(token string, deploy *config.Deploy) error {
	// Acquire a lock to make sure only a single operation occurs at any given point of time
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.raft.VerifyLeader().Error() != nil {
		// Marshal json into byte array
		data, _ := json.Marshal(deploy)

		// Get the raft leader addr
		addr := strings.Split(string(s.raft.Leader()), ":")[0]

		// Make the http request
		return makeRequest("POST", token, "http://"+string(addr)+":4122/v1/api/config/deploy", bytes.NewBuffer(data))
	}

	// Create a raft command
	c := &model.RaftCommand{Kind: utils.RaftCommandSetDeploy, Deploy: deploy}
	data, _ := json.Marshal(c)

	// Apply the command to the raft log
	return s.raft.Apply(data, 0).Error()
}

// DeleteConfig applies the config to the raft log
func (s *SyncManager) DeleteConfig(token, projectID string) error {
	// Acquire a lock to make sure only a single operation occurs at any given point of time
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.raft.VerifyLeader().Error() != nil {

		// Get the raft leader addr
		addr := strings.Split(string(s.raft.Leader()), ":")[0]

		// Make the http request
		return makeRequest("DELETE", token, "http://"+string(addr)+":4122/v1/api/config/"+projectID, nil)
	}

	// Create a raft command
	c := &model.RaftCommand{Kind: utils.RaftCommandDelete, ID: projectID}
	data, _ := json.Marshal(c)

	// Apply the command to the raft log
	return s.raft.Apply(data, 0).Error()
}

// GetConfig returns the config present in the state
func (s *SyncManager) GetConfig(projectID string) (*config.Project, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Iterate over all projects stored
	for _, p := range s.projectConfig.Projects {
		if projectID == p.ID {
			return p, nil
		}
	}

	return nil, errors.New("Given project is not present in state")
}

// GetClusterSize returns the size of the cluster
func (s *SyncManager) GetClusterSize() int {
	return s.list.NumNodes()
}