package clialert

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/crowdsecurity/crowdsec/cmd/crowdsec-cli/cstable"
	"github.com/crowdsecurity/crowdsec/pkg/models"
)

func alertsTable(out io.Writer, wantColor string, alerts *models.GetAlertsResponse, printMachine bool) {
	t := cstable.New(out, wantColor)
	t.SetRowLines(false)

	header := []string{"ID", "value", "reason", "country", "as", "decisions", "created_at"}
	if printMachine {
		header = append(header, "machine")
	}

	t.SetHeaders(header...)

	for _, alertItem := range *alerts {
		displayVal := *alertItem.Source.Scope
		if len(alertItem.Decisions) > 1 {
			displayVal = fmt.Sprintf("%s (%d %ss)", *alertItem.Source.Scope, len(alertItem.Decisions), *alertItem.Decisions[0].Scope)
		} else if *alertItem.Source.Value != "" {
			displayVal += ":" + *alertItem.Source.Value
		}

		row := []string{
			strconv.Itoa(int(alertItem.ID)),
			displayVal,
			*alertItem.Scenario,
			alertItem.Source.Cn,
			alertItem.Source.GetAsNumberName(),
			decisionsFromAlert(alertItem),
			*alertItem.StartAt,
		}

		if printMachine {
			row = append(row, alertItem.MachineID)
		}

		t.AddRow(row...)
	}

	t.Render()
}

func alertDecisionsTable(out io.Writer, wantColor string, alert *models.Alert) {
	foundActive := false
	t := cstable.New(out, wantColor)
	t.SetRowLines(false)
	t.SetHeaders("ID", "scope:value", "action", "expiration", "created_at")

	for _, decision := range alert.Decisions {
		parsedDuration, err := time.ParseDuration(*decision.Duration)
		if err != nil {
			log.Error(err)
		}

		expire := time.Now().UTC().Add(parsedDuration)
		if time.Now().UTC().After(expire) {
			continue
		}

		foundActive = true
		scopeAndValue := *decision.Scope

		if *decision.Value != "" {
			scopeAndValue += ":" + *decision.Value
		}

		t.AddRow(
			strconv.Itoa(int(decision.ID)),
			scopeAndValue,
			*decision.Type,
			*decision.Duration,
			alert.CreatedAt,
		)
	}

	if foundActive {
		t.Writer.SetTitle("Active Decisions")
		t.Render() // Send output
	}
}

func alertEventTable(out io.Writer, wantColor string, event *models.Event) {
	fmt.Fprintf(out, "\n- Date: %s\n", *event.Timestamp)

	t := cstable.New(out, wantColor)
	t.SetHeaders("Key", "Value")
	sort.Slice(event.Meta, func(i, j int) bool {
		return event.Meta[i].Key < event.Meta[j].Key
	})

	for _, meta := range event.Meta {
		t.AddRow(
			meta.Key,
			meta.Value,
		)
	}

	t.Render() // Send output
}
