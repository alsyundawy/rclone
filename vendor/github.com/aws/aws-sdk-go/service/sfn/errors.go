// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package sfn

const (

	// ErrCodeActivityDoesNotExist for service response error code
	// "ActivityDoesNotExist".
	//
	// The specified activity does not exist.
	ErrCodeActivityDoesNotExist = "ActivityDoesNotExist"

	// ErrCodeActivityLimitExceeded for service response error code
	// "ActivityLimitExceeded".
	//
	// The maximum number of activities has been reached. Existing activities must
	// be deleted before a new activity can be created.
	ErrCodeActivityLimitExceeded = "ActivityLimitExceeded"

	// ErrCodeActivityWorkerLimitExceeded for service response error code
	// "ActivityWorkerLimitExceeded".
	//
	// The maximum number of workers concurrently polling for activity tasks has
	// been reached.
	ErrCodeActivityWorkerLimitExceeded = "ActivityWorkerLimitExceeded"

	// ErrCodeExecutionAlreadyExists for service response error code
	// "ExecutionAlreadyExists".
	//
	// The execution has the same name as another execution (but a different input).
	//
	// Executions with the same name and input are considered idempotent.
	ErrCodeExecutionAlreadyExists = "ExecutionAlreadyExists"

	// ErrCodeExecutionDoesNotExist for service response error code
	// "ExecutionDoesNotExist".
	//
	// The specified execution does not exist.
	ErrCodeExecutionDoesNotExist = "ExecutionDoesNotExist"

	// ErrCodeExecutionLimitExceeded for service response error code
	// "ExecutionLimitExceeded".
	//
	// The maximum number of running executions has been reached. Running executions
	// must end or be stopped before a new execution can be started.
	ErrCodeExecutionLimitExceeded = "ExecutionLimitExceeded"

	// ErrCodeInvalidArn for service response error code
	// "InvalidArn".
	//
	// The provided Amazon Resource Name (ARN) is invalid.
	ErrCodeInvalidArn = "InvalidArn"

	// ErrCodeInvalidDefinition for service response error code
	// "InvalidDefinition".
	//
	// The provided Amazon States Language definition is invalid.
	ErrCodeInvalidDefinition = "InvalidDefinition"

	// ErrCodeInvalidExecutionInput for service response error code
	// "InvalidExecutionInput".
	//
	// The provided JSON input data is invalid.
	ErrCodeInvalidExecutionInput = "InvalidExecutionInput"

	// ErrCodeInvalidName for service response error code
	// "InvalidName".
	//
	// The provided name is invalid.
	ErrCodeInvalidName = "InvalidName"

	// ErrCodeInvalidOutput for service response error code
	// "InvalidOutput".
	//
	// The provided JSON output data is invalid.
	ErrCodeInvalidOutput = "InvalidOutput"

	// ErrCodeInvalidToken for service response error code
	// "InvalidToken".
	//
	// The provided token is invalid.
	ErrCodeInvalidToken = "InvalidToken"

	// ErrCodeMissingRequiredParameter for service response error code
	// "MissingRequiredParameter".
	//
	// Request is missing a required parameter. This error occurs if both definition
	// and roleArn are not specified.
	ErrCodeMissingRequiredParameter = "MissingRequiredParameter"

	// ErrCodeStateMachineAlreadyExists for service response error code
	// "StateMachineAlreadyExists".
	//
	// A state machine with the same name but a different definition or role ARN
	// already exists.
	ErrCodeStateMachineAlreadyExists = "StateMachineAlreadyExists"

	// ErrCodeStateMachineDeleting for service response error code
	// "StateMachineDeleting".
	//
	// The specified state machine is being deleted.
	ErrCodeStateMachineDeleting = "StateMachineDeleting"

	// ErrCodeStateMachineDoesNotExist for service response error code
	// "StateMachineDoesNotExist".
	//
	// The specified state machine does not exist.
	ErrCodeStateMachineDoesNotExist = "StateMachineDoesNotExist"

	// ErrCodeStateMachineLimitExceeded for service response error code
	// "StateMachineLimitExceeded".
	//
	// The maximum number of state machines has been reached. Existing state machines
	// must be deleted before a new state machine can be created.
	ErrCodeStateMachineLimitExceeded = "StateMachineLimitExceeded"

	// ErrCodeTaskDoesNotExist for service response error code
	// "TaskDoesNotExist".
	ErrCodeTaskDoesNotExist = "TaskDoesNotExist"

	// ErrCodeTaskTimedOut for service response error code
	// "TaskTimedOut".
	ErrCodeTaskTimedOut = "TaskTimedOut"
)
