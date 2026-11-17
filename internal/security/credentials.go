package security

import (
	"fmt"

	"github.com/danieljoos/wincred"
)

const TargetName = "IFMChat/APIKey"
const TargetTavilyKey = "IFMChat/TavilyKey"

type CredentialService struct{}

func NewCredentialService() *CredentialService {
	return &CredentialService{}
}

// SetTavilyKey stores Tavily API key in Windows Credential Manager
func (s *CredentialService) SetTavilyKey(key string) error {
	cred := wincred.NewGenericCredential(TargetTavilyKey)
	cred.CredentialBlob = []byte(key)
	cred.Persist = wincred.PersistLocalMachine
	cred.UserName = "IFMChatUser"
	if err := cred.Write(); err != nil {
		return fmt.Errorf("failed storing Tavily key: %w", err)
	}
	return nil
}

// GetTavilyKey retrieves Tavily API key from Windows Credential Manager
func (s *CredentialService) GetTavilyKey() (string, error) {
	cred, err := wincred.GetGenericCredential(TargetTavilyKey)
	if err != nil || cred == nil || len(cred.CredentialBlob) == 0 {
		return "", nil
	}
	return string(cred.CredentialBlob), nil
}

// DeleteTavilyKey removes Tavily API key
func (s *CredentialService) DeleteTavilyKey() error {
	cred, err := wincred.GetGenericCredential(TargetTavilyKey)
	if err != nil {
		return nil
	}
	return cred.Delete()
}

// SetAPIKey securely stores the IFM API Key in Windows Credential Manager
func (s *CredentialService) SetAPIKey(apiKey string) error {
	cred := wincred.NewGenericCredential(TargetName)
	cred.CredentialBlob = []byte(apiKey)
	cred.Persist = wincred.PersistLocalMachine
	cred.UserName = "IFMChatUser"
	if err := cred.Write(); err != nil {
		return fmt.Errorf("failed storing API key in Windows Credential Manager: %w", err)
	}
	return nil
}

// GetAPIKey retrieves the IFM API Key securely from Windows Credential Manager
func (s *CredentialService) GetAPIKey() (string, error) {
	cred, err := wincred.GetGenericCredential(TargetName)
	if err != nil {
		// wincred returns error if not found
		return "", nil
	}
	if cred == nil || len(cred.CredentialBlob) == 0 {
		return "", nil
	}
	return string(cred.CredentialBlob), nil
}

// HasAPIKey checks whether an API key exists without exposing its content
func (s *CredentialService) HasAPIKey() bool {
	key, err := s.GetAPIKey()
	return err == nil && len(key) > 0
}

// DeleteAPIKey removes the API key from Windows Credential Manager
func (s *CredentialService) DeleteAPIKey() error {
	cred, err := wincred.GetGenericCredential(TargetName)
	if err != nil {
		return nil // Not found, nothing to delete
	}
	if err := cred.Delete(); err != nil {
		return fmt.Errorf("failed deleting API key from Windows Credential Manager: %w", err)
	}
	return nil
}

// SetKeyWithID stores a named key by ID in Windows Credential Manager
func (s *CredentialService) SetKeyWithID(id, apiKey string) error {
	target := "IFMChat/Key_" + id
	cred := wincred.NewGenericCredential(target)
	cred.CredentialBlob = []byte(apiKey)
	cred.Persist = wincred.PersistLocalMachine
	cred.UserName = "IFMChatUser"
	if err := cred.Write(); err != nil {
		return fmt.Errorf("failed storing key %s in Windows Credential Manager: %w", id, err)
	}
	return nil
}

// GetKeyWithID retrieves a specific key by ID from Windows Credential Manager
func (s *CredentialService) GetKeyWithID(id string) (string, error) {
	target := "IFMChat/Key_" + id
	cred, err := wincred.GetGenericCredential(target)
	if err != nil {
		return "", nil
	}
	if cred == nil || len(cred.CredentialBlob) == 0 {
		return "", nil
	}
	return string(cred.CredentialBlob), nil
}

// DeleteKeyWithID removes a specific key by ID from Windows Credential Manager
func (s *CredentialService) DeleteKeyWithID(id string) error {
	target := "IFMChat/Key_" + id
	cred, err := wincred.GetGenericCredential(target)
	if err != nil {
		return nil
	}
	return cred.Delete()
}
