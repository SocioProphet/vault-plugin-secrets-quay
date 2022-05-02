package quay

import (
	"fmt"

	qc "github.com/redhat-cop/vault-plugin-secrets-quay/client"
)

var (
	vaultCreator = fmt.Sprintf("%s-%s", Vault, TeamRoleCreator)
)

const (
	Vault string = "vault"
)

func (b *quayBackend) createRobot(client *client, robotName string, role *quayRoleEntry) (*qc.RobotAccount, error) {
	// Check if Account Exists
	robotAccount, existingRobotAccountResponse, apiError := client.GetRobotAccount(role.NamespaceType.String(), role.NamespaceName, robotName)

	if apiError.Error != nil {
		return nil, apiError.Error
		// A 400 response will be returned with a robot not found. If not, create it
	} else if existingRobotAccountResponse.StatusCode == 400 {

		// Create new Account
		robotAccount, _, apiError = client.CreateRobotAccount(role.NamespaceType.String(), role.NamespaceName, robotName)
		if apiError.Error != nil {
			return nil, apiError.Error
		}
	}

	if role.NamespaceType == organization {
		// Create Teams
		err := b.createAssignTeam(client, robotAccount.Name, role)

		if err != nil {
			return nil, err
		}

		// Create Default Permission
		if role.DefaultPermission != nil {
			organizationPrototypes, organizationPrototypesResponse, organizationPrototypesError := client.GetPrototypesByOrganization(role.NamespaceName)

			if organizationPrototypesError.Error != nil || organizationPrototypesResponse.StatusCode != 200 {
				return nil, organizationPrototypesError.Error
			}

			if found := isRobotAccountInPrototypeByRole(organizationPrototypes.Prototypes, robotAccount.Name, role.DefaultPermission.String()); !found {

				_, robotPrototypeResponse, robotPrototypeError := client.CreateRobotPermissionForOrganization(role.NamespaceName, robotAccount.Name, role.DefaultPermission.String())

				if robotPrototypeError.Error != nil || robotPrototypeResponse.StatusCode != 200 {
					return nil, robotPrototypeError.Error
				}

			}
		}

	}

	// Manage Repositories
	if role.Repositories != nil || role.DefaultPermission != nil {
		// Get Robot Permissions
		robotPermissions, robotPermissionsResponse, robotPermissionsError := client.GetRobotPermissions(role.NamespaceName, robotName)

		if robotPermissionsError.Error != nil || robotPermissionsResponse.StatusCode != 200 {
			return nil, robotPermissionsError.Error
		}

		// Get Repositories
		namespaceRepositories, namespaceRepositoriesResponse, namespaceRepositoriesError := client.GetRepositoriesForNamespace(role.NamespaceName)

		if namespaceRepositoriesError.Error != nil || namespaceRepositoriesResponse.StatusCode != 200 {
			return nil, robotPermissionsError.Error
		}

		// Loop through Quay repositories
		for _, namespaceRepository := range namespaceRepositories {

			var desiredPermission *Permission

			// Check if a Default Permission should be applied
			if role.DefaultPermission != nil {
				desiredPermission = role.DefaultPermission
			}

			// Check if explicit permission desired
			if role.Repositories != nil {
				// Deference repositories
				roleRepositories := *role.Repositories
				if rolePermission, ok := roleRepositories[namespaceRepository.Name]; ok {
					desiredPermission = &rolePermission
				}
			}

			if desiredPermission != nil {
				// Check to see if permission already exists on robot account
				if updatePermissions := shouldNeedUpdateRepositoryPermissions(namespaceRepository.Name, desiredPermission.String(), &robotPermissions.Permissions); updatePermissions {
					_, repositoryPermissionUpdateResponse, repositoryPermissionError := client.UpdateRepositoryUserPermission(role.NamespaceName, namespaceRepository.Name, robotName, desiredPermission.String())

					if repositoryPermissionError.Error != nil || repositoryPermissionUpdateResponse.StatusCode != 200 {
						return nil, repositoryPermissionError.Error
					}
				}

			}

		}

		/*
			for repositoryName, permission := range *role.Repositories {

				// Verify repository exists in the organization
				if updateRepository := repositoryExists(repositoryName, &namespaceRepositories); updateRepository {
					// Check to see if permission already exists on robot account
					if updatePermissions := shouldNeedUpdateRepositoryPermissions(repositoryName, permission.String(), &robotPermissions.Permissions); updatePermissions {
						_, repositoryPermissionUpdateResponse, repositoryPermissionError := client.UpdateRepositoryUserPermission(role.NamespaceName, repositoryName, robotName, permission.String())

						if repositoryPermissionError.Error != nil || repositoryPermissionUpdateResponse.StatusCode != 200 {
							return nil, repositoryPermissionError.Error
						}
					}
				}
			}
		*/
	}

	return &robotAccount, nil
}

func (b *quayBackend) deleteRobot(client *client, robotName string, role *quayRoleEntry) error {

	_, apiError := client.DeleteRobotAccount(role.NamespaceType.String(), role.NamespaceName, robotName)

	return apiError.Error
}

func (b *quayBackend) regenerateRobotPassword(client *client, robotName string, role *quayRoleEntry) (*qc.RobotAccount, error) {

	robotAccount, _, apiError := client.RegenerateRobotAccountPassword(role.NamespaceType.String(), role.NamespaceName, robotName)

	return &robotAccount, apiError.Error
}

func (b *quayBackend) createAssignTeam(client *client, robotName string, role *quayRoleEntry) error {

	teams := b.assembleTeams(role)

	for _, team := range teams {
		// Create Team
		_, _, err := client.CreateTeam(role.NamespaceName, team)

		if err.Error != nil {
			return err.Error
		}

		// Add member to team
		_, err = client.AddTeamMember(role.NamespaceName, team.Name, robotName)

		if err.Error != nil {
			return err.Error
		}

	}

	return nil
}

func (*quayBackend) assembleTeams(role *quayRoleEntry) map[string]*qc.Team {
	teams := map[string]*qc.Team{}

	// Build Teams
	if role.Teams != nil {
		for teamName, team := range *role.Teams {

			teams[teamName] = &qc.Team{
				Name: teamName,
				Role: qc.QuayTeamRole(team.String()),
			}

		}
	}

	// Create a Team called vault_creator for access to
	if role.CreateRepositories {
		teams[vaultCreator] = &qc.Team{
			Name: vaultCreator,
			Role: qc.QuayTeamRoleCreator,
		}
	}

	return teams
}

func isRobotAccountInPrototypeByRole(prototypes []qc.Prototype, robotAccount string, role string) bool {

	for _, prototype := range prototypes {

		if prototype.Role == role && prototype.Delegate.Robot == true && prototype.Delegate.Name == robotAccount {
			return true
		}

	}

	return false

}

func shouldNeedUpdateRepositoryPermissions(repositoryName string, repositoryPermission string, quayPermissions *[]qc.Permission) bool {

	for _, quayPermission := range *quayPermissions {
		if repositoryName == quayPermission.Repository.Name && repositoryPermission == quayPermission.Role.String() {
			return false
		}
	}

	return true
}

func repositoryExists(repositoryName string, repositories *[]qc.Repository) bool {

	for _, repository := range *repositories {
		if repositoryName == repository.Name {
			return true
		}
	}

	return false
}
