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

package options

import (
	"context"
	"errors"
	"fmt"
	"os"
)

type Options struct {
	CloudStackAPIURL    string
	CloudStackAPIKey    string
	CloudStackSecretKey string
	CloudStackVerifySSL bool
	ClusterName         string
}

func (o *Options) AddFlags(fs interface{}) {
	// Flags would be added here if using flag package
	// For now, we read from environment variables
}

func (o *Options) Parse(ctx context.Context, _ ...interface{}) error {
	var errs error

	o.CloudStackAPIURL = os.Getenv("CLOUDSTACK_API_URL")
	if o.CloudStackAPIURL == "" {
		errs = errors.Join(errs, fmt.Errorf("CLOUDSTACK_API_URL is required"))
	}

	o.CloudStackAPIKey = os.Getenv("CLOUDSTACK_API_KEY")
	if o.CloudStackAPIKey == "" {
		errs = errors.Join(errs, fmt.Errorf("CLOUDSTACK_API_KEY is required"))
	}

	o.CloudStackSecretKey = os.Getenv("CLOUDSTACK_SECRET_KEY")
	if o.CloudStackSecretKey == "" {
		errs = errors.Join(errs, fmt.Errorf("CLOUDSTACK_SECRET_KEY is required"))
	}

	o.CloudStackVerifySSL = os.Getenv("CLOUDSTACK_VERIFY_SSL") != "false"

	o.ClusterName = os.Getenv("CLUSTER_NAME")
	if o.ClusterName == "" {
		errs = errors.Join(errs, fmt.Errorf("CLUSTER_NAME is required"))
	}

	return errs
}

type optionsKey struct{}

func ToContext(ctx context.Context, opts *Options) context.Context {
	return context.WithValue(ctx, optionsKey{}, opts)
}

func FromContext(ctx context.Context) *Options {
	data := ctx.Value(optionsKey{})
	if data == nil {
		// Return default options if not found
		return &Options{
			CloudStackVerifySSL: true,
		}
	}
	return data.(*Options)
}
