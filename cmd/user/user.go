package user

import (
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/denisbrodbeck/machineid"
	"github.com/spf13/viper"
)

const (
	INTERNAL_START_PARTITION     = 10 // exclusive
	EXPERIMENTAL_START_PARTITION = 20 // exclusive
	NB_OF_USER_PARTITIONS        = 10
	// internal commands are the commands with start partition number > INTERNAL_START_PARTITION
	INTERNAL_COMMAND_ENABLED_KEY = "INTERNAL_COMMAND_ENABLED"
	// experimental commands are the commands with start partition number > EXPERIMENTAL_START_PARTITION
	EXPERIMENTAL_COMMAND_ENABLED_KEY = "EXPERIMENTAL_COMMAND_ENABLED"
)

type User struct {
	UID                    string
	Partition              uint8
	InternalCmdEnabled     bool
	ExperimentalCmdEnabled bool
}

func GetUser() (User, error) {
	uid, err := machineid.ID()
	if err != nil {
		return User{}, err
	}

	return User{
		UID:                    uid,
		Partition:              uint8(helper.Hash(uid) % NB_OF_USER_PARTITIONS),
		InternalCmdEnabled:     viper.GetBool(INTERNAL_COMMAND_ENABLED_KEY),
		ExperimentalCmdEnabled: viper.GetBool(EXPERIMENTAL_COMMAND_ENABLED_KEY),
	}, nil

}

func (u User) InPartition(start uint8, end uint8) bool {
	if u.InternalCmdEnabled && start > INTERNAL_START_PARTITION && start < EXPERIMENTAL_START_PARTITION {
		return true
	}

	if u.ExperimentalCmdEnabled && start > EXPERIMENTAL_START_PARTITION {
		return true
	}

	return u.Partition >= start && u.Partition <= end
}
