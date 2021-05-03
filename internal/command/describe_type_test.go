/*
 * Copyright © 2021 The Knative Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	camelkv1alpha1 "github.com/apache/camel-k/pkg/client/camel/clientset/versioned/typed/camel/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/util"
	"knative.dev/kn-plugin-source-kamelet/internal/client"

	"gotest.tools/v3/assert"
)

func TestDescribeTypeSetup(t *testing.T) {
	p := KameletPluginParams{
		Context: context.TODO(),
	}

	describeCmd := NewDescribeTypeCommand(&p)
	assert.Equal(t, describeCmd.Use, "describe-type")
	assert.Equal(t, describeCmd.Short, "Show details of given Kamelet source type")
	assert.Assert(t, describeCmd.RunE != nil)
}
func TestDescribeTypeErrorCase(t *testing.T) {
	mockClient := client.NewMockKameletClient(t)
	recorder := mockClient.Recorder()

	_, err := runDescribeTypeCmd(mockClient)
	assert.Error(t, err, "'kn-source-kamelet describe-type' requires the Kamelet name given as single argument")
	recorder.Validate()
}

func TestDescribeTypeErrorCaseNotFound(t *testing.T) {
	mockClient := client.NewMockKameletClient(t)
	recorder := mockClient.Recorder()

	kamelet := createKamelet("k1")
	recorder.Get(kamelet, errors.New("not found"))

	_, err := runDescribeTypeCmd(mockClient, "k1")
	assert.Error(t, err, "not found")
	recorder.Validate()
}

func TestDescribeTypeErrorCaseNoEventSource(t *testing.T) {
	mockClient := client.NewMockKameletClient(t)
	recorder := mockClient.Recorder()

	kamelet := createKamelet("k1")
	kamelet.Labels = map[string]string{
		"camel.apache.org/kamelet.type": "sink",
	}
	recorder.Get(kamelet, nil)

	_, err := runDescribeTypeCmd(mockClient, "k1")
	assert.Error(t, err, "Kamelet k1 is not an event source")
	recorder.Validate()
}

func TestDescribeTypeOutput(t *testing.T) {
	mockClient := client.NewMockKameletClient(t)
	recorder := mockClient.Recorder()

	kamelet := createKamelet("k1")
	recorder.Get(kamelet, nil)

	output, err := runDescribeTypeCmd(mockClient, "k1")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")

	assert.Check(t, util.ContainsAll(outputLines[0], "Name:", "k1"))
	assert.Check(t, util.ContainsAll(outputLines[1], "Namespace:", "default"))
	assert.Check(t, util.ContainsAll(outputLines[2], "Labels:", "camel.apache.org/kamelet.type=source", "camel.apache.org/kamelet.provider=Community"))
	assert.Check(t, util.ContainsAll(outputLines[3], "Age:", "0s"))
	assert.Check(t, util.ContainsAll(outputLines[4], "Description:", "Kamelet k1 - Sample Kamelet source"))
	assert.Check(t, util.ContainsAll(outputLines[5], "Provider:", "Community"))
	assert.Check(t, util.ContainsAll(outputLines[6], "Phase:", "Ready"))

	assert.Check(t, util.ContainsAll(outputLines[8], "Properties:", "k1_prop, k1_optional"))

	assert.Check(t, util.ContainsAll(outputLines[10], "Conditions:"))
	assert.Check(t, util.ContainsAll(outputLines[11], "OK", "TYPE", "AGE", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[12], "++", "Ready", "", ""))

	recorder.Validate()
}

func TestDescribeTypeVerboseOutput(t *testing.T) {
	mockClient := client.NewMockKameletClient(t)
	recorder := mockClient.Recorder()

	kamelet := createKamelet("k1")
	recorder.Get(kamelet, nil)

	output, err := runDescribeTypeCmd(mockClient, "k1", "--verbose")
	assert.NilError(t, err)

	outputLines := strings.Split(output, "\n")

	assert.Check(t, util.ContainsAll(outputLines[0], "Name:", "k1"))
	assert.Check(t, util.ContainsAll(outputLines[1], "Namespace:", "default"))
	assert.Check(t, util.ContainsAll(outputLines[2], "Labels:", "camel.apache.org/kamelet.provider=Community"))
	assert.Check(t, util.ContainsAll(outputLines[3], "camel.apache.org/kamelet.type=source"))
	assert.Check(t, util.ContainsAll(outputLines[4], "Age:", "0s"))
	assert.Check(t, util.ContainsAll(outputLines[5], "Description:", "Kamelet k1 - Sample Kamelet source"))
	assert.Check(t, util.ContainsAll(outputLines[6], "Provider:", "Community"))
	assert.Check(t, util.ContainsAll(outputLines[7], "Phase:", "Ready"))

	assert.Check(t, util.ContainsAll(outputLines[9], "Properties:"))
	assert.Check(t, util.ContainsAll(outputLines[10], "Name", "Required", "Type", "Description"))
	assert.Check(t, util.ContainsAll(outputLines[11], "k1_prop", "true", "string", "The k1 required property"))
	assert.Check(t, util.ContainsAll(outputLines[12], "k1_optional", "false", "boolean", "The k1 optional property"))

	assert.Check(t, util.ContainsAll(outputLines[14], "Conditions:"))
	assert.Check(t, util.ContainsAll(outputLines[15], "OK", "TYPE", "AGE", "REASON"))
	assert.Check(t, util.ContainsAll(outputLines[16], "++", "Ready", "", ""))

	recorder.Validate()
}

func TestDescribeTypeURL(t *testing.T) {
	mockClient := client.NewMockKameletClient(t)
	recorder := mockClient.Recorder()

	kamelet := createKamelet("k1")
	recorder.Get(kamelet, nil)

	output, err := runDescribeTypeCmd(mockClient, "k1", "-o", "url")
	assert.NilError(t, err, "Kamelet should be described with url as output")

	outputLines := strings.Split(output, "\n")

	assert.Check(t, util.ContainsAll(outputLines[0], "kamelets/k1"))
	recorder.Validate()
}

func runDescribeTypeCmd(c *client.MockKameletClient, options ...string) (string, error) {
	p := KameletPluginParams{
		KnParams: &commands.KnParams{},
		Context:  context.TODO(),
		NewKameletClient: func() (camelkv1alpha1.CamelV1alpha1Interface, error) {
			return c, nil
		},
	}

	describeCmd, _, output := commands.CreateSourcesTestKnCommand(NewDescribeTypeCommand(&p), p.KnParams)

	args := []string{"describe-type"}
	args = append(args, options...)
	describeCmd.SetArgs(args)
	err := describeCmd.Execute()

	return output.String(), err
}
