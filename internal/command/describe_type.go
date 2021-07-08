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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"knative.dev/client/pkg/printers"
	"knative.dev/pkg/apis"

	knerrors "knative.dev/client/pkg/errors"
	"knative.dev/client/pkg/kn/commands"

	"github.com/spf13/cobra"
)

var describeExample = `
  # Describe given Kamelets
  kn-source-kamelet describe-type NAME

  # Describe given Kamelets in YAML output format
  kn-source-kamelet describe-type NAME -o yaml`

// NewDescribeTypeCommand implements 'kn-source-kamelet describe-type' command
func NewDescribeTypeCommand(p *KameletPluginParams) *cobra.Command {
	printFlags := genericclioptions.NewPrintFlags("")

	cmd := &cobra.Command{
		Use:     "describe-type",
		Short:   "Show details of given Kamelet source type",
		Aliases: []string{"dt"},
		Example: describeExample,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("'kn-source-kamelet describe-type' requires the Kamelet name given as single argument")
			}
			name := args[0]

			namespace, err := p.GetNamespace(cmd)
			if err != nil {
				return err
			}

			client, err := p.NewKameletClient()
			if err != nil {
				return err
			}

			kamelet, err := client.Kamelets(namespace).Get(p.Context, name, v1.GetOptions{})
			if err != nil {
				return knerrors.GetError(err)
			}

			out := cmd.OutOrStdout()

			if !isEventSourceType(kamelet) {
				return fmt.Errorf("Kamelet %s is not an event source", name)
			}

			if printFlags.OutputFlagSpecified() {
				if strings.ToLower(*printFlags.OutputFormat) == "url" {
					fmt.Fprintf(out, "%s\n", kamelet.GetSelfLink())
					return nil
				}
				printer, err := printFlags.ToPrinter()
				if err != nil {
					return err
				}
				return printer.PrintObj(kamelet, out)
			}

			dw := printers.NewPrefixWriter(out)

			printDetails, err := cmd.Flags().GetBool("verbose")
			if err != nil {
				return err
			}

			writeKamelet(dw, kamelet, printDetails)
			dw.WriteLine()
			if err := dw.Flush(); err != nil {
				return err
			}

			// Condition info
			commands.WriteConditions(dw, asApiConditions(kamelet.Status.Conditions), printDetails)
			if err := dw.Flush(); err != nil {
				return err
			}

			return nil
		},
	}
	flags := cmd.Flags()
	commands.AddNamespaceFlags(flags, false)
	flags.BoolP("verbose", "v", false, "More output.")
	printFlags.AddFlags(cmd)
	cmd.Flag("output").Usage = fmt.Sprintf("Output format. One of: %s.", strings.Join(append(printFlags.AllowedFormats(), "url"), "|"))
	return cmd
}

func writeKamelet(dw printers.PrefixWriter, kamelet *v1alpha1.Kamelet, printDetails bool) {
	commands.WriteMetadata(dw, &kamelet.ObjectMeta, printDetails)
	if kamelet.Spec.Definition.Title != "" {
		dw.WriteAttribute("Description", fmt.Sprintf("%s - %s", kamelet.Spec.Definition.Title, kamelet.Spec.Definition.Description))
	} else {
		dw.WriteAttribute("Description", kamelet.Spec.Definition.Description)
	}

	dw.WriteAttribute("Provider", extractKameletProvider(kamelet))
	dw.WriteAttribute("Phase", string(kamelet.Status.Phase))

	dw.WriteLine()
	writeKameletProperties(dw, kamelet, printDetails)
}

func writeKameletProperties(dw printers.PrefixWriter, kamelet *v1alpha1.Kamelet, printDetails bool) {
	label := "Properties"
	if printDetails {
		section := dw.WriteAttribute("Properties", "")
		maxLen := getMaxPropertyNameLen(kamelet.Spec.Definition.Properties)
		format := "%-" + maxLen + "s %-8s %-8s %s\n"
		section.Writef(format, "Name", "Required", "Type", "Description")
		for name, property := range kamelet.Spec.Definition.Properties {
			section.Writef(format, name, isRequired(name, kamelet.Spec.Definition.Required), property.Type, property.Description)
		}
	} else {
		result := ""
		maxWidth := commands.TruncateAt - len(label) - 2
		for name := range kamelet.Spec.Definition.Properties {
			result += fmt.Sprintf("%s, ", name)
			if len(result) > maxWidth {
				break
			}
		}
		// cut of two latest chars
		result = strings.TrimRight(result, ", ")
		if len(result) > maxWidth {
			result = result[:maxWidth-4] + " ..."
		}

		dw.WriteColsLn(printers.Label(label), result)
	}
}

func isRequired(name string, required []string) string {
	for _, propertyName := range required {
		if propertyName == name {
			return "true"
		}
	}

	return "false"
}

func getMaxPropertyNameLen(properties map[string]v1alpha1.JSONSchemaProps) string {
	max := 0
	for name := range properties {
		if len(name) > max {
			max = len(name)
		}
	}
	return strconv.Itoa(max)
}

func extractKameletProvider(kamelet *v1alpha1.Kamelet) string {
	return kamelet.Labels["camel.apache.org/kamelet.provider"]
}

func isEventSourceType(kamelet *v1alpha1.Kamelet) bool {
	return kamelet.Labels["camel.apache.org/kamelet.type"] == "source"
}

func asApiConditions(conditions []v1alpha1.KameletCondition) apis.Conditions {
	var aConditions apis.Conditions

	for _, condition := range conditions {
		aConditions = append(aConditions, apis.Condition{
			Type:   apis.ConditionReady,
			Status: condition.Status,
			LastTransitionTime: apis.VolatileTime{
				Inner: condition.LastTransitionTime,
			},
			Reason:  condition.Reason,
			Message: condition.Message,
		})
	}

	return aConditions
}
