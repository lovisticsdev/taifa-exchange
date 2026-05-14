package policy

import (
	"context"
	"errors"
	"fmt"
)

type Service struct {
	repository *Repository
	evaluator  Evaluator
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
		evaluator:  NewEvaluator(),
	}
}

func (s *Service) Evaluate(ctx context.Context, input AuthorizationInput) (EvaluationResult, error) {
	if s == nil || s.repository == nil {
		return EvaluationResult{}, fmt.Errorf("policy service is not configured")
	}

	lookupInput := input.LookupInput()

	foundPolicy, err := s.repository.FindActiveByRoute(ctx, lookupInput)
	if err != nil {
		if errors.Is(err, ErrPolicyNotFound) {
			return s.evaluator.Evaluate(EvaluationInput{
				TargetSystem:             lookupInput.TargetSystem,
				Method:                   lookupInput.Method,
				Route:                    lookupInput.Route,
				Operation:                lookupInput.Operation,
				Policy:                   nil,
				ActorRoles:               input.ActorRoles,
				OrganizationCapabilities: input.OrganizationCapabilities,
			}), nil
		}

		return EvaluationResult{}, err
	}

	return s.evaluator.Evaluate(EvaluationInput{
		TargetSystem:             lookupInput.TargetSystem,
		Method:                   lookupInput.Method,
		Route:                    lookupInput.Route,
		Operation:                lookupInput.Operation,
		Policy:                   foundPolicy,
		ActorRoles:               input.ActorRoles,
		OrganizationCapabilities: input.OrganizationCapabilities,
	}), nil
}
