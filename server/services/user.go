package services

import (
	"errors"
	"fmt"
	"leetcodeduels/store"
	"leetcodeduels/util"
)

func GenerateUniqueDiscriminator(username string) (string, error) {
	maxRetries := 10
	for range maxRetries {
		num := util.RandInt(1, 9999)
		discriminator := fmt.Sprintf("%04d", num)

		exists, err := store.DataStore.DiscriminatorExists(username, discriminator)
		if err != nil {
			return "", err
		}
		if !exists {
			return discriminator, nil
		}
	}
	// todo: figure out what to do if we can't find a unique one after several tries
	return "", errors.New("could not generate a unique discriminator for provided username")
}

func GenerateUniqueDiscriminator_AlphaNumeric(username string) (string, error) {
	maxRetries := 10
	for range maxRetries {
		discriminator := util.RandAlphaNumericString(4)
		exists, err := store.DataStore.DiscriminatorExists(username, discriminator)
		if err != nil {
			return "", err
		}
		if !exists {
			return discriminator, nil
		}
	}
	return "", errors.New("could not generate a unique discriminator for provided username")
}
