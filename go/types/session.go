package types

import (
	"net/http"
	"strings"
)

// SessionVariables represents hasura session variables map
type SessionVariables map[string]string

// NewSessionVariables parse and create a session variables instance from string map
func NewSessionVariables(input map[string]string) SessionVariables {
	result := SessionVariables{}
	for k, v := range input {
		result[strings.ToLower(k)] = v
	}

	return result
}

// NewSessionVariables parse and create a session variables instance from http header
func NewSessionVariablesFromHeaders(header http.Header) SessionVariables {
	result := SessionVariables{}
	for k, v := range header {
		if len(v) > 0 {
			result[strings.ToLower(k)] = v[0]
		}
	}

	return result
}

// GetRole gets hasura role value
func (sv SessionVariables) GetRole() string {
	return sv.Get(XHasuraRole)
}

// IsRoleOf checks if the current role exists in the list
func (sv SessionVariables) IsRoleOf(roles ...string) bool {
	role := sv.GetRole()
	for _, r := range roles {
		if strings.EqualFold(role, r) {
			return true
		}
	}
	return false
}

// GetRole gets hasura role value
func (sv SessionVariables) GetRequestID() string {
	return sv.Get(XRequestId)
}

// IsAdmin checks if the current role is admin
func (sv SessionVariables) IsAdmin() bool {
	role := sv.GetRole()
	return role == RoleAdmin
}

// Clone clones a new session variables map
func (sv SessionVariables) Clone() SessionVariables {
	newVariables := SessionVariables{}
	for k, v := range sv {
		newVariables[k] = v
	}
	return newVariables
}

// FilterKey creates a new session variables map with values removed by keys
func (sv SessionVariables) FilterKey(key string, keys ...string) SessionVariables {
	keys = append(keys, key)
	newVariables := sv.Clone()
	for _, k := range keys {
		delete(newVariables, strings.ToLower(k))
	}

	return newVariables
}

// Set sets a session variable value.
func (sv *SessionVariables) Set(key string, value string) {
	(*sv)[strings.ToLower(key)] = value
}

// Del deletes a session variable by key.
func (sv *SessionVariables) Del(key string) {
	delete(*sv, strings.ToLower(key))
}

// Get gets the value associated with the given key.
// If there are no value associated with the key, Get returns "". It is case insensitive;
func (sv SessionVariables) Get(key string) string {
	value, _ := sv[strings.ToLower(key)]
	return value
}

// ToStringMap returns the raw string map
func (sv SessionVariables) ToStringMap() map[string]string {
	return sv
}
