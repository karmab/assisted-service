package featuresupport

import (
	"fmt"

	"github.com/go-openapi/swag"
	"github.com/openshift/assisted-service/internal/common"
	"github.com/openshift/assisted-service/models"
	"github.com/thoas/go-funk"
)

var featuresList = map[models.FeatureSupportLevelID]SupportLevelFeature{
	models.FeatureSupportLevelIDSNO:                      (&SnoFeature{}).New(),
	models.FeatureSupportLevelIDVIPAUTOALLOC:             (&VipAutoAllocFeature{}).New(),
	models.FeatureSupportLevelIDCUSTOMMANIFEST:           (&CustomManifestFeature{}).New(),
	models.FeatureSupportLevelIDCLUSTERMANAGEDNETWORKING: (&ClusterManagedNetworkingFeature{}).New(),
	models.FeatureSupportLevelIDUSERMANAGEDNETWORKING:    (&UserManagedNetworkingFeature{}).New(),
	models.FeatureSupportLevelIDSINGLENODEEXPANSION:      (&SingleNodeExpansionFeature{}).New(),
	models.FeatureSupportLevelIDDUALSTACKVIPS:            (&DualStackVipsFeature{}).New(),
	models.FeatureSupportLevelIDLVM:                      (&LvmFeature{}).New(),
	models.FeatureSupportLevelIDNUTANIXINTEGRATION:       (&NutanixIntegrationFeature{}).New(),
	models.FeatureSupportLevelIDVSPHEREINTEGRATION:       (&VsphereIntegrationFeature{}).New(),
	models.FeatureSupportLevelIDCNV:                      (&CnvFeature{}).New(),
	models.FeatureSupportLevelIDODF:                      (&OdfFeature{}).New(),
	models.FeatureSupportLevelIDMINIMALISO:               (&MinimalIso{}).New(),
}

func getFeatureSupportList(features map[models.FeatureSupportLevelID]SupportLevelFeature, filters SupportLevelFilters) models.SupportLevels {
	featureSupportList := models.SupportLevels{}

	for _, feature := range features {
		featureID := feature.getId()

		if !isFeatureCompatibleWithArchitecture(feature, filters.OpenshiftVersion, swag.StringValue(filters.CPUArchitecture)) {
			featureSupportList[string(featureID)] = models.SupportLevelUnavailable
		} else {
			featureSupportList[string(featureID)] = feature.getSupportLevel(filters)
		}
	}
	return featureSupportList
}

// GetFeatureSupportList Get features support level list, cpuArchitecture is optional and the default value is x86
func GetFeatureSupportList(openshiftVersion string, cpuArchitecture *string) models.SupportLevels {
	filters := SupportLevelFilters{
		OpenshiftVersion: openshiftVersion,
		CPUArchitecture:  cpuArchitecture,
	}

	if cpuArchitecture == nil {
		filters.CPUArchitecture = swag.String(common.DefaultCPUArchitecture)
	}
	featuresSupportList := overrideInvalidRequest(featuresList, *filters.CPUArchitecture, openshiftVersion)
	if featuresSupportList == nil {
		return getFeatureSupportList(featuresList, filters)
	}

	return featuresSupportList
}

// IsFeatureAvailable Get the support level of a given feature, cpuArchitecture is optional
//with default value of x86_64
func IsFeatureAvailable(featureId models.FeatureSupportLevelID, openshiftVersion string, cpuArchitecture *string) bool {
	filters := SupportLevelFilters{
		OpenshiftVersion: openshiftVersion,
		CPUArchitecture:  cpuArchitecture,
	}

	if cpuArchitecture == nil {
		filters.CPUArchitecture = swag.String(common.DefaultCPUArchitecture)
	}

	return GetSupportLevel(featureId, filters) != models.SupportLevelUnavailable
}

func isFeatureCompatible(feature SupportLevelFeature, features ...SupportLevelFeature) *SupportLevelFeature {
	incompatibilities := feature.getIncompatibleFeatures()
	if incompatibilities != nil {
		for _, f := range features {
			if funk.Contains(*incompatibilities, f.getId()) {
				return &f
			}
		}
	}

	return nil
}

// isFeaturesCompatibleWithFeatures Determine if feature is compatible with other activated features
func isFeaturesCompatibleWithFeatures(activatedFeatures []SupportLevelFeature) error {
	for _, feature := range activatedFeatures {
		if incompatibleFeature := isFeatureCompatible(feature, activatedFeatures...); incompatibleFeature != nil {
			return fmt.Errorf("cannot use %s because it's not compatible with %s", feature.getName(), (*incompatibleFeature).getName())
		}
	}

	return nil
}

// isFeaturesCompatibleWithArchitecture Determine if feature is compatible with CPU architecture in a given openshift-version
func isFeaturesCompatibleWithArchitecture(openshiftVersion, cpuArchitecture string, activatedFeatures []SupportLevelFeature) error {
	for _, feature := range activatedFeatures {
		if !isFeatureCompatibleWithArchitecture(feature, openshiftVersion, cpuArchitecture) {
			return fmt.Errorf("cannot use %s because it's not compatible with the %s architecture "+
				"on version %s of OpenShift", feature.getName(), cpuArchitecture, openshiftVersion)
		}
	}
	return nil
}
