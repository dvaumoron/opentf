// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"os"
	"strings"
	"testing"

	"github.com/zclconf/go-cty/cty"

	"github.com/opentofu/opentofu/internal/addrs"
	"github.com/opentofu/opentofu/internal/plans"
	"github.com/opentofu/opentofu/internal/states"
)

func TestGraph(t *testing.T) {
	td := t.TempDir()
	testCopyDir(t, testFixturePath("graph"), td)
	defer testChdir(t, td)()

	view, done := testView(t)
	c := &GraphCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(applyFixtureProvider()),
			View:             view,
		},
	}

	args := []string{}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, output.Stderr())
	}

	stdout := output.Stdout()
	if !strings.Contains(stdout, `provider[\"registry.opentofu.org/hashicorp/test\"]`) {
		t.Fatalf("doesn't look like digraph: %s", stdout)
	}
}

func TestGraph_multipleArgs(t *testing.T) {
	view, done := testView(t)
	c := &GraphCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(applyFixtureProvider()),
			View:             view,
		},
	}

	args := []string{
		"bad",
		"bad",
	}
	code := c.Run(args)
	output := done(t)
	if code != 1 {
		t.Fatalf("bad status code: %d\n\n%s", code, output.Stderr())
	}
}

func TestGraph_noArgs(t *testing.T) {
	td := t.TempDir()
	testCopyDir(t, testFixturePath("graph"), td)
	defer testChdir(t, td)()

	view, done := testView(t)
	c := &GraphCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(applyFixtureProvider()),
			View:             view,
		},
	}

	args := []string{}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, output.Stderr())
	}

	stdout := output.Stdout()
	if !strings.Contains(stdout, `provider[\"registry.opentofu.org/hashicorp/test\"]`) {
		t.Fatalf("doesn't look like digraph: %s", stdout)
	}
}

func TestGraph_noConfig(t *testing.T) {
	td := t.TempDir()
	os.MkdirAll(td, 0755)
	defer testChdir(t, td)()

	view, done := testView(t)
	c := &GraphCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(applyFixtureProvider()),
			View:             view,
		},
	}

	// Running the graph command without a config should not panic,
	// but this may be an error at some point in the future.
	args := []string{"-type", "apply"}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, output.Stderr())
	}
}

func TestGraph_plan(t *testing.T) {
	testCwd(t)

	plan := &plans.Plan{
		Changes: plans.NewChanges(),
	}
	plan.Changes.Resources = append(plan.Changes.Resources, &plans.ResourceInstanceChangeSrc{
		Addr: addrs.Resource{
			Mode: addrs.ManagedResourceMode,
			Type: "test_instance",
			Name: "bar",
		}.Instance(addrs.NoKey).Absolute(addrs.RootModuleInstance),
		ChangeSrc: plans.ChangeSrc{
			Action: plans.Delete,
			Before: plans.DynamicValue(`{}`),
			After:  plans.DynamicValue(`null`),
		},
		ProviderAddr: addrs.AbsProviderConfig{
			Provider: addrs.NewDefaultProvider("test"),
			Module:   addrs.RootModule,
		},
	})
	emptyConfig, err := plans.NewDynamicValue(cty.EmptyObjectVal, cty.EmptyObject)
	if err != nil {
		t.Fatal(err)
	}
	plan.Backend = plans.Backend{
		// Doesn't actually matter since we aren't going to activate the backend
		// for this command anyway, but we need something here for the plan
		// file writer to succeed.
		Type:   "placeholder",
		Config: emptyConfig,
	}
	_, configSnap := testModuleWithSnapshot(t, "graph")

	planPath := testPlanFile(t, configSnap, states.NewState(), plan)

	view, done := testView(t)
	c := &GraphCommand{
		Meta: Meta{
			testingOverrides: metaOverridesForProvider(applyFixtureProvider()),
			View:             view,
		},
	}

	args := []string{
		"-plan", planPath,
	}
	code := c.Run(args)
	output := done(t)
	if code != 0 {
		t.Fatalf("bad: %d\n\n%s", code, output.Stderr())
	}

	stdout := output.Stdout()
	if !strings.Contains(stdout, `provider[\"registry.opentofu.org/hashicorp/test\"]`) {
		t.Fatalf("doesn't look like digraph: %s", stdout)
	}
}
