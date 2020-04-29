package parameters

import (
	"fmt"
	"math/rand"
	"time"

	sdkValidator "github.com/harmony-one/go-lib/staking/validator"
	"github.com/harmony-one/harmony-tf/utils"
	harmonyTypes "github.com/harmony-one/harmony/staking/types"
)

// GenerateUniqueDetails - generates new unique details to bypass uniqueness validation
func GenerateUniqueDetails(details *sdkValidator.ValidatorDetails) {
	details.Identity = generateUniqueProperty(details.Identity, harmonyTypes.MaxIdentityLength)
}

func generateUniqueProperty(property string, maxLength int) string {
	utcString := utils.FormattedTimeString(time.Now().UTC())
	randomString := fmt.Sprintf("%s-%d", utcString, rand.Intn(1000))
	newProperty := fmt.Sprintf("%s-%s", property, randomString)

	if len(newProperty) > maxLength {
		return fmt.Sprintf("%s%s", newProperty[0:(maxLength-len(randomString))], randomString)
	}

	return newProperty
}
