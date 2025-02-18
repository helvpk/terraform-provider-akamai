package property

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/papi"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/logger"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func diffSuppressRules(_, oldRules, newRules string, _ *schema.ResourceData) bool {
	rulesEqual, err := rulesJSONEqual(oldRules, newRules)
	if err != nil {
		logger.Get("PAPI", "diffSuppressRules").Error(err.Error())
	}

	return rulesEqual
}

// rulesJSONEqual handles comparison between two papi.RulesUpdate JSON representations.
func rulesJSONEqual(old, new string) (bool, error) {
	if old == "" || new == "" {
		return old == new, nil
	}

	if old == new {
		return true, nil
	}

	var oldRules papi.RulesUpdate
	if err := json.Unmarshal([]byte(old), &oldRules); err != nil {
		return false, fmt.Errorf("'old' = %s, unmarshal: %w", old, err)
	}

	var newRules papi.RulesUpdate
	if err := json.Unmarshal([]byte(new), &newRules); err != nil {
		return false, fmt.Errorf("'new' = %s, unmarshal: %w", new, err)
	}

	return ruleTreesEqual(&oldRules, &newRules), nil
}

func ruleTreesEqual(old, new *papi.RulesUpdate) bool {
	if old.Comments != new.Comments {
		return false
	}

	return rulesEqual(&old.Rules, &new.Rules)
}

// rulesEqual handles comparison between two papi.Rules objects ignoring the order in
// collection of variables.
func rulesEqual(oldRules, newRules *papi.Rules) bool {
	if len(oldRules.Behaviors) != len(newRules.Behaviors) ||
		len(oldRules.Criteria) != len(newRules.Criteria) ||
		len(oldRules.Variables) != len(newRules.Variables) ||
		len(oldRules.Children) != len(newRules.Children) {
		return false
	}

	if len(oldRules.Children) > 0 {
		for i := range oldRules.Children {
			if !rulesEqual(&oldRules.Children[i], &newRules.Children[i]) {
				return false
			}
		}
	} else {
		oldRules.Children = nil
		newRules.Children = nil
	}

	if len(oldRules.Behaviors) == 0 {
		oldRules.Behaviors = nil
	}
	if len(newRules.Behaviors) == 0 {
		newRules.Behaviors = nil
	}

	if len(oldRules.Criteria) == 0 {
		oldRules.Criteria = nil
	}
	if len(newRules.Criteria) == 0 {
		newRules.Criteria = nil
	}

	oldRules.Variables = orderVariables(oldRules.Variables)
	newRules.Variables = orderVariables(newRules.Variables)

	removeNilOptions(oldRules)
	removeNilOptions(newRules)

	return reflect.DeepEqual(oldRules, newRules)
}

// PAPI sometimes adds fields (with value null) that are not present in configuration (e.g. exported in cli-terraform)
// these fields has to be diff suppressed and treated as no diff in customizeDiff to not provide diff after plan with no actual changes
func removeNilOptions(rules *papi.Rules) {
	for _, b := range rules.Behaviors {
		removeNils(b.Options)
	}
}

func removeNils(parent map[string]any) {
	for k, v := range parent {
		if v == nil {
			delete(parent, k)
		} else if vv, ok := v.(map[string]any); ok {
			removeNils(vv)
		}
	}
}

func orderVariables(variables []papi.RuleVariable) []papi.RuleVariable {
	if len(variables) == 0 {
		return nil
	}
	sort.Slice(variables, func(i, j int) bool {
		return variables[i].Name < variables[j].Name
	})
	return variables
}
