package hooks

// GetSection returns a section by its type
func (cm *CollapsibleManager) GetSection(sectionType SectionType) *Section {
	for i := range cm.Sections {
		if cm.Sections[i].Type == sectionType {
			return &cm.Sections[i]
		}
	}
	return nil
}

// GetSectionHeaderIndex returns the visual index of a section header
func (cm *CollapsibleManager) GetSectionHeaderIndex(sectionType SectionType) int {
	var currentIndex int = 0

	for _, section := range cm.Sections {
		if section.Type == sectionType {
			return currentIndex
		}
		
		currentIndex++ // Count the section header
		
		// Skip section items if expanded
		if section.IsExpanded {
			currentIndex += section.ItemCount
		}
	}
	
	return -1 // Section not found
}
