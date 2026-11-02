package mapper

import (
	"testing"
)

func TestCuratedMappings(t *testing.T) {
	m := NewMapper()

	// build-essential
	res := m.Resolve("build-essential", "apt", "dnf")
	if res.TargetName != "@development-tools" || res.Confidence != ConfidenceExact {
		t.Errorf("expected @development-tools exact, got: %+v", res)
	}

	resArch := m.Resolve("build-essential", "apt", "pacman")
	if resArch.TargetName != "base-devel" || resArch.Confidence != ConfidenceExact {
		t.Errorf("expected base-devel exact, got: %+v", resArch)
	}

	// libssl-dev
	resSSL := m.Resolve("libssl-dev", "apt", "dnf")
	if resSSL.TargetName != "openssl-devel" || resSSL.Confidence != ConfidenceExact {
		t.Errorf("expected openssl-devel exact, got: %+v", resSSL)
	}
}

func TestTransformationRules(t *testing.T) {
	m := NewMapper()

	// -dev to -devel
	res := m.Resolve("libyaml-dev", "apt", "dnf")
	if res.TargetName != "libyaml-devel" || res.Confidence != ConfidenceTransformed {
		t.Errorf("expected libyaml-devel transformed, got: %+v", res)
	}

	// python3- to python- on pacman
	resPy := m.Resolve("python3-cryptography", "apt", "pacman")
	if resPy.TargetName != "python-cryptography" || resPy.Confidence != ConfidenceTransformed {
		t.Errorf("expected python-cryptography transformed, got: %+v", resPy)
	}
}

func TestSameNameFallback(t *testing.T) {
	m := NewMapper()

	res := m.Resolve("custom-cli-tool", "apt", "pacman")
	if res.TargetName != "custom-cli-tool" || res.Confidence != ConfidenceRepoSame {
		t.Errorf("expected custom-cli-tool repo-same, got: %+v", res)
	}
	if !res.AutoSelect {
		t.Errorf("expected AutoSelect true for repo-same")
	}
}
