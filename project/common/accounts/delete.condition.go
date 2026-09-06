package accountspg

import (
	credentialspg "github.com/Johncode248/Rodzic-Chat/project/common/credentials"
)

type DeleteCondition struct {
	DeleteCredentialMutation []*credentialspg.DeleteMutation
}
