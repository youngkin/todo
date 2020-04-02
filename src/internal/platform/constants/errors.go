//
// **NOTE** When adding constants, please order them in alphabetical sequence
//

package constants

// PostgresDupInsertErrorCode is an alias for the Postgres error code for duplicate row insert attempt
const PostgresDupInsertErrorCode = "23505" // See https://www.postgresql.org/docs/current/errcodes-appendix.html

const (
	//
	// Miscellaneous errors
	//

	// DBDeleteError is an indication of a DB error during a DELETE operation
	DBDeleteError = "a DB error occurred during a DELETE operation"
	// DBInsertDuplicateToDoError indicates an attempt to insert a duplicate row
	DBInsertDuplicateToDoError = "attempt to insert duplicate todo"
	// DBInvalidRequest indicates an invalid DB request, like attempting to update a non-existent todo
	DBInvalidRequest = "attempted update on a non-existent todo"
	// DBRowScanError indicates results from DB query could not be processed
	DBRowScanError = "DB resultset processing failed"
	// DBUpSertError indications that there was a problem executing a DB insert or update operation
	DBUpSertError = "DB insert or update failed"

	// HTTPWriteError indicates that there was a problem writing an HTTP response body
	HTTPWriteError = "Error writing HTTP response body"

	// InvalidInsertError indicates that an unexpected ToDo.ID was detected in an insert request
	InvalidInsertError = "Unexpected ToDo.ID in insert request"

	// JSONDecodingError indicates that there was a problem decoding JSON input
	JSONDecodingError = "JSON Decoding Error"
	// JSONMarshalingError indicates that there was a problem un/marshaling JSON
	JSONMarshalingError = "JSON Marshaling Error"

	// MalformedURL indicates there was a problem with the structure of the URL
	MalformedURL = "Malformed URL"

	// NoError is needed for situations where ErrCode is returned, but no error occurred
	NoError = "No error occurred"

	// RqstParsingError indicates that an error occurred while the path and/or body of the was
	// being evaluated.
	RqstParsingError = "Request parsing error"

	// UnableToCreateHTTPHandler indications that there was a problem creating an http handler
	UnableToCreateHTTPHandler = "Unable to create HTTP handler"
	// UnableToGetConfig indicates there was a problem obtaining the application configuration
	UnableToGetConfig = "Unable to get information from configuration"
	// UnableToGetDBConnStr indicates there was a problem constructing a DB connection string
	UnableToGetDBConnStr = "Unable to get DB connection string"
	// UnableToLoadConfig indicates there was a problem loading the configuration
	UnableToLoadConfig = "Unable to load configuration"
	// UnableToLoadSecrets indicates there was a problem loading the application's secrets
	UnableToLoadSecrets = "Unable to load secrets"
	// UnableToOpenConfig indicates there was a problem opening the configuration file
	UnableToOpenConfig = "Unable to open configuration file"
	// UnableToOpenDBConn indicates there was a problem opening a database connection
	UnableToOpenDBConn = "Unable to open DB connection"

	//
	// Todo related error codes start at 1000 and go to 1999
	//

	// ToDoRqstError indicates that GET(or PUT) /todo or GET(or PUT) /todo/{id} failed in some way
	ToDoRqstError = "GET /todo or GET /todo/{id} failed"
	// ToDoTypeConversionError indicates that the payload returned from GET /todo/{id} could
	// not be converted to either a ToDo (/todo) or ToDo (/todo/{id}) type
	ToDoTypeConversionError = "Unable to convert payload to ToDo(s) type"
	// ToDoValidationError indicates a problem with the ToDo data
	ToDoValidationError = "invalid todo data"
)

// ErrCode is the application type for reporting error codes
type ErrCode int

const (
	//
	// Misc related error codes start at 0 and go to 99
	//

	// DBDeleteErrorCode indication of a DB error during a DELETE operation
	DBDeleteErrorCode ErrCode = iota
	// DBInsertDuplicateToDoErrorCode indicates an attempt to insert a duplicate row
	DBInsertDuplicateToDoErrorCode
	// DBInvalidRequestCode indication of an invalid request, e.g., an update was attempted on an existing todo
	DBInvalidRequestCode
	// DBQueryErrorCode is the error code associated with DBQueryError
	DBQueryErrorCode
	// DBRowScanErrorCode is the error code associated with DBRowScan
	DBRowScanErrorCode
	// DBUpSertErrorCode indications that there was a problem executing a DB insert or update operation
	DBUpSertErrorCode

	// HTTPWriteErrorCode indicates that there was a problem writing an HTTP response body
	HTTPWriteErrorCode

	// InvalidInsertErrorCode is the error code associated with InvalidInsertError
	InvalidInsertErrorCode

	// JSONDecodingErrorCode indicates that there was a problem decoding JSON input
	JSONDecodingErrorCode
	// JSONMarshalingErrorCode is the error code associated with JSONMarshaling
	JSONMarshalingErrorCode

	// MalformedURLErrorCode is the error code associated with MalformedURL
	MalformedURLErrorCode

	// NoErrorCode is needed for situations where ErrCode is returned, but no error occurred
	NoErrorCode

	// RqstParsingErrorCode is the error code associated with RqstParsingErrorCode
	RqstParsingErrorCode

	// UnableToCreateHTTPHandlerErrorCode is the error code associated with UnableToCreateHTTPHandler
	UnableToCreateHTTPHandlerErrorCode
	// UnableToGetConfigErrorCode is the error code associated with UnableToGetConfig
	UnableToGetConfigErrorCode
	// UnableToGetDBConnStrErrorCode is the error code associated with UnableToGetDBConnStr
	UnableToGetDBConnStrErrorCode
	// UnableToLoadConfigErrorCode is the error code associated with UnableToLoadConfig
	UnableToLoadConfigErrorCode
	// UnableToLoadSecretsErrorCode is the error code associated with UnableToLoadSecrets
	UnableToLoadSecretsErrorCode
	// UnableToOpenConfigErrorCode is the error code associated with UnableToOpenConfig
	UnableToOpenConfigErrorCode
	// UnableToOpenDBConnErrorCode is the error code associated with UnableToOpenDBConn
	UnableToOpenDBConnErrorCode
)

const (
	//
	// ToDo related error codes start at 1000 and go to 1999
	//

	// ToDoRqstErrorCode is the error code associated with ToDoRqstErrorCode
	ToDoRqstErrorCode ErrCode = iota + 1000
	// ToDoTypeConversionErrorCode is the error code associated with ToDoTypeConversion
	ToDoTypeConversionErrorCode
	// ToDoValidationErrorCode indicates a problem with the ToDo data
	ToDoValidationErrorCode
)
