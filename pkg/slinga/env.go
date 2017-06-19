package slinga

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"path/filepath"
)

// AptomiOject represents an aptomi entity, which gets stored in aptomi DB
type AptomiOject string

// AptomiResolutionDir is where results of the last run are stored
const aptomiCurrentRunDir = "last-run-results"

const (
	/*
		The following objects can be added to Aptomi
	*/

	// TypeCluster is k8s cluster or any other cluster
	TypeCluster AptomiOject = "cluster"

	// TypeService is service definitions
	TypeService AptomiOject = "service"

	// TypeContext is how service gets allocated
	TypeContext AptomiOject = "context"

	// TypeRules is global rules of the land
	TypeRules AptomiOject = "rules"

	// TypeDependencies is who requested what
	TypeDependencies AptomiOject = "dependencies"

	/*
		The following objects must be configured to point to external resources
	*/

	// TypeUsers is where users are stored (later in AD and LDAP)
	TypeUsers AptomiOject = "users"

	// TypeSecrets is where secret tokens are stored (later in Hashicorp Vault)
	TypeSecrets AptomiOject = "secrets"

	// TypeCharts is where binary charts/images are stored (later in external repo)
	TypeCharts AptomiOject = "charts"

	/*
		The following objects are generated by aptomi during or after dependency resolution via policy
	*/

	// TypePolicyRevision holds revision number for the last successful aptomi run
	TypeRevision AptomiOject = "revision"

	// TypePolicyResolution holds usage data for components/dependencies
	TypePolicyResolution AptomiOject = "db"

	// TypeLogs contains debug logs
	TypeLogs AptomiOject = "logs"

	// TypeGraphics contains images generated by graphviz
	TypeGraphics AptomiOject = "graphics"
)

// Return aptomi DB directory
func getAptomiEnvVarAsDir(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		debug.WithFields(log.Fields{
			"var": key,
		}).Fatal("Environment variable is not present. Must point to a directory")
	}
	if stat, err := os.Stat(value); err != nil || !stat.IsDir() {
		debug.WithFields(log.Fields{
			"var":       key,
			"directory": value,
			"error":     err,
		}).Fatal("Directory doesn't exist or error encountered")
	}
	return value
}

// GetAptomiBaseDir returns base directory, i.e. the value of APTOMI_DB environment variable
func GetAptomiBaseDir() string {
	return getAptomiEnvVarAsDir("APTOMI_DB")
}

// GetAptomiPolicyDir returns default aptomi policy dir
func GetAptomiPolicyDir() string {
	return filepath.Join(GetAptomiBaseDir(), "aptomi-demo")
}

// GetAptomiObjectFilePatternYaml returns file pattern for aptomi objects (so they can be loaded from those files)
func GetAptomiObjectFilePatternYaml(baseDir string, aptomiObject AptomiOject) string {
	return filepath.Join(baseDir, "**", string(aptomiObject)+"*.yaml")
}

// GetAptomiObjectFilePatternTgz returns file pattern for tgz objects (so they can be loaded from those files)
func GetAptomiObjectFilePatternTgz(baseDir string, aptomiObject AptomiOject, chartName string) string {
	return filepath.Join(baseDir, "**", chartName+".tgz")
}

// GetAptomiObjectWriteFileCurrentRun returns file name for global aptomi objects (e.g. revision)
// It will place files into aptomiCurrentRunDir. It will create the corresponding directories if they don't exist
func GetAptomiObjectWriteFileGlobal(baseDir string, aptomiObject AptomiOject) string {
	return filepath.Join(baseDir, string(aptomiObject)+".yaml")
}

// GetAptomiObjectWriteFileCurrentRun returns file name for aptomi objects (so they can be saved)
// It will place files into aptomiCurrentRunDir. It will create the corresponding directories if they don't exist
func GetAptomiObjectFileFromRun(baseDir string, revision AptomiRevision, aptomiObject AptomiOject, fileName string) string {
	return filepath.Join(baseDir, revision.getRunDirectory(), string(aptomiObject), fileName)
}

	// GetAptomiObjectWriteFileCurrentRun returns file name for aptomi objects (so they can be saved)
// It will place files into aptomiCurrentRunDir. It will create the corresponding directories if they don't exist
func GetAptomiObjectWriteFileCurrentRun(baseDir string, aptomiObject AptomiOject, fileName string) string {
	dir := filepath.Join(baseDir, aptomiCurrentRunDir, string(aptomiObject))
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			debug.WithFields(log.Fields{
				"directory": dir,
				"error":     err,
			}).Fatal("Directory can't be created or error encountered")
		}
	}
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		debug.WithFields(log.Fields{
			"directory": dir,
			"error":     err,
		}).Fatal("Directory can't be created or error encountered")
	}
	return filepath.Join(dir, fileName)
}

// ClearCurrentResolutionDirectory cleans contents of a "current run" directory
func PrepareCurrentRunDirectory(baseDir string) {
	dir := filepath.Join(baseDir, aptomiCurrentRunDir)
	err := deleteDirectoryContents(dir)
	if err != nil {
		debug.WithFields(log.Fields{
			"directory": dir,
			"error":     err,
		}).Fatal("Directory contents can't be deleted")
	}
}
