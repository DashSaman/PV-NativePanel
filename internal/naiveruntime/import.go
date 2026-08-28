package naiveruntime

import (
	"fmt"

	"github.com/DashSaman/PV-NaivePanel/internal/runtimecred"
)

// CredentialsForImport returns the parsed live credentials as internal domain
// values whose passwords are not JSON fields. InspectCaddyfile already
// validated each username/password pair; this method exists only for the
// privileged local import path.
func (i Inspection) CredentialsForImport() ([]runtimecred.DesiredCredential, error) {
	credentials := make([]runtimecred.DesiredCredential, 0, len(i.credentials))
	for index, parsed := range i.credentials {
		credential, err := runtimecred.NewImportedDesiredCredential(
			fmt.Sprintf("import-%d", index+1),
			parsed.username,
			parsed.password,
			runtimecred.CredentialActive,
		)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}
	return credentials, nil
}
