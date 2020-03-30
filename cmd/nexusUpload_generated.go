// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type nexusUploadOptions struct {
	Version               string `json:"version,omitempty"`
	Url                   string `json:"url,omitempty"`
	Repository            string `json:"repository,omitempty"`
	GroupID               string `json:"groupId,omitempty"`
	ArtifactID            string `json:"artifactId,omitempty"`
	GlobalSettingsFile    string `json:"globalSettingsFile,omitempty"`
	M2Path                string `json:"m2Path,omitempty"`
	AdditionalClassifiers string `json:"additionalClassifiers,omitempty"`
	User                  string `json:"user,omitempty"`
	Password              string `json:"password,omitempty"`
}

// NexusUploadCommand Upload artifacts to Nexus
func NexusUploadCommand() *cobra.Command {
	metadata := nexusUploadMetadata()
	var stepConfig nexusUploadOptions
	var startTime time.Time

	var createNexusUploadCmd = &cobra.Command{
		Use:   "nexusUpload",
		Short: "Upload artifacts to Nexus",
		Long:  `Upload build artifacts to a Nexus Repository Manager`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			startTime = time.Now()
			log.SetStepName("nexusUpload")
			log.SetVerbose(GeneralConfig.Verbose)
			return PrepareConfig(cmd, &metadata, "nexusUpload", &stepConfig, config.OpenPiperFile)
		},
		Run: func(cmd *cobra.Command, args []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetry.Send(&telemetryData)
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, "nexusUpload")
			nexusUpload(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
		},
	}

	addNexusUploadFlags(createNexusUploadCmd, &stepConfig)
	return createNexusUploadCmd
}

func addNexusUploadFlags(cmd *cobra.Command, stepConfig *nexusUploadOptions) {
	cmd.Flags().StringVar(&stepConfig.Version, "version", "nexus3", "The Nexus Repository Manager version. Currently supported are 'nexus2' and 'nexus3'.")
	cmd.Flags().StringVar(&stepConfig.Url, "url", os.Getenv("PIPER_url"), "URL of the nexus. The scheme part of the URL will not be considered, because only http is supported.")
	cmd.Flags().StringVar(&stepConfig.Repository, "repository", os.Getenv("PIPER_repository"), "Name of the nexus repository.")
	cmd.Flags().StringVar(&stepConfig.GroupID, "groupId", os.Getenv("PIPER_groupId"), "Group ID of the artifacts. Only used in MTA projects, ignored for Maven.")
	cmd.Flags().StringVar(&stepConfig.ArtifactID, "artifactId", os.Getenv("PIPER_artifactId"), "The artifact ID used for both the .mtar and mta.yaml files deployed for MTA projects, ignored for Maven.")
	cmd.Flags().StringVar(&stepConfig.GlobalSettingsFile, "globalSettingsFile", os.Getenv("PIPER_globalSettingsFile"), "Path to the mvn settings file that should be used as global settings file.")
	cmd.Flags().StringVar(&stepConfig.M2Path, "m2Path", os.Getenv("PIPER_m2Path"), "The path to the local .m2 directory, only used for Maven projects.")
	cmd.Flags().StringVar(&stepConfig.AdditionalClassifiers, "additionalClassifiers", os.Getenv("PIPER_additionalClassifiers"), "List of additional classifiers that should be deployed to nexus. Each item is a map of a type and a classifier name.")
	cmd.Flags().StringVar(&stepConfig.User, "user", os.Getenv("PIPER_user"), "User")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password")

	cmd.MarkFlagRequired("url")
	cmd.MarkFlagRequired("repository")
}

// retrieve step metadata
func nexusUploadMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    "nexusUpload",
			Aliases: []config.Alias{{Name: "mavenExecute", Deprecated: false}},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Parameters: []config.StepParameters{
					{
						Name:        "version",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "nexus/version"}},
					},
					{
						Name:        "url",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "nexus/url"}},
					},
					{
						Name:        "repository",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "nexus/repository"}},
					},
					{
						Name:        "groupId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "nexus/groupId"}},
					},
					{
						Name:        "artifactId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "globalSettingsFile",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/globalSettingsFile"}},
					},
					{
						Name:        "m2Path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "maven/m2Path"}},
					},
					{
						Name:        "additionalClassifiers",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "nexus/additionalClassifiers"}},
					},
					{
						Name:        "user",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
					},
				},
			},
		},
	}
	return theMetaData
}
