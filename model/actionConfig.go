package model

import (
	"github.com/benpate/datatype"
)

// ActionConfig stores the configuration information for each Action that can be taken on a Stream
type ActionConfig struct {
	ActionID string         `json:"actionId"`
	Method   string         `json:"method"`
	States   []string       `json:"states"`
	Roles    []string       `json:"roles"`
	Commands []datatype.Map `json:"commands"`
}

// NewActionConfig returns a fully initialized ActionConfig object
func NewActionConfig() ActionConfig {
	return ActionConfig{
		States:   make([]string, 0),
		Roles:    make([]string, 0),
		Commands: make([]datatype.Map, 0),
	}
}

// UserCan returns TRUE if this action is permitted on a stream (using the provided authorization)
func (actionConfig ActionConfig) UserCan(stream *Stream, authorization *Authorization) bool {

	// If present, "States" limits the states where this action can take place
	if len(actionConfig.States) > 0 {
		// If states are present, then the current state MUST be included in the list.
		// Otherwise, reject this action.
		if !matchOne(actionConfig.States, stream.StateID) {
			return false
		}
	}

	// If present, "Roles" limits the user roles that can take this action
	if len(actionConfig.Roles) > 0 {

		// The user must have AT LEAST ONE of the named roles to take this action.
		// If not, reject this action.
		roles := stream.Roles(authorization)

		if !matchAny(roles, actionConfig.Roles) {
			return false
		}
	}

	// All filters have passed.  Allow this action.
	return true
}

// matchOne returns TRUE if the value matches one (or more) of the values in the slice
func matchOne(slice []string, value string) bool {
	for index := range slice {
		if slice[index] == value {
			return true
		}
	}

	return false
}

// matchAny returns TRUE if any of the values in slice1 are equal to any of the values in slice2
func matchAny(slice1 []string, slice2 []string) bool {

	for index1 := range slice1 {
		for index2 := range slice2 {
			if slice1[index1] == slice2[index2] {
				return true
			}
		}
	}

	return false
}
