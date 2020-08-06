package parameters

import (
	"fmt"
	"math/rand"
	"time"

	harmonyTypes "github.com/harmony-one/harmony/staking/types"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
	"github.com/hyperion-hyn/hyperion-tf/utils"
)

// GenerateUniqueDetails - generates new unique details to bypass uniqueness validation
func GenerateUniqueDetails(details *sdkValidator.ValidatorDetails) {
	if len(details.Identity) > 0 && len(details.Identity) <= harmonyTypes.MaxIdentityLength {
		details.Identity = generateUniqueProperty(details.Identity, harmonyTypes.MaxIdentityLength)
	}
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
