/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudstack

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CloudStackAPI is an interface for CloudStack API operations
// This allows for easier testing and mocking
type CloudStackAPI interface {
	// VirtualMachine operations
	DeployVirtualMachine(p *cloudstack.DeployVirtualMachineParams) (*cloudstack.DeployVirtualMachineResponse, error)
	GetVirtualMachineID(name string, opts ...cloudstack.OptionFunc) (string, int, error)
	ListVirtualMachines(p *cloudstack.ListVirtualMachinesParams) (*cloudstack.ListVirtualMachinesResponse, error)
	DestroyVirtualMachine(p *cloudstack.DestroyVirtualMachineParams) (*cloudstack.DestroyVirtualMachineResponse, error)

	// Service Offering operations
	ListServiceOfferings(p *cloudstack.ListServiceOfferingsParams) (*cloudstack.ListServiceOfferingsResponse, error)
	GetServiceOfferingID(name string, opts ...cloudstack.OptionFunc) (string, int, error)

	// Template operations
	ListTemplates(p *cloudstack.ListTemplatesParams) (*cloudstack.ListTemplatesResponse, error)
	GetTemplateID(name string, filter string, zoneid string, opts ...cloudstack.OptionFunc) (string, int, error)

	// Network operations
	ListNetworks(p *cloudstack.ListNetworksParams) (*cloudstack.ListNetworksResponse, error)
	GetNetworkID(name string, opts ...cloudstack.OptionFunc) (string, int, error)

	// Zone operations
	ListZones(p *cloudstack.ListZonesParams) (*cloudstack.ListZonesResponse, error)
	GetZoneID(name string, opts ...cloudstack.OptionFunc) (string, int, error)

	// Disk Offering operations
	ListDiskOfferings(p *cloudstack.ListDiskOfferingsParams) (*cloudstack.ListDiskOfferingsResponse, error)
	GetDiskOfferingID(name string, opts ...cloudstack.OptionFunc) (string, int, error)

	// Tag operations
	CreateTags(p *cloudstack.CreateTagsParams) (*cloudstack.CreateTagsResponse, error)
	DeleteTags(p *cloudstack.DeleteTagsParams) (*cloudstack.DeleteTagsResponse, error)
	ListTags(p *cloudstack.ListTagsParams) (*cloudstack.ListTagsResponse, error)
}

// Client wraps the official CloudStack Go SDK client
type Client struct {
	*cloudstack.CloudStackClient
}

// Config contains the configuration for the CloudStack client
type Config struct {
	APIURL    string
	APIKey    string
	SecretKey string
	VerifySSL bool
	Timeout   time.Duration
}

// NewClient creates a new CloudStack client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("CloudStack API URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("CloudStack API key is required")
	}
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("CloudStack secret key is required")
	}

	// Set default timeout if not specified
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}

	// Create HTTP client with custom transport
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
	}

	// Create CloudStack client with custom HTTP client
	cs := cloudstack.NewAsyncClient(
		cfg.APIURL,
		cfg.APIKey,
		cfg.SecretKey,
		!cfg.VerifySSL,
		cloudstack.WithHTTPClient(httpClient),
	)

	log.FromContext(ctx).Info("CloudStack client initialized",
		"apiURL", cfg.APIURL,
		"verifySSL", cfg.VerifySSL,
		"timeout", cfg.Timeout)

	return &Client{CloudStackClient: cs}, nil
}

// WaitForAsyncJob waits for an async job to complete and returns the result
func (c *Client) WaitForAsyncJob(ctx context.Context, jobID string, timeout time.Duration) (*cloudstack.QueryAsyncJobResultResponse, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeoutChan:
			return nil, fmt.Errorf("timeout waiting for async job %s", jobID)
		case <-ticker.C:
			p := c.Asyncjob.NewQueryAsyncJobResultParams(jobID)
			result, err := c.Asyncjob.QueryAsyncJobResult(p)
			if err != nil {
				return nil, fmt.Errorf("querying async job %s: %w", jobID, err)
			}

			switch result.Jobstatus {
			case 1: // Success
				return result, nil
			case 2: // Failed
				return nil, fmt.Errorf("async job %s failed: %s", jobID, result.Jobresult)
			case 0: // In progress
				continue
			default:
				return nil, fmt.Errorf("unknown job status %d for job %s", result.Jobstatus, jobID)
			}
		}
	}
}

