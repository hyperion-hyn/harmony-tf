package parameters

import (
	"fmt"
	"github.com/hyperion-hyn/hyperion-tf/extension/go-lib/microstake/map3node"
	"math/rand"
	"time"

	restakingTypes "github.com/ethereum/go-ethereum/staking/types/restaking"
	sdkValidator "github.com/hyperion-hyn/hyperion-tf/extension/go-lib/staking/validator"
	"github.com/hyperion-hyn/hyperion-tf/utils"
)

// GenerateUniqueDetails - generates new unique details to bypass uniqueness validation
func GenerateUniqueDetails(details *sdkValidator.ValidatorDetails) {
	if len(details.Identity) > 0 && len(details.Identity) <= restakingTypes.MaxIdentityLength {
		details.Identity = generateUniqueProperty(details.Identity, restakingTypes.MaxIdentityLength)
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

// GenerateUniqueDetails - generates new unique details to bypass uniqueness validation
func GenerateMap3NodeUniqueDetails(details *map3node.Map3NodeDetails) {
	if len(details.Identity) > 0 && len(details.Identity) <= restakingTypes.MaxIdentityLength {
		details.Identity = generateUniqueProperty(details.Identity, restakingTypes.MaxIdentityLength)
	}
}
