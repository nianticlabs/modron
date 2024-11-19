package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/nianticlabs/modron/src/collector"
	"github.com/nianticlabs/modron/src/storage"
)

func validateArgs() error {
	var errArr []error
	errArr = append(errArr, validateStorage()...)
	errArr = append(errArr, validateProductionArgs()...)
	errArr = append(errArr, validateGCPArgs()...)
	errArr = append(errArr, validateRiskArgs())
	if err := errors.Join(errArr...); err != nil {
		return err
	}
	warnArgs()
	return nil
}

func validateRiskArgs() error {
	var impactMap map[string]string
	if err := json.Unmarshal([]byte(args.ImpactMap), &impactMap); err != nil {
		return fmt.Errorf("unable to decode impact map: %w", err)
	}

	return nil
}

func validateStorage() (errors []error) {
	switch storage.Type(strings.ToLower(args.Storage)) {
	case storage.Memory:
		break
	case storage.SQL:
		errors = append(validateSQL(), errors...)
	default:
		errors = append(errors, fmt.Errorf("invalid storage type: %s", args.Storage))
	}
	return
}

func validateSQL() (errors []error) {
	switch strings.ToLower(args.SQLBackendDriver) {
	case "postgres":
		break
	default:
		errors = append(errors, fmt.Errorf("invalid SQL backend driver: %s", args.SQLBackendDriver))
		return
	}

	if args.SQLConnectionString == "" {
		errors = append(errors, fmt.Errorf("SQL connection string is required"))
	}

	if args.DbBatchSize < 1 {
		errors = append(errors, fmt.Errorf("DB batch size must be greater than 0"))
	}

	if args.DbMaxConnections < 1 {
		errors = append(errors, fmt.Errorf("DB max connections must be greater than 0"))
	}

	if args.DbMaxIdleConnections < 1 {
		errors = append(errors, fmt.Errorf("DB max idle connections must be greater than 0"))
	}
	return
}

func validateProductionArgs() (errors []error) {
	if strings.EqualFold(args.Environment, "production") {
		if args.NotificationService == "" {
			errors = append(errors, fmt.Errorf("notification service is required in production"))
		}

		if args.Collector == collector.Fake {
			errors = append(errors, fmt.Errorf("fake collector cannot be used in production"))
		}

		if args.SkipIAP {
			errors = append(errors, fmt.Errorf("IAP cannot be skipped in production"))
		}

		if args.PersistentCache {
			errors = append(errors, fmt.Errorf("persistent cache cannot be used in production"))
		}
	}
	return
}

func validateGCPArgs() (errors []error) {
	if args.OrgID == "" {
		errors = append(errors, fmt.Errorf("organization ID is required"))
	}

	if args.OrgSuffix == "" {
		errors = append(errors, fmt.Errorf("organization suffix is required"))
	}
	return
}

func warnArgs() {
	if args.Collector == collector.Fake {
		log.Warnf("Using fake collector")
	}
	if args.SkipIAP {
		log.Warnf("Skipping IAP, if you see this in production, reach out to security.")
	}
	if args.NotificationService == "" {
		log.Warnf("Notification service address is empty, logging instead")
	}
	if len(args.AdditionalAdminRoles) == 0 {
		log.Warnf("No additional admin roles specified")
	}
}
