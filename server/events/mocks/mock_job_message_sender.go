// Code generated by pegomock. DO NOT EDIT.
// Source: github.com/runatlantis/atlantis/server/events (interfaces: JobMessageSender)

package mocks

import (
	pegomock "github.com/petergtz/pegomock/v4"
	command "github.com/runatlantis/atlantis/server/events/command"
	"reflect"
	"time"
)

type MockJobMessageSender struct {
	fail func(message string, callerSkip ...int)
}

func NewMockJobMessageSender(options ...pegomock.Option) *MockJobMessageSender {
	mock := &MockJobMessageSender{}
	for _, option := range options {
		option.Apply(mock)
	}
	return mock
}

func (mock *MockJobMessageSender) SetFailHandler(fh pegomock.FailHandler) { mock.fail = fh }
func (mock *MockJobMessageSender) FailHandler() pegomock.FailHandler      { return mock.fail }

func (mock *MockJobMessageSender) Send(ctx command.ProjectContext, msg string, operationComplete bool) {
	if mock == nil {
		panic("mock must not be nil. Use myMock := NewMockJobMessageSender().")
	}
	_params := []pegomock.Param{ctx, msg, operationComplete}
	pegomock.GetGenericMockFrom(mock).Invoke("Send", _params, []reflect.Type{})
}

func (mock *MockJobMessageSender) VerifyWasCalledOnce() *VerifierMockJobMessageSender {
	return &VerifierMockJobMessageSender{
		mock:                   mock,
		invocationCountMatcher: pegomock.Times(1),
	}
}

func (mock *MockJobMessageSender) VerifyWasCalled(invocationCountMatcher pegomock.InvocationCountMatcher) *VerifierMockJobMessageSender {
	return &VerifierMockJobMessageSender{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
	}
}

func (mock *MockJobMessageSender) VerifyWasCalledInOrder(invocationCountMatcher pegomock.InvocationCountMatcher, inOrderContext *pegomock.InOrderContext) *VerifierMockJobMessageSender {
	return &VerifierMockJobMessageSender{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		inOrderContext:         inOrderContext,
	}
}

func (mock *MockJobMessageSender) VerifyWasCalledEventually(invocationCountMatcher pegomock.InvocationCountMatcher, timeout time.Duration) *VerifierMockJobMessageSender {
	return &VerifierMockJobMessageSender{
		mock:                   mock,
		invocationCountMatcher: invocationCountMatcher,
		timeout:                timeout,
	}
}

type VerifierMockJobMessageSender struct {
	mock                   *MockJobMessageSender
	invocationCountMatcher pegomock.InvocationCountMatcher
	inOrderContext         *pegomock.InOrderContext
	timeout                time.Duration
}

func (verifier *VerifierMockJobMessageSender) Send(ctx command.ProjectContext, msg string, operationComplete bool) *MockJobMessageSender_Send_OngoingVerification {
	_params := []pegomock.Param{ctx, msg, operationComplete}
	methodInvocations := pegomock.GetGenericMockFrom(verifier.mock).Verify(verifier.inOrderContext, verifier.invocationCountMatcher, "Send", _params, verifier.timeout)
	return &MockJobMessageSender_Send_OngoingVerification{mock: verifier.mock, methodInvocations: methodInvocations}
}

type MockJobMessageSender_Send_OngoingVerification struct {
	mock              *MockJobMessageSender
	methodInvocations []pegomock.MethodInvocation
}

func (c *MockJobMessageSender_Send_OngoingVerification) GetCapturedArguments() (command.ProjectContext, string, bool) {
	ctx, msg, operationComplete := c.GetAllCapturedArguments()
	return ctx[len(ctx)-1], msg[len(msg)-1], operationComplete[len(operationComplete)-1]
}

func (c *MockJobMessageSender_Send_OngoingVerification) GetAllCapturedArguments() (_param0 []command.ProjectContext, _param1 []string, _param2 []bool) {
	_params := pegomock.GetGenericMockFrom(c.mock).GetInvocationParams(c.methodInvocations)
	if len(_params) > 0 {
		if len(_params) > 0 {
			_param0 = make([]command.ProjectContext, len(c.methodInvocations))
			for u, param := range _params[0] {
				_param0[u] = param.(command.ProjectContext)
			}
		}
		if len(_params) > 1 {
			_param1 = make([]string, len(c.methodInvocations))
			for u, param := range _params[1] {
				_param1[u] = param.(string)
			}
		}
		if len(_params) > 2 {
			_param2 = make([]bool, len(c.methodInvocations))
			for u, param := range _params[2] {
				_param2[u] = param.(bool)
			}
		}
	}
	return
}
