package maven

import (
	"bytes"
	"fmt"
	"github.com/bmatcuk/doublestar"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

// ExecuteOptions are used by Execute() to construct the Maven command line.
type ExecuteOptions struct {
	PomPath                     string   `json:"pomPath,omitempty"`
	ProjectSettingsFile         string   `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile          string   `json:"globalSettingsFile,omitempty"`
	M2Path                      string   `json:"m2Path,omitempty"`
	Goals                       []string `json:"goals,omitempty"`
	Defines                     []string `json:"defines,omitempty"`
	Flags                       []string `json:"flags,omitempty"`
	LogSuccessfulMavenTransfers bool     `json:"logSuccessfulMavenTransfers,omitempty"`
	ReturnStdout                bool     `json:"returnStdout,omitempty"`
}

// EvaluateOptions are used by Evaluate() to construct the Maven command line.
// In contrast to ExecuteOptions, fewer settings are required for Evaluate and thus a separate type is needed.
type EvaluateOptions struct {
	PomPath             string `json:"pomPath,omitempty"`
	ProjectSettingsFile string `json:"projectSettingsFile,omitempty"`
	GlobalSettingsFile  string `json:"globalSettingsFile,omitempty"`
	M2Path              string `json:"m2Path,omitempty"`
}

type mavenExecRunner interface {
	Stdout(out io.Writer)
	Stderr(err io.Writer)
	RunExecutable(e string, p ...string) error
}

type mavenUtils interface {
	FileExists(path string) (bool, error)
	DownloadFile(url, filename string, header http.Header, cookies []*http.Cookie) error
	glob(pattern string) (matches []string, err error)
	getwd() (dir string, err error)
	chdir(dir string) error
}

type utilsBundle struct {
	*piperhttp.Client
	*piperutils.Files
}

func newUtils() *utilsBundle {
	return &utilsBundle{
		Client: &piperhttp.Client{},
		Files:  &piperutils.Files{},
	}
}

func (u *utilsBundle) glob(pattern string) (matches []string, err error) {
	return doublestar.Glob(pattern)
}

func (u *utilsBundle) getwd() (dir string, err error) {
	return os.Getwd()
}

func (u *utilsBundle) chdir(dir string) error {
	return os.Chdir(dir)
}

const mavenExecutable = "mvn"

// Execute constructs a mvn command line from the given options, and uses the provided
// mavenExecRunner to execute it.
func Execute(options *ExecuteOptions, command mavenExecRunner) (string, error) {
	stdOutBuf, stdOut := evaluateStdOut(options)
	command.Stdout(stdOut)
	command.Stderr(log.Writer())

	parameters, err := getParametersFromOptions(options, newUtils())
	if err != nil {
		return "", fmt.Errorf("failed to construct parameters from options: %w", err)
	}

	err = command.RunExecutable(mavenExecutable, parameters...)
	if err != nil {
		commandLine := append([]string{mavenExecutable}, parameters...)
		return "", fmt.Errorf("failed to run executable, command: '%s', error: %w", commandLine, err)
	}

	if stdOutBuf == nil {
		return "", nil
	}
	return string(stdOutBuf.Bytes()), nil
}

// Evaluate constructs ExecuteOptions for using the maven-help-plugin's 'evaluate' goal to
// evaluate a given expression from a pom file. This allows to retrieve the value of - for
// example - 'project.version' from a pom file exactly as Maven itself evaluates it.
func Evaluate(options *EvaluateOptions, expression string, command mavenExecRunner) (string, error) {
	expressionDefine := "-Dexpression=" + expression
	executeOptions := ExecuteOptions{
		PomPath:             options.PomPath,
		M2Path:              options.M2Path,
		ProjectSettingsFile: options.ProjectSettingsFile,
		GlobalSettingsFile:  options.GlobalSettingsFile,
		Goals:               []string{"org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate"},
		Defines:             []string{expressionDefine, "-DforceStdout", "-q"},
		ReturnStdout:        true,
	}
	value, err := Execute(&executeOptions, command)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(value, "null object or invalid expression") {
		return "", fmt.Errorf("expression '%s' in file '%s' could not be resolved", expression, options.PomPath)
	}
	return value, nil
}

// InstallFile installs a maven artifact and its pom into the local maven repository.
// If "file" is empty, only the pom is installed. "pomFile" must not be empty.
func InstallFile(file, pomFile, m2Path string, command mavenExecRunner) error {
	if len(pomFile) == 0 {
		return fmt.Errorf("pomFile can't be empty")
	}

	var defines []string
	if len(file) > 0 {
		defines = append(defines, "-Dfile="+file)
		if strings.Contains(file, ".jar") {
			defines = append(defines, "-Dpackaging=jar")
		}
		if strings.Contains(file, "-classes") {
			defines = append(defines, "-Dclassifier=classes")
		}

	} else {
		defines = append(defines, "-Dfile="+pomFile)
	}
	defines = append(defines, "-DpomFile="+pomFile)
	mavenOptionsInstall := ExecuteOptions{
		Goals:   []string{"install:install-file"},
		Defines: defines,
		PomPath: pomFile,
		M2Path:  m2Path,
	}
	_, err := Execute(&mavenOptionsInstall, command)
	if err != nil {
		return fmt.Errorf("failed to install maven artifacts: %w", err)
	}
	return nil
}

// InstallMavenArtifacts finds maven modules (identified by pom.xml files) and installs the artifacts into the local maven repository.
func InstallMavenArtifacts(command mavenExecRunner, options EvaluateOptions) error {
	return doInstallMavenArtifacts(command, options, newUtils())
}

func doInstallMavenArtifacts(command mavenExecRunner, options EvaluateOptions, utils mavenUtils) error {
	err := flattenPom(command)
	if err != nil {
		return err
	}

	pomFiles, err := utils.glob(filepath.Join("**", "pom.xml"))
	if err != nil {
		return err
	}

	oldWorkingDirectory, err := utils.getwd()
	if err != nil {
		return err
	}

	// Set pom path fix here because we will change into the respective pom's directory
	options.PomPath = "pom.xml"
	for _, pomFile := range pomFiles {
		log.Entry().Info("Installing maven artifacts from module: " + pomFile)
		dir := path.Dir(pomFile)
		err = utils.chdir(dir)
		if err != nil {
			return err
		}

		packaging, err := Evaluate(&options, "project.packaging", command)
		if err != nil {
			return err
		}

		if packaging == "pom" {
			err = InstallFile("", "pom.xml", options.M2Path, command)
			if err != nil {
				return err
			}
		} else {
			err = installJarWarArtifacts(command, utils, options)
			if err != nil {
				return err
			}
		}

		err = utils.chdir(oldWorkingDirectory)
		if err != nil {
			return err
		}
	}
	return err
}

func installJarWarArtifacts(command mavenExecRunner, utils mavenUtils, options EvaluateOptions) error {
	finalName, err := Evaluate(&options, "project.build.finalName", command)
	if err != nil {
		return err
	}
	if finalName == "" {
		log.Entry().Warn("project.build.finalName is empty, skipping install of artifact. Installing only the pom file.")
		err = InstallFile("", "pom.xml", options.M2Path, command)
		if err != nil {
			return err
		}
		return nil
	}
	jarExists, _ := utils.FileExists(jarFile(finalName))
	warExists, _ := utils.FileExists(warFile(finalName))
	classesJarExists, _ := utils.FileExists(classesJarFile(finalName))

	if jarExists {
		err = InstallFile(jarFile(finalName), "pom.xml", options.M2Path, command)
		if err != nil {
			return err
		}
	}

	if warExists {
		err = InstallFile(warFile(finalName), "pom.xml", options.M2Path, command)
		if err != nil {
			return err
		}
	}

	if classesJarExists {
		err = InstallFile(classesJarFile(finalName), "pom.xml", options.M2Path, command)
		if err != nil {
			return err
		}
	}
	return nil
}

func jarFile(finalName string) string {
	return "target/" + finalName + ".jar"
}

func classesJarFile(finalName string) string {
	return "target/" + finalName + "-classes.jar"
}

func warFile(finalName string) string {
	return "target/" + finalName + ".war"
}

func flattenPom(command mavenExecRunner) error {
	mavenOptionsFlatten := ExecuteOptions{
		Goals:   []string{"flatten:flatten"},
		Defines: []string{"-Dflatten.mode=resolveCiFriendliesOnly"},
		PomPath: "pom.xml",
	}
	_, err := Execute(&mavenOptionsFlatten, command)
	return err
}

func evaluateStdOut(options *ExecuteOptions) (*bytes.Buffer, io.Writer) {
	var stdOutBuf *bytes.Buffer
	stdOut := log.Writer()
	if options.ReturnStdout {
		stdOutBuf = new(bytes.Buffer)
		stdOut = io.MultiWriter(stdOut, stdOutBuf)
	}
	return stdOutBuf, stdOut
}

func getParametersFromOptions(options *ExecuteOptions, utils mavenUtils) ([]string, error) {
	var parameters []string

	if options.GlobalSettingsFile != "" {
		globalSettingsFileName, err := downloadSettingsIfURL(options.GlobalSettingsFile, ".pipeline/mavenGlobalSettings.xml", utils)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, "--global-settings", globalSettingsFileName)
	}

	if options.ProjectSettingsFile != "" {
		projectSettingsFileName, err := downloadSettingsIfURL(options.ProjectSettingsFile, ".pipeline/mavenProjectSettings.xml", utils)
		if err != nil {
			return nil, err
		}
		parameters = append(parameters, "--settings", projectSettingsFileName)
	}

	if options.M2Path != "" {
		parameters = append(parameters, "-Dmaven.repo.local="+options.M2Path)
	}

	if options.PomPath != "" {
		parameters = append(parameters, "--file", options.PomPath)
	}

	if options.Flags != nil {
		parameters = append(parameters, options.Flags...)
	}

	if options.Defines != nil {
		parameters = append(parameters, options.Defines...)
	}

	if !options.LogSuccessfulMavenTransfers {
		parameters = append(parameters, "-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn")
	}

	parameters = append(parameters, "--batch-mode")

	parameters = append(parameters, options.Goals...)

	return parameters, nil
}

func downloadSettingsIfURL(settingsFileOption, settingsFile string, utils mavenUtils) (string, error) {
	result := settingsFileOption
	if strings.HasPrefix(settingsFileOption, "http:") || strings.HasPrefix(settingsFileOption, "https:") {
		err := downloadSettingsFromURL(settingsFileOption, settingsFile, utils)
		if err != nil {
			return "", err
		}
		result = settingsFile
	}
	return result, nil
}

// ToDo replace with pkg/maven/settings GetSettingsFile
func downloadSettingsFromURL(url, filename string, utils mavenUtils) error {
	exists, _ := utils.FileExists(filename)
	if exists {
		log.Entry().Infof("Not downloading maven settings file, because it already exists at '%s'", filename)
		return nil
	}
	err := utils.DownloadFile(url, filename, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to download maven settings from URL '%s' to file '%s': %w",
			url, filename, err)
	}
	return nil
}

func GetTestModulesExcludes() []string {
	return getTestModulesExcludes(newUtils())
}

func getTestModulesExcludes(utils mavenUtils) []string {
	var excludes []string
	exists, _ := utils.FileExists("unit-tests/pom.xml")
	if exists {
		excludes = append(excludes, "-pl", "!unit-tests")
	}
	exists, _ = utils.FileExists("integration-tests/pom.xml")
	if exists {
		excludes = append(excludes, "-pl", "!integration-tests")
	}
	return excludes
}
