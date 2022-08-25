package config

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// MySQL DataSource info.
type DataSource struct {
	Database string
	User     string
}

// DO NOT keep DB info here.
// This is just for convenience of testing.
var MySqlConnection = DataSource{
	Database: "mydb",
	User:     "cow",
}

// DO NOT modify this function, change stack name by 'cdk.json/context/stackName'.
func StackName(scope constructs.Construct) string {
	stackName := "RdsMysqlCluster"

	ctxValue := scope.Node().TryGetContext(jsii.String("stackName"))
	if v, ok := ctxValue.(string); ok {
		stackName = v
	}

	return stackName
}

// DO NOT modify this function, set whether to use rds proxy by 'cdk.json/context/enableProxy'.
// The valid value are: true | false
func EnableProxy(scope constructs.Construct) bool {
	enableProxy := false

	ctxValue := scope.Node().TryGetContext(jsii.String("enableProxy"))
	if v, ok := ctxValue.(bool); ok {
		enableProxy = v
	}

	return enableProxy
}

// DO NOT modify this function, set whether to use rds replica by 'cdk.json/context/enableReplica'.
// The valid value are: true | false
func EnableReplica(scope constructs.Construct) bool {
	enableReplica := false

	ctxValue := scope.Node().TryGetContext(jsii.String("enableReplica"))
	if v, ok := ctxValue.(bool); ok {
		enableReplica = v
	}

	return enableReplica
}

// DO NOT modify this function, change event subscription by 'cdk.json/context/eventSubEmail'.
func EventSubEmail(scope constructs.Construct) string {
	eventSubEmail := "aws@amazon.com"

	ctxValue := scope.Node().TryGetContext(jsii.String("eventSubEmail"))
	if v, ok := ctxValue.(string); ok {
		eventSubEmail = v
	}

	return eventSubEmail
}
