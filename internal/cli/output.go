package cli

func (i *invocation) queryResult(value any, err error, jsonOutput bool) ([]byte, bool, error) {
	if err != nil {
		return nil, false, err
	}
	data, err := i.formatResult(value, jsonOutput)
	return data, false, err
}
