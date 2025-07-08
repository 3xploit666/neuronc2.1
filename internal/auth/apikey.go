package auth

import (
	"neuronc2/internal/utils"
)

func GenerateAPIKey() string {
	return utils.GenerateSecureAPIKey()
}
