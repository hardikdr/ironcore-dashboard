package model

import (
	"fmt"
	"strings"
)

const missingKeySegment = "_"

// ResourceKey is the stable, globally unique identity of a normalized record.
// Its canonical form is source/group/kind/namespace/name/uid.
type ResourceKey string

type ResourceIdentity struct {
	SourceID  string
	Group     string
	Kind      string
	Namespace string
	Name      string
	UID       string
}

func NewResourceKey(identity ResourceIdentity) (ResourceKey, error) {
	for label, value := range map[string]string{
		"source": identity.SourceID,
		"group":  identity.Group,
		"kind":   identity.Kind,
		"name":   identity.Name,
	} {
		if err := validateKeySegment(label, value, false); err != nil {
			return "", err
		}
	}
	for label, value := range map[string]string{
		"namespace": identity.Namespace,
		"uid":       identity.UID,
	} {
		if err := validateKeySegment(label, value, true); err != nil {
			return "", err
		}
	}
	if strings.EqualFold(identity.Kind, "Machine") {
		return "", fmt.Errorf("kind %q is ambiguous; qualify it as CapiMachine, IronCoreVM, or McmMachine", identity.Kind)
	}

	namespace := identity.Namespace
	if namespace == "" {
		namespace = missingKeySegment
	}
	uid := identity.UID
	if uid == "" {
		uid = missingKeySegment
	}
	return ResourceKey(strings.Join([]string{
		identity.SourceID,
		identity.Group,
		identity.Kind,
		namespace,
		identity.Name,
		uid,
	}, "/")), nil
}

func ParseResourceKey(key ResourceKey) (ResourceIdentity, error) {
	parts := strings.Split(string(key), "/")
	if len(parts) != 6 {
		return ResourceIdentity{}, fmt.Errorf("resource key must contain 6 segments, got %d", len(parts))
	}
	identity := ResourceIdentity{
		SourceID:  parts[0],
		Group:     parts[1],
		Kind:      parts[2],
		Namespace: parts[3],
		Name:      parts[4],
		UID:       parts[5],
	}
	if identity.Namespace == missingKeySegment {
		identity.Namespace = ""
	}
	if identity.UID == missingKeySegment {
		identity.UID = ""
	}
	canonical, err := NewResourceKey(identity)
	if err != nil {
		return ResourceIdentity{}, err
	}
	if canonical != key {
		return ResourceIdentity{}, fmt.Errorf("resource key %q is not canonical", key)
	}
	return identity, nil
}

func validateKeySegment(label, value string, optional bool) error {
	if value == "" {
		if optional {
			return nil
		}
		return fmt.Errorf("%s is required", label)
	}
	if value == missingKeySegment {
		return fmt.Errorf("%s cannot use reserved placeholder %q", label, missingKeySegment)
	}
	if value != strings.TrimSpace(value) || strings.Contains(value, "/") {
		return fmt.Errorf("%s %q is not a valid key segment", label, value)
	}
	return nil
}
