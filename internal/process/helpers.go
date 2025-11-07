package process

// cleanupProgramUnsafe limpia las instancias muertas de un programa
func (m *Manager) cleanupProgramUnsafe(name string) int {
	instances, exists := m.processes[name]
	if !exists {
		return 0
	}

	activeInstances := []*ProcessInstance{}
	cleanedCount := 0

	for _, instance := range instances {
		if m.isActiveInstance(instance) {
			activeInstances = append(activeInstances, instance)
		} else {
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		if len(activeInstances) == 0 {
			delete(m.processes, name)
		} else {
			m.processes[name] = activeInstances
		}
	}

	return cleanedCount
}

// isActiveInstance verifica si una instancia está activa
func (m *Manager) isActiveInstance(instance *ProcessInstance) bool {
	return instance.State == StateRunning ||
		instance.State == StateStarting ||
		instance.State == StateRestarting
}

// countActiveInstances cuenta las instancias activas en una lista
func (m *Manager) countActiveInstances(instances []*ProcessInstance) int {
	count := 0
	for _, instance := range instances {
		if m.isActiveInstance(instance) {
			count++
		}
	}
	return count
}

// copyProcessMap crea una copia profunda del mapa de procesos
func (m *Manager) copyProcessMap() map[string][]*ProcessInstance {
	status := make(map[string][]*ProcessInstance)
	for name, instances := range m.processes {
		status[name] = make([]*ProcessInstance, len(instances))
		copy(status[name], instances)
	}
	return status
}
