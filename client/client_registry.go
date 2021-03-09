package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/packer-plugin-sdk/net"
)

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#registry-list
type RegistryListResponse struct {
	Latest string `json:"latest"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

// Run command against the registry
type RegistryParams struct {
	RegistryName string
	RegistryURL  string
	NodeCertPath string
	NodeKeyPath  string
	CaRootPath   string
	IsInsecure   bool
}

// https://ankadocs.veertu.com/docs/anka-virtualization/command-reference/#registry-push
type RegistryPushParams struct {
	VMID        string
	Tag         string
	Description string
	RemoteVM    string
	Local       bool
}

func (c *AnkaClient) RegistryList(registryParams RegistryParams) ([]RegistryListResponse, error) {
	output, err := runRegistryCommand(registryParams, "list")
	if err != nil {
		return nil, err
	}
	if output.Status != "OK" {
		log.Print("Error executing registry list command: ", output.ExceptionType, " ", output.Message)
		return nil, fmt.Errorf(output.Message)
	}

	var response []RegistryListResponse
	err = json.Unmarshal(output.Body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (c *AnkaClient) RegistryPush(registryParams RegistryParams, pushParams RegistryPushParams) error {
	cmdArgs := []string{"push"}
	if pushParams.Tag != "" {
		cmdArgs = append(cmdArgs, "--tag", pushParams.Tag)
	}
	if pushParams.Description != "" {
		cmdArgs = append(cmdArgs, "--description", pushParams.Description)
	}
	if pushParams.RemoteVM != "" {
		cmdArgs = append(cmdArgs, "--remote-vm", pushParams.RemoteVM)
	}
	if pushParams.Local {
		cmdArgs = append(cmdArgs, "--local")
	}
	cmdArgs = append(cmdArgs, pushParams.VMID)

	output, err := runRegistryCommand(registryParams, cmdArgs...)
	if err != nil {
		return err
	}
	if output.Status != "OK" {
		log.Print("Error executing registry push command: ", output.ExceptionType, " ", output.Message)
		return fmt.Errorf(output.Message)
	}
	return nil
}

// https://ankadocs.veertu.com/docs/anka-build-cloud/working-with-registry-and-api/#revert
func (c *AnkaClient) RegistryRevert(url string, id string) error {
	response, err := registryRESTRequest("DELETE", fmt.Sprintf("%s/registry/revert?id=%s", url, id), nil)
	if err != nil {
		return err
	}
	if response.Status != statusOK {
		return fmt.Errorf("failed to revert VM on registry: %s", response.Message)
	}

	return nil
}

func registryRESTRequest(method string, url string, body io.Reader) (MachineReadableOutput, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return MachineReadableOutput{}, err
	}

	log.Printf("[API REQUEST] [%s] %s", method, url)

	httpClient := net.HttpClientWithEnvironmentProxy()
	resp, err := httpClient.Do(request)
	if err != nil {
		return MachineReadableOutput{}, err
	}

	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return MachineReadableOutput{}, err
		}

		log.Printf("[API RESPONSE] %s", string(body))

		return parseOutput(body)
	}

	return MachineReadableOutput{}, fmt.Errorf("unsupported http response code: %d", resp.StatusCode)
}
