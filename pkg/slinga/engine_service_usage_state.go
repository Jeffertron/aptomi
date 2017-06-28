package slinga

import (
	. "github.com/Frostman/aptomi/pkg/slinga/db"
	. "github.com/Frostman/aptomi/pkg/slinga/fileio"
	. "github.com/Frostman/aptomi/pkg/slinga/language"
	. "github.com/Frostman/aptomi/pkg/slinga/language/yaml"
	. "github.com/Frostman/aptomi/pkg/slinga/maputil"
	"time"
)

// ServiceUsageState contains resolution data for services - who is using what, as well as contains processing order and additional data
type ServiceUsageState struct {
	// reference to a policy
	Policy *Policy

	// reference to dependencies
	Dependencies *GlobalDependencies

	// reference to users
	users *GlobalUsers

	// Date when it was created
	CreatedOn time.Time

	// Diff stored as text
	DiffAsText string

	// resolved usage - stores full information about dependencies which have been successfully resolved. should ideally be accessed by a getter
	ResolvedData *ServiceUsageData

	// unresolved usage - stores full information about dependencies which were not resolved. including rule logs with reasons, etc
	UnresolvedData *ServiceUsageData
}

// ServiceUsageData contains all the data that gets resolved for one or more dependencies
// When adding new fields to this object, it's crucial to modify appendData() method as well (!)
type ServiceUsageData struct {
	// resolved component instances: componentKey -> componentInstance
	ComponentInstanceMap map[string]*ComponentInstance

	// resolved component processing order in which components/services have to be processed
	componentProcessingOrderHas map[string]bool
	ComponentProcessingOrder    []string
}

// NewResolvedServiceUsageData creates new empty ServiceUsageData
func newServiceUsageData() *ServiceUsageData {
	return &ServiceUsageData{
		ComponentInstanceMap:        make(map[string]*ComponentInstance),
		componentProcessingOrderHas: make(map[string]bool),
		ComponentProcessingOrder:    []string{},
	}
}

// NewServiceUsageState creates new empty ServiceUsageState
func NewServiceUsageState(policy *Policy, dependencies *GlobalDependencies, users *GlobalUsers) ServiceUsageState {
	return ServiceUsageState{
		Policy:         policy,
		Dependencies:   dependencies,
		users:          users,
		CreatedOn:      time.Now(),
		ResolvedData:   newServiceUsageData(),
		UnresolvedData: newServiceUsageData(),
	}
}

// GetResolvedData returns usage.ResolvedData
// TODO: we can get likely rid of this method (but need to check serialization, etc)
func (state *ServiceUsageState) GetResolvedData() *ServiceUsageData {
	if state.ResolvedData == nil {
		state.ResolvedData = newServiceUsageData()
	}
	return state.ResolvedData
}

// Gets a component instance entry or creates an new entry if it doesn't exist
func (data *ServiceUsageData) getComponentInstanceEntry(key string) *ComponentInstance {
	if _, ok := data.ComponentInstanceMap[key]; !ok {
		data.ComponentInstanceMap[key] = newComponentInstance(key)
	}
	return data.ComponentInstanceMap[key]
}

// Record dependency for component instance
func (data *ServiceUsageData) recordResolvedAndDependency(key string, dependency *Dependency) {
	data.getComponentInstanceEntry(key).setResolved(true)
	data.getComponentInstanceEntry(key).addDependency(dependency.ID)
}

// Record processing order for component instance
func (data *ServiceUsageData) recordProcessingOrder(key string) {
	if !data.componentProcessingOrderHas[key] {
		data.componentProcessingOrderHas[key] = true
		data.ComponentProcessingOrder = append(data.ComponentProcessingOrder, key)
	}
}

// Stores calculated discovery params for component instance
func (data *ServiceUsageData) recordCodeParams(key string, codeParams NestedParameterMap) {
	data.getComponentInstanceEntry(key).addCodeParams(codeParams)
}

// Stores calculated discovery params for component instance
func (data *ServiceUsageData) recordDiscoveryParams(key string, discoveryParams NestedParameterMap) {
	data.getComponentInstanceEntry(key).addDiscoveryParams(discoveryParams)
}

// Stores calculated labels for component instance
func (data *ServiceUsageData) recordLabels(key string, labels LabelSet) {
	data.getComponentInstanceEntry(key).addLabels(labels)

}

// Stores an outgoing edge for component instance as we are traversing the graph
func (data *ServiceUsageData) storeEdge(key string, keyDst string) {
	// Arrival key can be empty at the very top of the recursive function in engine, so let's check for that
	if len(key) > 0 && len(keyDst) > 0 {
		data.getComponentInstanceEntry(key).addEdgeOut(keyDst)
		data.getComponentInstanceEntry(keyDst).addEdgeIn(key)
	}
}

// Stores rule log entry, attaching it to component instance by dependency
func (data *ServiceUsageData) storeRuleLogEntry(key string, dependency *Dependency, entry *RuleLogEntry) {
	data.getComponentInstanceEntry(key).addRuleLogEntries(dependency.ID, entry)
}

// Appends data to the current ServiceUsageData
func (data *ServiceUsageData) appendData(ops *ServiceUsageData) {
	for key, instance := range ops.ComponentInstanceMap {
		data.getComponentInstanceEntry(key).appendData(instance)
	}
	for _, key := range ops.ComponentProcessingOrder {
		data.recordProcessingOrder(key)
	}
}

// LoadServiceUsageState loads usage state from a file under Aptomi DB
func LoadServiceUsageState() ServiceUsageState {
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), lastRevision.GetRunDirectory(), TypePolicyResolution, "db.yaml")
	return loadServiceUsageStateFromFile(fileName)
}

// LoadServiceUsageStatesAll loads all usage states from files under Aptomi DB
func LoadServiceUsageStatesAll() map[int]ServiceUsageState {
	result := make(map[int]ServiceUsageState)
	lastRevision := GetLastRevision(GetAptomiBaseDir())
	for rev := lastRevision; rev > LastRevisionAbsentValue; rev-- {
		fileName := GetAptomiObjectFileFromRun(GetAptomiBaseDir(), rev.GetRunDirectory(), TypePolicyResolution, "db.yaml")
		state := loadServiceUsageStateFromFile(fileName)
		if state.Policy != nil {
			// add only non-empty revisions. don't add revision which got deleted
			result[int(rev)] = state
		}
	}
	return result
}

// SaveServiceUsageState saves usage state in a file under Aptomi DB
func (state ServiceUsageState) SaveServiceUsageState() {
	fileName := GetAptomiObjectWriteFileCurrentRun(GetAptomiBaseDir(), TypePolicyResolution, "db.yaml")
	SaveObjectToFile(fileName, state)
}

// Loads usage state from file
func loadServiceUsageStateFromFile(fileName string) ServiceUsageState {
	return *LoadObjectFromFileDefaultEmpty(fileName, new(ServiceUsageState)).(*ServiceUsageState)
}