// DeployVirtualMachine deploys a virtual machine
func (c *Client) DeployVirtualMachine(p *cloudstack.DeployVirtualMachineParams) (*cloudstack.DeployVirtualMachineResponse, error) {
	return c.VirtualMachine.DeployVirtualMachine(p)
}

// GetVirtualMachineID gets the VM ID by name
func (c *Client) GetVirtualMachineID(name string, opts ...cloudstack.OptionFunc) (string, int, error) {
	return c.VirtualMachine.GetVirtualMachineID(name, opts...)
}

// ListVirtualMachines lists virtual machines
func (c *Client) ListVirtualMachines(p *cloudstack.ListVirtualMachinesParams) (*cloudstack.ListVirtualMachinesResponse, error) {
	return c.VirtualMachine.ListVirtualMachines(p)
}

// DestroyVirtualMachine destroys a virtual machine
func (c *Client) DestroyVirtualMachine(p *cloudstack.DestroyVirtualMachineParams) (*cloudstack.DestroyVirtualMachineResponse, error) {
	return c.VirtualMachine.DestroyVirtualMachine(p)
}

// ListServiceOfferings lists service offerings
func (c *Client) ListServiceOfferings(p *cloudstack.ListServiceOfferingsParams) (*cloudstack.ListServiceOfferingsResponse, error) {
	return c.ServiceOffering.ListServiceOfferings(p)
}

// GetServiceOfferingID gets the service offering ID by name
func (c *Client) GetServiceOfferingID(name string, opts ...cloudstack.OptionFunc) (string, int, error) {
	return c.ServiceOffering.GetServiceOfferingID(name, opts...)
}

// ListTemplates lists templates
func (c *Client) ListTemplates(p *cloudstack.ListTemplatesParams) (*cloudstack.ListTemplatesResponse, error) {
	return c.Template.ListTemplates(p)
}

// GetTemplateID gets the template ID by name
func (c *Client) GetTemplateID(name string, filter string, zoneid string, opts ...cloudstack.OptionFunc) (string, int, error) {
	return c.Template.GetTemplateID(name, filter, zoneid, opts...)
}

// ListNetworks lists networks
func (c *Client) ListNetworks(p *cloudstack.ListNetworksParams) (*cloudstack.ListNetworksResponse, error) {
	return c.Network.ListNetworks(p)
}

// GetNetworkID gets the network ID by name
func (c *Client) GetNetworkID(name string, opts ...cloudstack.OptionFunc) (string, int, error) {
	return c.Network.GetNetworkID(name, opts...)
}

// ListZones lists zones
func (c *Client) ListZones(p *cloudstack.ListZonesParams) (*cloudstack.ListZonesResponse, error) {
	return c.Zone.ListZones(p)
}

// GetZoneID gets the zone ID by name
func (c *Client) GetZoneID(name string, opts ...cloudstack.OptionFunc) (string, int, error) {
	return c.Zone.GetZoneID(name, opts...)
}

// ListDiskOfferings lists disk offerings
func (c *Client) ListDiskOfferings(p *cloudstack.ListDiskOfferingsParams) (*cloudstack.ListDiskOfferingsResponse, error) {
	return c.DiskOffering.ListDiskOfferings(p)
}

// GetDiskOfferingID gets the disk offering ID by name
func (c *Client) GetDiskOfferingID(name string, opts ...cloudstack.OptionFunc) (string, int, error) {
	return c.DiskOffering.GetDiskOfferingID(name, opts...)
}

// CreateTags creates tags
func (c *Client) CreateTags(p *cloudstack.CreateTagsParams) (*cloudstack.CreateTagsResponse, error) {
	return c.Resourcetags.CreateTags(p)
}

// DeleteTags deletes tags
func (c *Client) DeleteTags(p *cloudstack.DeleteTagsParams) (*cloudstack.DeleteTagsResponse, error) {
	return c.Resourcetags.DeleteTags(p)
}

// ListTags lists tags
func (c *Client) ListTags(p *cloudstack.ListTagsParams) (*cloudstack.ListTagsResponse, error) {
	return c.Resourcetags.ListTags(p)
}

// Ensure Client implements CloudStackAPI
var _ CloudStackAPI = (*Client)(nil)
