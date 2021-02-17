package anka

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	c "github.com/veertuinc/packer-builder-veertu-anka/client"
	"github.com/veertuinc/packer-builder-veertu-anka/testutils"
	"gotest.tools/assert"
)

func TestRun(t *testing.T) {
	step := StepSetHyperThreading{}
	ui := packer.TestUi(t)
	ctx := context.Background()
	state := new(multistep.BasicStateBag)

	state.Put("ui", ui)
	state.Put("vm_name", "foo")

	t.Run("test disabled or nil htt values", func(t *testing.T) {
		expectedResults := make(map[string]c.MachineReadableOutput)
		expectedErrors := make(map[string]error)

		client := &testutils.TestClient{
			Results: expectedResults,
			Errors:  expectedErrors,
		}

		config := &Config{
			EnableHtt:  false,
			DisableHtt: false,
		}

		state.Put("client", client)
		state.Put("config", config)

		stepAction := step.Run(ctx, state)
		assert.Equal(t, stepAction, multistep.ActionContinue)
	})

	t.Run("conflicting htt values", func(t *testing.T) {
		expectedResults := make(map[string]c.MachineReadableOutput)
		expectedErrors := make(map[string]error)

		client := &testutils.TestClient{
			Results: expectedResults,
			Errors:  expectedErrors,
		}

		config := &Config{
			EnableHtt:  true,
			DisableHtt: true,
		}

		state.Put("client", client)
		state.Put("config", config)

		stepAction := step.Run(ctx, state)
		assert.Equal(t, stepAction, multistep.ActionHalt)
	})

	t.Run("test enable htt", func(t *testing.T) {
		expectedResults := make(map[string]c.MachineReadableOutput)
		expectedErrors := make(map[string]error)

		client := &testutils.TestClient{
			Results: expectedResults,
			Errors:  expectedErrors,
		}

		expectedResults["describe foo"] = c.MachineReadableOutput{
			Body: json.RawMessage(`{}`),
		}
		expectedResults["show foo"] = c.MachineReadableOutput{
			Body: json.RawMessage(`{}`),
		}
		expectedResults["stop --force foo"] = c.MachineReadableOutput{}
		expectedResults["modify foo set cpu --htt"] = c.MachineReadableOutput{
			Status: "OK",
		}

		expectedErrors["describe foo"] = nil
		expectedErrors["show foo"] = nil
		expectedErrors["stop --force foo"] = nil
		expectedErrors["modify foo set cpu --htt"] = nil

		config := &Config{
			EnableHtt:  true,
			DisableHtt: false,
		}

		state.Put("client", client)
		state.Put("config", config)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, "describe foo", client.Commands[0])
		assert.Equal(t, "show foo", client.Commands[1])
		assert.Equal(t, "stop --force foo", client.Commands[2])
		assert.Equal(t, "modify foo set cpu --htt", client.Commands[3])

		assert.Equal(t, stepAction, multistep.ActionContinue)
	})

	t.Run("test disable htt with 0 threads", func(t *testing.T) {
		expectedResults := make(map[string]c.MachineReadableOutput)
		expectedErrors := make(map[string]error)

		client := &testutils.TestClient{
			Results: expectedResults,
			Errors:  expectedErrors,
		}

		expectedResults["describe foo"] = c.MachineReadableOutput{
			Body: json.RawMessage(`{"CPU": {"Threads": 0}}`),
		}

		expectedErrors["describe foo"] = nil

		config := &Config{
			EnableHtt:  false,
			DisableHtt: true,
		}

		state.Put("client", client)
		state.Put("config", config)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, "describe foo", client.Commands[0])

		assert.Equal(t, stepAction, multistep.ActionContinue)
	})

	t.Run("test disable htt with > 0 threads", func(t *testing.T) {
		expectedResults := make(map[string]c.MachineReadableOutput)
		expectedErrors := make(map[string]error)

		client := &testutils.TestClient{
			Results: expectedResults,
			Errors:  expectedErrors,
		}

		expectedResults["describe foo"] = c.MachineReadableOutput{
			Body: json.RawMessage(`{"CPU": {"Threads": 1}}`),
		}
		expectedResults["show foo"] = c.MachineReadableOutput{
			Body: json.RawMessage(`{}`),
		}
		expectedResults["stop --force foo"] = c.MachineReadableOutput{}
		expectedResults["modify foo set cpu --no-htt"] = c.MachineReadableOutput{
			Status: "OK",
		}

		expectedErrors["describe foo"] = nil
		expectedErrors["show foo"] = nil
		expectedErrors["stop --force foo"] = nil
		expectedErrors["modify foo set cpu --no-htt"] = nil

		config := &Config{
			EnableHtt:  false,
			DisableHtt: true,
		}

		state.Put("client", client)
		state.Put("config", config)

		stepAction := step.Run(ctx, state)

		assert.Equal(t, "describe foo", client.Commands[0])
		assert.Equal(t, "show foo", client.Commands[1])
		assert.Equal(t, "stop --force foo", client.Commands[2])
		assert.Equal(t, "modify foo set cpu --no-htt", client.Commands[3])

		assert.Equal(t, stepAction, multistep.ActionContinue)
	})
}
